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

	return rslt, nil
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
