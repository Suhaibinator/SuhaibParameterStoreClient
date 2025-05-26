package client

import (
	"context"
	"fmt"
	"os"
	"time"
)

// CertificateSource holds either a file path to a certificate/key or its raw byte content.
type CertificateSource struct {
	FilePath string
	Bytes    []byte
}

// IsProvided checks if either a file path or byte content has been supplied.
func (cs *CertificateSource) IsProvided() bool {
	return cs.FilePath != "" || len(cs.Bytes) > 0
}

// GetData returns the certificate/key data.
// It prioritizes Bytes if available, otherwise reads from FilePath.
func (cs *CertificateSource) GetData() ([]byte, error) {
	if len(cs.Bytes) > 0 {
		return cs.Bytes, nil
	}
	if cs.FilePath != "" {
		data, err := os.ReadFile(cs.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read certificate/key from %s: %w", cs.FilePath, err)
		}
		return data, nil
	}
	// Nothing provided, return nil data and no error.
	// IsProvided() should be checked by the caller first if data is mandatory.
	return nil, nil
}

// Client is the main parameter store client
type Client struct {
	Host       string
	Port       int
	Timeout    time.Duration
	ClientCert CertificateSource
	ClientKey  CertificateSource
	CACert     CertificateSource
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// NewClient creates a new parameter store client with sensible defaults
func NewClient(host string, port int, opts ...ClientOption) (*Client, error) {
	if host == "" {
		return nil, fmt.Errorf("host cannot be empty")
	}
	if port <= 0 {
		return nil, fmt.Errorf("port must be a positive integer")
	}

	client := &Client{
		Host:    host,
		Port:    port,
		Timeout: 30 * time.Second, // Default timeout
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	// Validate mTLS configuration
	clientCertProvided := client.ClientCert.IsProvided()
	clientKeyProvided := client.ClientKey.IsProvided()
	caCertProvided := client.CACert.IsProvided()

	// If any part of mTLS config is provided, all parts must be provided.
	if clientCertProvided || clientKeyProvided || caCertProvided {
		if !(clientCertProvided && clientKeyProvided && caCertProvided) {
			return nil, fmt.Errorf("for mTLS, client certificate, client key, and CA certificate must all be provided")
		}
	}

	return client, nil
}

// WithTimeout sets a custom timeout for the client
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		if timeout > 0 {
			c.Timeout = timeout
		}
	}
}

// WithMTLS configures the client for mutual TLS
func WithMTLS(clientCert, clientKey, caCert CertificateSource) ClientOption {
	return func(c *Client) {
		c.ClientCert = clientCert
		c.ClientKey = clientKey
		c.CACert = caCert
	}
}

// Retrieve fetches a value for the given key using the provided secret.
// It automatically uses mTLS if certificate sources are properly configured.
func (c *Client) Retrieve(key, secret string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	serverAddress := fmt.Sprintf("%s:%v", c.Host, c.Port)

	var (
		value string
		err   error
	)

	useMTLS := c.ClientCert.IsProvided() &&
		c.ClientKey.IsProvided() &&
		c.CACert.IsProvided()

	if useMTLS {
		value, err = GrpcSimpleRetrieveWithMTLS(ctx, serverAddress, secret, key, c)
	} else {
		value, err = GrpcSimpleRetrieve(ctx, serverAddress, secret, key)
	}

	if err != nil && ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("parameter store operation timed out: %w", err)
	}

	return value, err
}

// Store stores a key-value pair in the parameter store.
// It automatically uses mTLS if certificate sources are properly configured.
func (c *Client) Store(key, secret, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	serverAddress := fmt.Sprintf("%s:%v", c.Host, c.Port)

	var err error

	useMTLS := c.ClientCert.IsProvided() &&
		c.ClientKey.IsProvided() &&
		c.CACert.IsProvided()

	if useMTLS {
		err = GrpcSimpleStoreWithMTLS(ctx, serverAddress, secret, key, value, c.ClientCert.FilePath, c.ClientKey.FilePath, c.CACert.FilePath)
	} else {
		err = GrpcSimpleStore(ctx, serverAddress, secret, key, value)
	}

	if err != nil && ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("parameter store operation timed out: %w", err)
	}

	return err
}