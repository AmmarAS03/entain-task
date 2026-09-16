// Package repository contains data storage methods
package repository

import (
	"context"
	"errors"
	"time"

	"git.neds.sh/technology/pricekinetics/tools/codetest/model"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const eventCollection = "events"

// serverSelectionTimeout bounds how long the driver waits to discover a
// server before failing. The driver's default is 30s, which turns an
// unreachable Mongo into a 30s hang on every construction site (including
// test setup) instead of a fast, clear connection error.
const serverSelectionTimeout = 5 * time.Second

type mongoRepo struct {
	client *mongo.Client
	events *mongo.Collection
}

// NewMongoRepository creates a new instance of a Repository using MongoDB as the persistance layer
func NewMongoRepository(ctx context.Context, uri string, database string) (Repository, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetServerSelectionTimeout(serverSelectionTimeout))
	if err != nil {
		logrus.Errorf("could not connect to MongoDB: %v", err)
		return nil, err
	}

	rslt := &mongoRepo{
		client: client,
		events: client.Database(database).Collection(eventCollection),
	}
	if !rslt.HealthCheck(ctx) {
		if errDisconnect := client.Disconnect(ctx); errDisconnect != nil {
			logrus.Errorf("could not disconnect from MongoDB: %v", errDisconnect)
		}
		return nil, errors.New("failed_to_init_mongo")
	}

	if err := rslt.ensureIndexes(ctx); err != nil {
		// Indexes are a performance optimisation, not a correctness requirement -
		// SearchEvents returns identical results without them, so a service that
		// can't create one is strictly worse off refusing to boot than booting slow.
		logrus.WithError(err).Warn("could not ensure indexes")
	}

	return rslt, nil
}

// ensureIndexes creates the indexes SearchEvents filters on. Index creation is
// idempotent, so this is safe on every startup. Three single-field indexes,
// rather than one compound index, because the filters are independently
// optional - a compound index only serves queries that use its prefix, while
// three single-field indexes let Mongo pick the selective one (or intersect)
// for any combination.
func (c *mongoRepo) ensureIndexes(ctx context.Context) error {
	_, err := c.events.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "starttime.value", Value: 1}}},
		{Keys: bson.D{{Key: "bettingstatus.value", Value: 1}}},
		{Keys: bson.D{{Key: "display.value", Value: 1}}},
	})
	return err
}

func (c *mongoRepo) HealthCheck(ctx context.Context) bool {
	if err := c.client.Ping(ctx, readpref.Primary()); err != nil {
		logrus.Errorf("could not connect to MongoDB: %v", err)
		return false
	}

	return true
}

func (c *mongoRepo) UpdateEvent(ctx context.Context, event *model.Event) error {
	_, err := c.events.ReplaceOne(ctx, bson.M{"_id": event.ID}, event, options.Replace().SetUpsert(true))
	if err != nil {
		logrus.Errorf("could not update event %v", err)
		return err
	}

	return nil
}

func (c *mongoRepo) GetEventByID(ctx context.Context, id string) (*model.Event, error) {
	event := &model.Event{}
	err := c.events.FindOne(ctx, bson.M{"_id": id}).Decode(event)
	if errors.Is(err, mongo.ErrNoDocuments) {
		logrus.Infof("Event not found")
		return nil, nil
	} else if err != nil {
		logrus.Errorf("could not get event %v", err)
		return nil, err
	}

	return event, nil
}

func (c *mongoRepo) DeleteEventByID(ctx context.Context, id string) error {
	_, err := c.events.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		logrus.Errorf("could not delete event %v", err)
	}

	return err
}

func (c *mongoRepo) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// SearchEvents returns every Event matching the supplied filter, sorted
// ascending by StartTime with ID as a tiebreak for deterministic ordering.
func (c *mongoRepo) SearchEvents(ctx context.Context, filter EventFilter) ([]*model.Event, error) {
	query := bson.M{}

	if filter.StartTimeFrom != nil || filter.StartTimeTo != nil {
		window := bson.M{}
		if filter.StartTimeFrom != nil {
			window["$gte"] = *filter.StartTimeFrom
		}
		if filter.StartTimeTo != nil {
			window["$lt"] = *filter.StartTimeTo
		}
		query["starttime.value"] = window
	}

	if filter.BettingStatus != nil {
		query["bettingstatus.value"] = int32(*filter.BettingStatus)
	}

	if filter.Display != nil {
		if *filter.Display {
			query["display.value"] = true
		} else {
			// Unset Display means hidden (see Task 1), and an unset Optional is
			// stored as a null "display" field, so "hidden" is every document
			// that is not explicitly true. $ne matches false, null and missing
			// alike.
			query["display.value"] = bson.M{"$ne": true}
		}
	}

	opts := options.Find().SetSort(bson.D{{Key: "starttime.value", Value: 1}, {Key: "_id", Value: 1}})
	cursor, err := c.events.Find(ctx, query, opts)
	if err != nil {
		logrus.Errorf("could not search events %v", err)
		return nil, err
	}
	defer func() {
		if errClose := cursor.Close(ctx); errClose != nil {
			logrus.Errorf("could not close search cursor %v", errClose)
		}
	}()

	events := []*model.Event{}
	if err := cursor.All(ctx, &events); err != nil {
		logrus.Errorf("could not decode searched events %v", err)
		return nil, err
	}

	return events, nil
}
