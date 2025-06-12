package client

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
	"software.sslmate.com/src/go-pkcs12"
)

// CertificateSource holds either a file path to a certificate/key or its raw byte content.
type CertificateSource struct {
	FilePath string
	Bytes    []byte
}

// PasswordCallback is a function type for securely obtaining passwords.
// It should return the password and any error that occurred during retrieval.
type PasswordCallback func() (string, error)

// TLSConfig holds TLS configuration for the client, supporting multiple certificate formats.
type TLSConfig struct {
	// Separate certificate and key (PEM format)
	ClientCert CertificateSource
	ClientKey  CertificateSource
	CACert     CertificateSource

	// PKCS#12 (.p12/.pfx) support
	P12File       string
	P12Bytes      []byte
	P12PasswordFn PasswordCallback

	// ServerName for TLS verification (optional)
	ServerName string

	// InsecureSkipVerify disables certificate verification (for testing only)
	InsecureSkipVerify bool
}

// IsConfigured returns true if any TLS configuration is provided.
func (tc *TLSConfig) IsConfigured() bool {
	return tc.hasP12Config() || tc.hasSeparateCertConfig()
}

// hasP12Config returns true if P12 configuration is provided.
func (tc *TLSConfig) hasP12Config() bool {
	return tc.P12File != "" || len(tc.P12Bytes) > 0
}

// hasSeparateCertConfig returns true if separate cert/key configuration is provided.
func (tc *TLSConfig) hasSeparateCertConfig() bool {
	return tc.ClientCert.IsProvided() && tc.ClientKey.IsProvided() && tc.CACert.IsProvided()
}

// GetTLSConfig builds and returns a *tls.Config based on the configuration.
func (tc *TLSConfig) GetTLSConfig() (*tls.Config, error) {
	if !tc.IsConfigured() {
		return nil, fmt.Errorf("no TLS configuration provided")
	}

	tlsConfig := &tls.Config{
		ServerName:         tc.ServerName,
		InsecureSkipVerify: tc.InsecureSkipVerify,
	}

	if tc.hasP12Config() {
		return tc.configureTLSFromP12(tlsConfig)
	}

	return tc.configureTLSFromSeparateCerts(tlsConfig)
}

