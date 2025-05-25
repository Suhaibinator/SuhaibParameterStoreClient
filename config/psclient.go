package config

import (
	"fmt"
	"os" // Added for os.ReadFile
	"time"
)

// RetrieveFunc is the function that will be used to retrieve values from the parameter store.
// It must be initialized by the application before using ParameterStoreClient.
var RetrieveFunc func(c *ParameterStoreClient, key, secret string) (string, error)

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

// ParameterStoreClient holds connection details for the parameter store.
type ParameterStoreClient struct {
	Host       string
	Port       int
	Timeout    time.Duration
	ClientCert CertificateSource
	ClientKey  CertificateSource
	CACert     CertificateSource
}

// NewParameterStoreClient creates a new client for the parameter store.
// For mTLS, all three CertificateSource parameters (clientCert, clientKey, caCert)
// must be provided (i.e., their IsProvided() method returns true).
func NewParameterStoreClient(
	host string,
	port int,
	timeout time.Duration,
	clientCert CertificateSource,
	clientKey CertificateSource,
	caCert CertificateSource,
) (*ParameterStoreClient, error) {
	if host == "" {
		return nil, fmt.Errorf("host cannot be empty")
	}
	if port <= 0 {
		return nil, fmt.Errorf("port must be a positive integer")
	}
	if timeout <= 0 {
		// Default timeout if not specified or invalid
		timeout = 30 * time.Second
	}

	clientCertProvided := clientCert.IsProvided()
	clientKeyProvided := clientKey.IsProvided()
	caCertProvided := caCert.IsProvided()

	// If any part of mTLS config is provided, all parts must be provided.
	if clientCertProvided || clientKeyProvided || caCertProvided {
		if !(clientCertProvided && clientKeyProvided && caCertProvided) {
			return nil, fmt.Errorf("for mTLS, client certificate, client key, and CA certificate must all be provided")
		}
	}

	return &ParameterStoreClient{
		Host:       host,
		Port:       port,
		Timeout:    timeout,
		ClientCert: clientCert,
		ClientKey:  clientKey,
		CACert:     caCert,
	}, nil
}

// retrieve fetches a value for the given key using the provided secret.
// It automatically uses mTLS if certificate sources are properly configured.
func (c *ParameterStoreClient) retrieve(key, secret string) (string, error) {
	if RetrieveFunc == nil {
		return "", fmt.Errorf("RetrieveFunc not initialized. Please call config.InitializeRetrieveFunc() after importing the client package")
	}
	return RetrieveFunc(c, key, secret)
}

// Placeholder for functions assumed to be defined elsewhere (e.g., client/grpc_client.go or config/helper.go)
// These are not part of the changes to psclient.go itself but are called by it.
// Their actual signatures and implementations would need to align with these changes.

// func grpcSimpleRetrieveFunc(ctx context.Context, serverAddress, secret, key string) (string, error)
// func grpcSimpleRetrieveWithMTLSFunc(ctx context.Context, serverAddress, secret, key string, client *ParameterStoreClient) (string, error)
// Note: The signature for grpcSimpleRetrieveWithMTLSFunc is updated to reflect passing *ParameterStoreClient.
// The original psclient.go had:
// grpcSimpleRetrieveWithMTLSFunc(ctx, serverAddress, secret, key, c.ClientCertFile, c.ClientKeyFile, c.CACertFile)
