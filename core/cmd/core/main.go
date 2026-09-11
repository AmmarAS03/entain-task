// Package main is the main entry point for the core service
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.neds.sh/technology/pricekinetics/tools/codetest/core/repository"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/service"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms"
	"git.neds.sh/technology/pricekinetics/tools/codetest/core/transforms/sporttransform"
	"git.neds.sh/technology/pricekinetics/tools/codetest/merger"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	_ "google.golang.org/grpc/encoding/proto"
)

var (
	// AppName - to allow overrides at build time
	AppName = "merger"

	// Version - Should be set by CI pipeline
	Version = 1
)

// Application entry point for Feeds Core Service Host.
func main() {
	app := cli.NewApp()
	app.Name = AppName
	app.Version = fmt.Sprintf("%v", Version)
	app.Usage = "Core"
	app.Description = "Runs Transformations on the Core system model, persists and exposes the data vcia an API"
	app.Action = func(_ *cli.Context) error {
		log.SetFormatter(&log.TextFormatter{})

		repo, err := repository.NewRedisRepository(context.Background(), "localhost:6379", "")
		if err != nil {
			return err
		}

		upstreaams := &service.Upstreams{
			MergerClient: merger.NewInlineMergerClient(),
			Repo:         repo,
			Transforms: []transforms.TransformClient{
				sporttransform.NewSportTransformClient(),
			},
		}

		// Run the service as a goroutine, watching for errors
		service := service.NewService(50051, 8080, upstreaams)
		errors := make(chan error, 1)
		go func() {
			err := service.Run()
			if err != nil {
				errors <- err
			}
		}()

		// Wait for the signal to die
		signals := make(chan os.Signal, 1)
		signal.Notify(signals,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		select {
		case err := <-errors:
			log.WithError(err).Error("service_error")
		case sig := <-signals:
			log.WithField("signal", sig).Warn("shutdown_signal")
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(10))
		defer cancel()
		errStop := service.Stop(shutdownCtx)
		if errStop != nil {
			log.WithError(errStop).Warn("shutdown_error")
		}
		log.Info("shutdown_complete")
		return nil
	}

	errRun := app.Run(os.Args)
	if errRun != nil {
		os.Exit(-1)
	}
}
