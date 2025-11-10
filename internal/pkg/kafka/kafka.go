package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/rodney-b/swish-test-publisher/internal/pkg/config"
	"github.com/rodney-b/swish-test-publisher/pkg/certs"
)

func NewClient(ctx context.Context, cp config.ConfigProvider) (*kgo.Client, error) {
	tlsConfig, err := certs.CreateTLSConfig(cp.GetMessageQueueClientCA(), cp.GetMessageQueueClientCert(), cp.GetMessageQueueClientCertKey())
	if err != nil {
		return nil, err
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(cp.GetMessageQueueURL()),
		kgo.DialTLSConfig(tlsConfig),
	}

	return kgo.NewClient(opts...)
}
