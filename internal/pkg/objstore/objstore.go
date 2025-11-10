package objstore

import (
	"context"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/rodney-b/swish-test-publisher/internal/pkg/config"
	"github.com/rodney-b/swish-test-publisher/pkg/certs"
)

func NewMinioClient(ctx context.Context, cp config.ConfigProvider) (*minio.Client, error) {
	tlsConfig, err := certs.CreateTLSConfig(cp.GetDataSourceClientCA(), cp.GetDataSourceClientCert(), cp.GetDataSourceClientCertKey())
	if err != nil {
		return nil, err
	}
	transport := http.Transport{
		TLSClientConfig: tlsConfig,
	}
	miniOpts := minio.Options{
		Creds:     credentials.NewStaticV4(cp.GetConsoleAccessKey(), cp.GetConsoleSecretKey(), ""),
		Secure:    true,
		Transport: &transport,
	}
	client, err := minio.New(cp.GetDataSourceURL(), &miniOpts)
	if err != nil {
		return nil, err
	}

	return client, nil
}