// configureTLSFromP12 configures TLS using PKCS#12 data.
func (tc *TLSConfig) configureTLSFromP12(tlsConfig *tls.Config) (*tls.Config, error) {
	// Get P12 data
	p12Data, err := tc.getP12Data()
	if err != nil {
		return nil, fmt.Errorf("failed to get P12 data: %w", err)
	}

	// Get password
	password, err := tc.getP12Password()
	if err != nil {
		return nil, fmt.Errorf("failed to get P12 password: %w", err)
	}
	defer func() {
		// Clear password from memory
		passwordBytes := []byte(password)
		for i := range passwordBytes {
			passwordBytes[i] = 0
		}
	}()

	// Parse P12
	privateKey, certificate, caCerts, err := pkcs12.DecodeChain(p12Data, password)
	if err != nil {
		// Check if it's an unsupported digest algorithm error
		errStr := err.Error()
		if strings.Contains(errStr, "unknown digest algorithm") || strings.Contains(errStr, "2.16.840.1.101.3.4.2.1") {
			return nil, fmt.Errorf("failed to decode P12: unsupported digest algorithm (SHA-256). The Go PKCS#12 library doesn't support modern digest algorithms.\n\nSolutions:\n1. Convert P12 to use SHA-1: openssl pkcs12 -in cert.p12 -out temp.pem -nodes && openssl pkcs12 -export -in temp.pem -out cert_sha1.p12 -macalg SHA1\n2. Use separate PEM files instead\n\nOriginal error: %w", err)
		}
		return nil, fmt.Errorf("failed to decode P12: %w", err)
	}

	// Convert to PEM format for tls.X509KeyPair
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificate.Raw,
	})
	
	// Handle different private key types
	var keyPEM []byte
	var keyType string
	var keyBytes []byte
	
	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		keyType = "RSA PRIVATE KEY"
		keyBytes = x509.MarshalPKCS1PrivateKey(key)
	case *ecdsa.PrivateKey:
		keyType = "EC PRIVATE KEY"
		var err error
		keyBytes, err = x509.MarshalECPrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal EC private key: %w", err)
		}
	default:
		keyType = "PRIVATE KEY"
		var err error
		keyBytes, err = x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal private key: %w", err)
		}
	}
	
	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  keyType,
		Bytes: keyBytes,
	})

	if certPEM == nil || keyPEM == nil {
		return nil, fmt.Errorf("P12 file missing required certificate or private key")
	}
	
	// Convert CA certificates to PEM format
	var caCertsPEM [][]byte
	for _, caCert := range caCerts {
		caCertPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: caCert.Raw,
		})
		caCertsPEM = append(caCertsPEM, caCertPEM)
	}

	// Create client certificate
	clientCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to create X509 key pair: %w", err)
	}
	tlsConfig.Certificates = []tls.Certificate{clientCert}

	// Configure CA certificates if provided
	if len(caCertsPEM) > 0 {
		caCertPool := x509.NewCertPool()
		for _, caCertPEM := range caCertsPEM {
			if !caCertPool.AppendCertsFromPEM(caCertPEM) {
				return nil, fmt.Errorf("failed to parse CA certificate from P12")
			}
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}

// configureTLSFromSeparateCerts configures TLS using separate cert/key files.
func (tc *TLSConfig) configureTLSFromSeparateCerts(tlsConfig *tls.Config) (*tls.Config, error) {
	// Load client certificate and key
	clientCertData, err := tc.ClientCert.GetData()
	if err != nil {
		return nil, fmt.Errorf("failed to get client cert data: %w", err)
	}
	clientKeyData, err := tc.ClientKey.GetData()
	if err != nil {
		return nil, fmt.Errorf("failed to get client key data: %w", err)
	}

	clientCert, err := tls.X509KeyPair(clientCertData, clientKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert/key pair: %w", err)
	}
	tlsConfig.Certificates = []tls.Certificate{clientCert}

	// Load CA certificate
	caCertData, err := tc.CACert.GetData()
	if err != nil {
		return nil, fmt.Errorf("failed to get CA cert data: %w", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCertData) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}
	tlsConfig.RootCAs = caCertPool

	return tlsConfig, nil
}

// getP12Data returns the P12 data from either file or bytes.
func (tc *TLSConfig) getP12Data() ([]byte, error) {
	if len(tc.P12Bytes) > 0 {
		return tc.P12Bytes, nil
	}
	if tc.P12File != "" {
		return os.ReadFile(tc.P12File)
	}
	return nil, fmt.Errorf("no P12 data provided")
}

// getP12Password gets the P12 password using the configured callback.
func (tc *TLSConfig) getP12Password() (string, error) {
	if tc.P12PasswordFn != nil {
		return tc.P12PasswordFn()
	}
	return "", fmt.Errorf("no P12 password callback provided")
}

// DefaultPasswordCallbacks provides common password input methods.
var DefaultPasswordCallbacks = struct {
	// TerminalPrompt prompts the user for a password via terminal input.
	TerminalPrompt PasswordCallback
	// EnvVar returns a callback that reads the password from an environment variable.
	EnvVar func(envVarName string) PasswordCallback
	// Static returns a callback that returns a static password (not recommended for production).
	Static func(password string) PasswordCallback
}{
	TerminalPrompt: func() (string, error) {
		fmt.Print("Enter P12 password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // Print newline after password input
		if err != nil {
			return "", fmt.Errorf("failed to read password: %w", err)
		}
		return string(passwordBytes), nil
	},
	EnvVar: func(envVarName string) PasswordCallback {
		return func() (string, error) {
			password := os.Getenv(envVarName)
			if password == "" {
				return "", fmt.Errorf("environment variable %s not set or empty", envVarName)
			}
			return password, nil
		}
	},
	Static: func(password string) PasswordCallback {
		return func() (string, error) {
			return password, nil
		}
	},
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
	Host      string
	Port      int
	Timeout   time.Duration
	TLSConfig *TLSConfig
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

	// Validate TLS configuration
	if client.TLSConfig != nil {
		if err := client.validateTLSConfig(); err != nil {
			return nil, fmt.Errorf("invalid TLS configuration: %w", err)
		}
	}

	return client, nil
}

// validateTLSConfig validates the TLS configuration.
func (c *Client) validateTLSConfig() error {
	if c.TLSConfig == nil {
		return nil
	}

	// Check for conflicting configurations
	hasP12 := c.TLSConfig.hasP12Config()
	hasSeparate := c.TLSConfig.hasSeparateCertConfig()

	if hasP12 && hasSeparate {
		return fmt.Errorf("cannot specify both P12 and separate cert/key configurations")
	}

	if !hasP12 && !hasSeparate {
		return fmt.Errorf("TLS configuration provided but no certificates specified")
	}

	// Validate P12 configuration
	if hasP12 {
		if c.TLSConfig.P12File == "" && len(c.TLSConfig.P12Bytes) == 0 {
			return fmt.Errorf("P12 configuration requires either P12File or P12Bytes")
		}
		if c.TLSConfig.P12PasswordFn == nil {
			return fmt.Errorf("P12 configuration requires a password callback function")
		}
	}

	// Validate separate cert configuration
	if hasSeparate {
		if !c.TLSConfig.ClientCert.IsProvided() {
			return fmt.Errorf("client certificate is required for separate cert configuration")
		}
		if !c.TLSConfig.ClientKey.IsProvided() {
			return fmt.Errorf("client key is required for separate cert configuration")
		}
		if !c.TLSConfig.CACert.IsProvided() {
			return fmt.Errorf("CA certificate is required for separate cert configuration")
		}
	}

	return nil
}

// WithTimeout sets a custom timeout for the client
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		if timeout > 0 {
			c.Timeout = timeout
		}
	}
}

