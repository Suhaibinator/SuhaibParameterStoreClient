package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	pb "github.com/Suhaibinator/SuhaibParameterStoreClient/proto"

	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Default timeout for gRPC operations if no context deadline is set.
const defaultGrpcTimeout = 5 * time.Second

// getTransportCredentialsOptions returns the appropriate gRPC dial options for transport credentials
// based on the provided certificate paths. If all paths are non-empty, it configures mTLS.
// Otherwise, it falls back to insecure credentials.
func getTransportCredentialsOptions(caCertPath, clientCertPath, clientKeyPath string) ([]grpc.DialOption, error) {
	// If any of the certificate paths are empty, use insecure credentials
	if caCertPath == "" || clientCertPath == "" || clientKeyPath == "" {
		return []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, nil
	}

	// Load CA certificate
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate to pool")
	}

	// Load client certificate and key
	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate/key: %w", err)
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
	}

	// Return dial option with TLS credentials
	return []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))}, nil
}

// GrpcimpleRetrieve retrieves a value from the parameter store using gRPC.
// It accepts a context for timeout/cancellation control, paths for mTLS certificates, and optional grpc.DialOptions.
// If caCertPath, clientCertPath, and clientKeyPath are all provided, mTLS will be used.
// Otherwise, insecure credentials will be used.
func GrpcimpleRetrieve(ctx context.Context, ServerAddress string, AuthenticationPassword string, caCertPath string, clientCertPath string, clientKeyPath string, key string, opts ...grpc.DialOption) (val string, err error) {
	// Ensure context has a deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultGrpcTimeout)
		defer cancel()
	}

	// Get transport credentials options based on certificate paths
	transportOpts, err := getTransportCredentialsOptions(caCertPath, clientCertPath, clientKeyPath)
	if err != nil {
		log.Printf("failed to configure transport credentials: %v", err)
		return "", fmt.Errorf("failed to configure transport credentials: %w", err)
	}

	// Append any additional options passed by the caller.
	allOpts := append(transportOpts, opts...)

	// Create a client connection to the gRPC server using the combined options and context.
	// Use DialContext to respect the context deadline during connection attempt.
	conn, err := grpc.DialContext(ctx, ServerAddress, allOpts...)
	if err != nil {
		log.Printf("did not connect to %s: %v", ServerAddress, err)
		return "", fmt.Errorf("failed to connect to gRPC server at %s: %w", ServerAddress, err)
	}
	defer conn.Close()

	// Create a new client
	client := pb.NewParameterStoreClient(conn)

	// Retrieve the stored value using the provided context
	retrieveReq := &pb.RetrieveRequest{
		Key:      key,
		Password: AuthenticationPassword,
	}
	// Pass the context to the gRPC call
	retrieveResp, err := client.Retrieve(ctx, retrieveReq)
	if err != nil {
		log.Printf("could not retrieve value for key '%s': %v", key, err) // Log key for better debugging
		// Error could be context deadline exceeded if timeout occurred during the call
		return "", fmt.Errorf("gRPC retrieve call failed: %w", err)
	}
	// Check if response is nil before accessing GetValue (defensive check)
	if retrieveResp == nil {
		log.Printf("received nil response for key '%s'", key)
		return "", fmt.Errorf("received nil response from gRPC server for key '%s'", key)
	}
	return retrieveResp.GetValue(), nil // Return nil error on success
}

// GrpcSimpleStore stores a key-value pair using gRPC.
// Accepts a context for timeout/cancellation, paths for mTLS certificates, and optional grpc.DialOptions.
// If caCertPath, clientCertPath, and clientKeyPath are all provided, mTLS will be used.
// Otherwise, insecure credentials will be used.
func GrpcSimpleStore(ctx context.Context, ServerAddress string, AuthenticationPassword string, caCertPath string, clientCertPath string, clientKeyPath string, key string, value string, opts ...grpc.DialOption) (err error) {
	// Ensure context has a deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultGrpcTimeout)
		defer cancel()
	}

	// Get transport credentials options based on certificate paths
	transportOpts, err := getTransportCredentialsOptions(caCertPath, clientCertPath, clientKeyPath)
	if err != nil {
		log.Printf("failed to configure transport credentials: %v", err)
		return fmt.Errorf("failed to configure transport credentials: %w", err)
	}

	// Append any additional options passed by the caller.
	allOpts := append(transportOpts, opts...)

	// Create a client connection to the gRPC server using the combined options and context.
	conn, err := grpc.DialContext(ctx, ServerAddress, allOpts...)
	if err != nil {
		log.Printf("did not connect to %s: %v", ServerAddress, err)
		return fmt.Errorf("failed to connect to gRPC server at %s: %w", ServerAddress, err)
	}
	defer conn.Close()

	// Create a new client
	client := pb.NewParameterStoreClient(conn)

	// Store a value
	storeReq := &pb.StoreRequest{
		Key:      key,
		Value:    value,
		Password: AuthenticationPassword,
	}
	// Pass the context to the gRPC call
	storeResp, err := client.Store(ctx, storeReq)
	if err != nil {
		log.Printf("could not store value for key '%s': %v", key, err)
		return fmt.Errorf("gRPC store call failed: %w", err) // Return error properly
	}
	// Log success message instead of printing to stdout directly
	log.Printf("Store response for key '%s': %s", key, storeResp.GetMessage())
	return nil // Return nil error on success
}
