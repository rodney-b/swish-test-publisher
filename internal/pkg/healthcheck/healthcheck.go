package healthcheck

import (
	"context"
	"errors"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/rodney-b/swish-test-publisher/internal/pkg/config"
	"github.com/rodney-b/swish-test-publisher/internal/pkg/logger"
)

var (
	healthServer     *health.Server
	grpcServer       *grpc.Server
	serverOnce       sync.Once
	livenessSvcName  string
	readinessSvcName string
)

const (
	livenessSuffix  = "-liveness"
	readinessSuffix = "-readiness"
)

func Start(cp config.ConfigProvider) error {
	log := logger.New("healthcheck")
	var err error

	serverOnce.Do(func() {
		healthcheckAddress := ":" + cp.GetHealthcheckPort()
		var listener net.Listener
		listener, err = net.Listen("tcp", healthcheckAddress)
		if err != nil {
			err = errors.Join(err, errors.New("failed to set up healthcheck server"))
			return
		}

		servicePrefix := cp.GetHealthcheckServicePrefix()
		livenessSvcName = servicePrefix + livenessSuffix
		readinessSvcName = servicePrefix + readinessSuffix
		grpcServer = grpc.NewServer([]grpc.ServerOption{}...)
		healthServer = health.NewServer()
		healthgrpc.RegisterHealthServer(grpcServer, healthServer)

		log.Info("Starting healthcheck server",
			"app", cp.GetAppName(),
			"port", cp.GetHealthcheckPort())

		go func() {
			err = grpcServer.Serve(listener)
			if err != nil {
				log.Error("healthcheck server error", "error", err.Error())
				return
			}
			log.Info("Stopping healthcheck server")
		}()
	})
	if err != nil {
		return err
	}

	SetAppLivenessStatus(healthgrpc.HealthCheckResponse_SERVING)

	log.Info("health server started successfully")
	return nil
}

func SetAppLivenessStatus(status healthgrpc.HealthCheckResponse_ServingStatus) {
	healthServer.SetServingStatus(livenessSvcName, status)
}

func SetAppReadinessStatus(status healthgrpc.HealthCheckResponse_ServingStatus) {
	healthServer.SetServingStatus(readinessSvcName, status)
}

// SetServiceStatus sets the health status for service
func SetServiceStatus(service string, status healthgrpc.HealthCheckResponse_ServingStatus) {
	healthServer.SetServingStatus(service, status)
}

// GetServiceStatus gets the health status for service
func GetServiceStatus(ctx context.Context, service string) (*healthgrpc.HealthCheckResponse, error) {
	req := healthgrpc.HealthCheckRequest{
		Service: service,
	}
	return healthServer.Check(ctx, &req)
}

// GetAppLivenessStatus returns the liveness status for the application
func GetAppLivenessStatus(ctx context.Context) (*healthgrpc.HealthCheckResponse, error) {
	return GetServiceStatus(ctx, livenessSvcName)
}

// GetAppReadinessStatus returns the readiness status for the application
func GetAppReadinessStatus(ctx context.Context) (*healthgrpc.HealthCheckResponse, error) {
	return GetServiceStatus(ctx, readinessSvcName)
}
