package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gorilla/mux"
	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Service describes the main service of the application housing a grpc and http server
type Service struct {
	Upstreams         *Upstreams
	grpcServer        *grpc.Server
	httpServer        *http.Server
	GRPCPort          int
	HTTPPort          int
	shutdownCallbacks []func() // Shutdown cleanup callbacks
}

// Run executes the current service in a blocking fashion.
func (host *Service) Run() error {
	// defer tracing.EnsureGlobalTracer(wrapper.serviceName)()

	// run the http server in non-blocking mode (we'll run the grpc server in blocking mode)
	errHealth := host.runHTTPServer()
	if errHealth != nil {
		return fmt.Errorf("service_health_start_failed")
	}

	log.Info("constructing_grpc_server")

	opts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) error {
			log.WithField("panic", p).WithField("Stack", string(debug.Stack())).Error("request_panic_caught")
			return fmt.Errorf("server_panic")
		}),
	}

	serverOptions := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024 * 1024 * 64),
		grpc.MaxSendMsgSize(1024 * 1024 * 64),
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(opts...),
		),
		grpc.ChainStreamInterceptor(
			recovery.StreamServerInterceptor(opts...),
		),
	}

	host.grpcServer = grpc.NewServer(serverOptions...)

	// the service needs to register it's grpc implementations
	host.RegisterGRPCServerImplementations(host.grpcServer)

	reflection.Register(host.grpcServer)

	serviceAddress := fmt.Sprintf(":%d", host.GRPCPort)
	lis, err := net.Listen("tcp", serviceAddress)
	if err != nil {
		return fmt.Errorf("service_start_failed: %v", err)
	}
	host.shutdownCallbacks = append(host.shutdownCallbacks, func() {
		errClose := lis.Close()
		if errClose != nil {
			log.WithError(errClose).Warn("error_closing_grpc_listener")
		}
	})

	log.WithField("address", serviceAddress).Info("service_start")
	defer log.WithField("address", serviceAddress).Info("service_stopping")

	return host.grpcServer.Serve(lis)
}

// Stop the wrapper service servers
func (host *Service) Stop(_ context.Context) error {
	log.Info("service_stop_requested")
	defer log.Info("service_stop_completed")
	if host.grpcServer != nil {
		log.Info("service_grpc_stopping")
		host.grpcServer.Stop()
	} else {
		log.Warn("grpc_service_shutdown_skipped")
	}

	log.Info("service_stop_callbacks")
	// Stop all shutdown callbacks
	for i, callback := range host.shutdownCallbacks {
		log := log.WithField("callback", i)
		log.Debug("service_stop_callback_start")
		callback()
		log.Debug("service_stop_callback_complete")
	}

	return nil
}

// runHealthEndpoint runs the health listener
func (host *Service) runHTTPServer() error {
	router := mux.NewRouter()

	// Run our HTTP handler
	address := fmt.Sprintf(":%v", host.HTTPPort)
	srv := &http.Server{
		Addr:              address,
		Handler:           router,
		ReadHeaderTimeout: time.Second * 60,
	}

	host.httpServer = srv
	routine := sync.WaitGroup{}
	routine.Add(1)

	mtx := sync.Mutex{}
	var errService error

	go func() {
		log.WithFields(log.Fields{
			"address": address,
		}).Info("http_listener_starting")
		routine.Done()
		err := srv.ListenAndServe()
		if err != nil {
			mtx.Lock()
			errService = err
			mtx.Unlock()
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("http_listener_failed")
		}
		log.Info("http_listener_stopped")
	}()

	routine.Wait()
	time.Sleep(time.Second)

	errHealth := SucceedWithin(time.Minute, func() error {
		mtx.Lock()
		if errService != nil {
			mtx.Unlock()
			return nil // Stop the retry loop
		}
		mtx.Unlock()
		return nil
	})

	mtx.Lock()
	defer mtx.Unlock()
	if errService != nil {
		return errService
	}

	return errHealth
}

// Retryable is a function that can be called over and over
type Retryable func() error

// SucceedWithin checks that a retryable operation succeeds within the
// given timeout
func SucceedWithin(d time.Duration, cb Retryable) error {
	ticker := time.After(d)
	delay := time.Millisecond
	for {
		select {
		case <-ticker:
			return fmt.Errorf("timed_out_retry")
		default:
			// Succeed this time?
			errRetry := cb()
			if errRetry == nil {
				return nil
			}
			if errRetry.Error() == "aborted_due_to_failure" {
				return fmt.Errorf("aborted_due_to_failure")
			}
		}
		time.Sleep(delay)
		delay *= 2
	}
}
