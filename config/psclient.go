package config

import (
	"context"
	"fmt"
	"time"
)

// ParameterStoreClient holds connection details for the parameter store.
type ParameterStoreClient struct {
	Host           string
	Port           int
	Timeout        time.Duration
	ClientCertFile string
	ClientKeyFile  string
	CACertFile     string
}

// retrieve fetches a value for the given key using the provided secret.
// It automatically uses mTLS if certificate paths are present.
func (c *ParameterStoreClient) retrieve(key, secret string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	serverAddress := fmt.Sprintf("%s:%v", c.Host, c.Port)

	var (
		value string
		err   error
	)

	if c.ClientCertFile != "" && c.ClientKeyFile != "" && c.CACertFile != "" {
		value, err = grpcSimpleRetrieveWithMTLSFunc(ctx, serverAddress, secret, key, c.ClientCertFile, c.ClientKeyFile, c.CACertFile)
	} else {
		value, err = grpcSimpleRetrieveFunc(ctx, serverAddress, secret, key)
	}

	if err != nil && ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("parameter store operation timed out: %w", err)
	}

	return value, err
}
