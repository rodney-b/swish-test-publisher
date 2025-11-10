package certs

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
)

var (
	ErrCaCertAppendFail       = fmt.Errorf("failed to append server CA certificate")
	ErrParseClientKeyPairFail = fmt.Errorf("failed to parse client:key pair")
)

func CreateTLSConfig(caCertPEM, clientCertPEM, clientKeyPEM []byte) (*tls.Config, error) {
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertPEM); !ok {
		return nil, ErrCaCertAppendFail
	}

	clientCert, err := tls.X509KeyPair(clientCertPEM, clientKeyPEM)
	if err != nil {
		return nil, errors.Join(ErrParseClientKeyPairFail, err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS12, // Optional: enforce minimum TLS version
	}

	return tlsConfig, nil
}
