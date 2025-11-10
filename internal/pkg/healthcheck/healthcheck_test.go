package healthcheck_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/rodney-b/swish-test-publisher/internal/pkg/config"
	"github.com/rodney-b/swish-test-publisher/internal/pkg/healthcheck"
	"github.com/rodney-b/swish-test-publisher/internal/pkg/logger"
)

const (
	msgUnnexpectedHealthStatus = "unexpected health status: expected %s but found %s"
)

type closeClientFunc func() error

func initHealthClient(cp config.ConfigProvider) (healthgrpc.HealthClient, closeClientFunc, error) {
	port := cp.GetHealthcheckPort()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient("localhost:"+port, opts...)
	if err != nil {
		return nil, nil, err
	}

	hClient := healthgrpc.NewHealthClient(conn)

	return hClient, conn.Close, nil
}

func failUnnexpectedStatus(t *testing.T, expected, actual healthgrpc.HealthCheckResponse_ServingStatus) {
	if actual != expected {
		t.Fatalf(msgUnnexpectedHealthStatus, expected, actual)
	}
}

func testClientStatusCheck(t *testing.T) {
	t.Parallel()

	appConfig, err := config.InitAppConfig()
	if err != nil {
		t.Fatalf("failed to init env provider: %v", err)
	}

	hClient, closeClientFunc, err := initHealthClient(appConfig)
	if err != nil {
		t.Fatalf("failed to init healthcheck client: %v", err)
	}
	defer closeClientFunc()

	newService := "new-client-service"
	healthcheck.SetServiceStatus(newService, healthgrpc.HealthCheckResponse_SERVING)

	req := healthgrpc.HealthCheckRequest{
		Service: newService,
	}

	resp, err := hClient.Check(context.Background(), &req)
	if err != nil {
		t.Fatalf("failed to check the health of service  %s: %v", newService, err)
	}
	failUnnexpectedStatus(t, healthgrpc.HealthCheckResponse_SERVING, resp.GetStatus())
}

// testInternalAppStatus tests internal wrapper funcs for setting and getting the
// health status of this app and its services
func testInternalAppStatus(t *testing.T) {
	t.Parallel()

	type args struct {
		status healthgrpc.HealthCheckResponse_ServingStatus
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		setter     func(healthgrpc.HealthCheckResponse_ServingStatus)
		getter     func(ctx context.Context) (*healthgrpc.HealthCheckResponse, error)
		wantStatus healthgrpc.HealthCheckResponse_ServingStatus
	}{
		{
			name: "get initial liveness status successfully",
			args: args{
				status: healthgrpc.HealthCheckResponse_SERVING,
			},
			setter:     nil,
			getter:     healthcheck.GetAppLivenessStatus,
			wantStatus: healthgrpc.HealthCheckResponse_SERVING,
		},
		{
			name: "set & get liveness status successfully",
			args: args{
				status: healthgrpc.HealthCheckResponse_NOT_SERVING,
			},
			setter:     healthcheck.SetAppLivenessStatus,
			getter:     healthcheck.GetAppLivenessStatus,
			wantStatus: healthgrpc.HealthCheckResponse_NOT_SERVING,
		},
		{
			name: "set & get readiness status successfully",
			args: args{
				status: healthgrpc.HealthCheckResponse_NOT_SERVING,
			},
			setter:     healthcheck.SetAppReadinessStatus,
			getter:     healthcheck.GetAppReadinessStatus,
			wantStatus: healthgrpc.HealthCheckResponse_NOT_SERVING,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setter != nil {
				tt.setter(tt.args.status)
			}

			resp, err := tt.getter(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("app status check error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			failUnnexpectedStatus(t, tt.wantStatus, resp.GetStatus())
		})
	}

	// service status
	newService := "new-internal-service"
	healthcheck.SetServiceStatus(newService, healthgrpc.HealthCheckResponse_SERVING)

	resp, err := healthcheck.GetServiceStatus(context.Background(), newService)
	if err != nil {
		t.Fatalf("error geting the health status of service %s: %v", newService, err)
	}

	failUnnexpectedStatus(t, healthgrpc.HealthCheckResponse_SERVING, resp.GetStatus())
}

func TestHealthcheck(t *testing.T) {
	appConfig, err := config.InitAppConfig()
	if err != nil {
		t.Fatalf("failed to init env provider: %v", err)
	}

	logger.Initialize(appConfig)

	err = healthcheck.Start(appConfig)
	if err != nil {
		t.Fatalf("failed to start healthcheck: %v", err)
	}

	// giving the server a couple of seconds to start
	time.Sleep(2 * time.Second)

	// testing internal healthcheck funcs
	tests := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{
			name:     "Test Internal App Status",
			testFunc: testInternalAppStatus,
		},
		{
			name:     "Test Client Status Check",
			testFunc: testClientStatusCheck,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.testFunc)
	}
}
