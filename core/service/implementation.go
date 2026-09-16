// Package service contains the grpc/http server implementation
package service

import (
	"context"
	"fmt"

	"git.neds.sh/technology/pricekinetics/tools/codetest/core"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/repository"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms"
	"git.neds.sh/technology/pricekinetics/tools/codetest/merger"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Upstreams defines dependencies the service has on other services
type Upstreams struct {
	MergerClient merger.ServiceClient
	Repo         repository.Repository
	Transforms   []transforms.TransformClient
}

// NewService creqtes a new instancxe of Service
func NewService(grpcPort, httpport int, upstreams *Upstreams) *Service {
	return &Service{GRPCPort: grpcPort, HTTPPort: httpport, Upstreams: upstreams}
}

// RegisterGRPCServerImplementations registers the grpc service contract implemented by this server
func (host *Service) RegisterGRPCServerImplementations(grpcServer *grpc.Server) {
	core.RegisterServiceServer(grpcServer, host)
}

// Update updates an Event and runs the pipeline of transformations
func (host *Service) Update(ctx context.Context, req *core.UpdateRequest) (*core.UpdateResponse, error) {
	existing, err := host.Upstreams.Repo.GetEventByID(ctx, req.GetEvent().GetID())
	if err != nil {
		logrus.WithError(err).Error("Update: failed to retrieve event")
		return nil, err
	}

	resp := &core.UpdateResponse{Message: "Success"}

	update := req.GetEvent()
	if existing == nil {
		resp.Message = fmt.Sprintf("New Event born %v", req.GetEvent().GetID())
	} else {
		update, err = host.Upstreams.MergerClient.MergeEvent(context.Background(), existing, req.GetEvent())
		if err != nil {
			logrus.WithError(err).Error("Update: failed to merge event")
			return nil, err
		}
	}

	for _, t := range host.Upstreams.Transforms {
		upd, tErr := t.TransformEvent(ctx, req.Event, update)
		if tErr != nil {
			logrus.WithError(tErr).Errorf("Update: failed to run transform %v", t.GetName())
		}
		if upd != nil {
			update, err = host.Upstreams.MergerClient.MergeEvent(context.Background(), update, upd)
			if err != nil {
				logrus.WithError(err).Errorf("Update: failed to merge event in transform %v", t.GetName())
				return nil, err
			}
		}
	}

	err = host.Upstreams.Repo.UpdateEvent(ctx, update)
	if err != nil {
		logrus.WithError(err).Error("Update: failed to update event")
		return nil, err
	}

	return resp, nil
}

// GetSportEvent retrieves a model.Event from the database and returns a core.SportEvent - this is a more UserConsumable representation of the model that is specific to sport events
func (host *Service) GetSportEvent(ctx context.Context, req *core.GetSportEventRequest) (*core.GetSportEventResponse, error) {
	existing, err := host.Upstreams.Repo.GetEventByID(ctx, req.GetEventID())
	if err != nil {
		logrus.WithError(err).Error("GetSportEvent: failed to retrieve event")
		return nil, err
	}

	resp := &core.GetSportEventResponse{}

	if existing == nil {
		return resp, nil
	}

	rslt := &core.SportEvent{}
	rslt.ConvertFromModel(existing)
	resp.Event = rslt

	return resp, nil
}

// GetRacingEvent retrieves a model.Event from the database and returns a core.RacingEvent - a race card style view for consumers of racing events
func (host *Service) GetRacingEvent(ctx context.Context, req *core.GetRacingEventRequest) (*core.GetRacingEventResponse, error) {
	existing, err := host.Upstreams.Repo.GetEventByID(ctx, req.GetEventID())
	if err != nil {
		logrus.WithError(err).Error("GetRacingEvent: failed to retrieve event")
		return nil, err
	}

	resp := &core.GetRacingEventResponse{}

	if existing == nil {
		return resp, nil
	}

	rslt := &core.RacingEvent{}
	rslt.ConvertFromModel(existing)
	resp.Event = rslt

	return resp, nil
}

// SearchEvents returns a summary of every Event matching the supplied criteria.
func (host *Service) SearchEvents(ctx context.Context, req *core.SearchEventsRequest) (*core.SearchEventsResponse, error) {
	filter := repository.EventFilter{}
	if v := req.GetStartTimeFrom(); v != nil {
		from := v.GetValue()
		filter.StartTimeFrom = &from
	}
	if v := req.GetStartTimeTo(); v != nil {
		to := v.GetValue()
		filter.StartTimeTo = &to
	}
	if v := req.GetBettingStatus(); v != nil {
		status := v.GetValue()
		filter.BettingStatus = &status
	}
	if v := req.GetDisplay(); v != nil {
		display := v.GetValue()
		filter.Display = &display
	}

	events, err := host.Upstreams.Repo.SearchEvents(ctx, filter)
	if err != nil {
		logrus.WithError(err).Error("SearchEvents: failed to search events")
		return nil, err
	}

	resp := &core.SearchEventsResponse{Events: make([]*core.EventSummary, 0, len(events))}
	for _, event := range events {
		summary := &core.EventSummary{}
		summary.ConvertFromModel(event)
		resp.Events = append(resp.Events, summary)
	}

	return resp, nil
}