// WithTLS configures the client with TLS settings.
func WithTLS(tlsConfig *TLSConfig) ClientOption {
	return func(c *Client) {
		c.TLSConfig = tlsConfig
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

	useTLS := c.TLSConfig != nil && c.TLSConfig.IsConfigured()

	if useTLS {
		value, err = GrpcSimpleRetrieveWithTLS(ctx, serverAddress, secret, key, c.TLSConfig)
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

	useTLS := c.TLSConfig != nil && c.TLSConfig.IsConfigured()

	if useTLS {
		err = GrpcSimpleStoreWithTLS(ctx, serverAddress, secret, key, value, c.TLSConfig)
	} else {
		err = GrpcSimpleStore(ctx, serverAddress, secret, key, value)
	}

	if err != nil && ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("parameter store operation timed out: %w", err)
	}

	return err
}

// TLS Configuration Factory Functions

// NewTLSConfigFromSeparateCerts creates a TLS config using separate cert/key files.
func NewTLSConfigFromSeparateCerts(clientCertFile, clientKeyFile, caCertFile string) *TLSConfig {
	return &TLSConfig{
		ClientCert: CertificateSource{FilePath: clientCertFile},
		ClientKey:  CertificateSource{FilePath: clientKeyFile},
		CACert:     CertificateSource{FilePath: caCertFile},
	}
}

// NewTLSConfigFromSeparateCertBytes creates a TLS config using separate cert/key data.
func NewTLSConfigFromSeparateCertBytes(clientCertData, clientKeyData, caCertData []byte) *TLSConfig {
	return &TLSConfig{
		ClientCert: CertificateSource{Bytes: clientCertData},
		ClientKey:  CertificateSource{Bytes: clientKeyData},
		CACert:     CertificateSource{Bytes: caCertData},
	}
}

// NewTLSConfigFromP12File creates a TLS config using a PKCS#12 file.
func NewTLSConfigFromP12File(p12File string, passwordFn PasswordCallback) *TLSConfig {
	return &TLSConfig{
		P12File:       p12File,
		P12PasswordFn: passwordFn,
	}
}

// NewTLSConfigFromP12Bytes creates a TLS config using PKCS#12 data.
func NewTLSConfigFromP12Bytes(p12Data []byte, passwordFn PasswordCallback) *TLSConfig {
	return &TLSConfig{
		P12Bytes:      p12Data,
		P12PasswordFn: passwordFn,
	}
}
