package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/Suhaibinator/SuhaibParameterStoreClient/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCDialContextFunc is a function variable that can be replaced for testing.
// It defaults to the standard grpc.NewClient function.
// grpc.NewClient creates a client connection to a gRPC server. It does not directly
// accept a context.Context argument for the dial operation itself.
// The context (e.g., `ctx`) passed to wrapper functions like GrpcSimpleRetrieve is
// still used for:
// 1. Pre-dial check (ctx.Err()).
// 2. Timeout/cancellation of the dial operation if grpc.WithBlock() is used as a DialOption.
// 3. Timeout/cancellation of subsequent RPC calls (e.g., client.Retrieve(ctx, ...)).
var GRPCDialContextFunc = grpc.NewClient

// Default timeout for gRPC operations if no context deadline is set.
const defaultGrpcTimeout = 5 * time.Second

// GrpcSimpleRetrieve retrieves a value from the parameter store using gRPC.
// It accepts a context for timeout/cancellation control and optional grpc.DialOptions.
func GrpcSimpleRetrieve(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) { //nolint:all // Existing function
	// Ensure context has a deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultGrpcTimeout)
		defer cancel()
	}

	// Default options: insecure credentials.
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// grpc.WithBlock(), // Let caller decide if they need blocking dial via opts
	}
	// Append any additional options passed by the caller.
	allOpts := append(defaultOpts, opts...)

	// Check if the context is already done before attempting to dial.
	if err := ctx.Err(); err != nil {
		log.Printf("gRPC dial context for %s already expired/canceled: %v", ServerAddress, err)
		return "", fmt.Errorf("gRPC dial context for %s already expired/canceled: %w", ServerAddress, err)
	}

	// Create a client connection to the gRPC server using the combined options,
	// via GRPCDialContextFunc (which defaults to grpc.NewClient).
	// See the comment on GRPCDialContextFunc for how context.Context is handled.
	target := "passthrough:///" + ServerAddress
	conn, err := GRPCDialContextFunc(target, allOpts...)
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

// GrpcimpleRetrieve is deprecated. Use GrpcSimpleRetrieve instead.
func GrpcimpleRetrieve(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (string, error) {
	return GrpcSimpleRetrieve(ctx, ServerAddress, AuthenticationPassword, key, opts...)
}

// GrpcSimpleRetrieveWithMTLS retrieves a value from the parameter store using gRPC with mTLS.
// It accepts a context, server details, a key, a Client config, and optional grpc.DialOptions.
func GrpcSimpleRetrieveWithMTLS(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientConfig *Client, opts ...grpc.DialOption) (val string, err error) {
	// Ensure context has a deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultGrpcTimeout)
		defer cancel()
	}

	// Get certificate data using GetData methods from ParameterStoreClient's CertificateSource fields
	clientCertBytes, err := clientConfig.ClientCert.GetData()
	if err != nil {
		log.Printf("Failed to get client certificate data: %v", err)
		return "", fmt.Errorf("failed to get client certificate data: %w", err)
	}
	if len(clientCertBytes) == 0 {
		log.Printf("Client certificate data is empty")
		return "", fmt.Errorf("client certificate data is empty")
	}

	clientKeyBytes, err := clientConfig.ClientKey.GetData()
	if err != nil {
		log.Printf("Failed to get client key data: %v", err)
		return "", fmt.Errorf("failed to get client key data: %w", err)
	}
	if len(clientKeyBytes) == 0 {
		log.Printf("Client key data is empty")
		return "", fmt.Errorf("client key data is empty")
	}

	// Create tls.Certificate from bytes
	clientCertTLS, err := tls.X509KeyPair(clientCertBytes, clientKeyBytes)
	if err != nil {
		log.Printf("Failed to create client TLS certificate from bytes: %v", err)
		return "", fmt.Errorf("failed to create client TLS certificate from bytes: %w", err)
	}

	// Get CA certificate data
	caCertBytes, err := clientConfig.CACert.GetData()
	if err != nil {
		log.Printf("Failed to get CA certificate data: %v", err)
		return "", fmt.Errorf("failed to get CA certificate data: %w", err)
	}
	if len(caCertBytes) == 0 {
		log.Printf("CA certificate data is empty")
		return "", fmt.Errorf("CA certificate data is empty")
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertBytes); !ok {
		log.Printf("Failed to append CA certs to the pool from bytes: no valid certificates found")
		return "", fmt.Errorf("failed to append CA certs to the pool from bytes: no valid certificates found")
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCertTLS},
		RootCAs:      caCertPool,
	})

	// Default options: mTLS credentials.
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}
	// Append any additional options passed by the caller.
	allOpts := append(defaultOpts, opts...)

	// Check if the context is already done before attempting to dial.
	if err := ctx.Err(); err != nil {
		log.Printf("gRPC dial context for %s already expired/canceled: %v", ServerAddress, err)
		return "", fmt.Errorf("gRPC dial context for %s already expired/canceled: %w", ServerAddress, err)
	}

	// Create a client connection to the gRPC server using the combined options,
	// via GRPCDialContextFunc (which defaults to grpc.NewClient).
	// See the comment on GRPCDialContextFunc for how context.Context is handled.
	target := "passthrough:///" + ServerAddress
	conn, err := GRPCDialContextFunc(target, allOpts...)
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
	retrieveResp, err := client.Retrieve(ctx, retrieveReq)
	if err != nil {
		log.Printf("could not retrieve value for key '%s': %v", key, err)
		return "", fmt.Errorf("gRPC retrieve call failed: %w", err)
	}
	if retrieveResp == nil {
		log.Printf("received nil response for key '%s'", key)
		return "", fmt.Errorf("received nil response from gRPC server for key '%s'", key)
	}
	return retrieveResp.GetValue(), nil
}

// GrpcSimpleStore stores a key-value pair using gRPC.
// Accepts a context for timeout/cancellation and optional grpc.DialOptions.
func GrpcSimpleStore(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, value string, opts ...grpc.DialOption) (err error) { //nolint:all // Existing function
	// Ensure context has a deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultGrpcTimeout)
		defer cancel()
	}

	// Default options: insecure credentials.
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	// Append any additional options passed by the caller.
	allOpts := append(defaultOpts, opts...)

	// Check if the context is already done before attempting to dial.
	if err := ctx.Err(); err != nil {
		log.Printf("gRPC dial context for %s already expired/canceled: %v", ServerAddress, err)
		return fmt.Errorf("gRPC dial context for %s already expired/canceled: %w", ServerAddress, err)
	}

	// Create a client connection to the gRPC server using the combined options,
	// via GRPCDialContextFunc (which defaults to grpc.NewClient).
	// See the comment on GRPCDialContextFunc for how context.Context is handled.
	target := "passthrough:///" + ServerAddress
	conn, err := GRPCDialContextFunc(target, allOpts...)
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

// GrpcSimpleStoreWithMTLS stores a key-value pair using gRPC with mTLS.
// Accepts a context for timeout/cancellation, certificate file paths, and optional grpc.DialOptions.
// TODO: This function also needs to be updated to use clientConfig *psconfig.ParameterStoreClient if it's intended to be used with the new config structure.
// For now, leaving its signature as is, as it was not directly part of the user's request for the retrieve operation.
func GrpcSimpleStoreWithMTLS(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, value string, clientCertFile string, clientKeyFile string, caCertFile string, opts ...grpc.DialOption) (err error) {
	// Ensure context has a deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultGrpcTimeout)
		defer cancel()
	}

	// Load client's certificate and private key
	clientCertTLS, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile) // Corrected variable name if this was a copy-paste error, should be clientCertTLS
	if err != nil {
		log.Printf("Failed to load client cert: %v", err)
		return fmt.Errorf("failed to load client cert: %w", err)
	}

	// Load CA cert
	caCertBytes, err := os.ReadFile(caCertFile) // Changed ioutil.ReadFile to os.ReadFile
	if err != nil {
		log.Printf("Failed to load CA cert: %v", err)
		return fmt.Errorf("failed to load CA cert: %w", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertBytes) // Use caCertBytes

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{clientCertTLS}, // Use clientCertTLS
		RootCAs:      caCertPool,
	})

	// Default options: mTLS credentials.
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}
	// Append any additional options passed by the caller.
	allOpts := append(defaultOpts, opts...)

	// Check if the context is already done before attempting to dial.
	if err := ctx.Err(); err != nil {
		log.Printf("gRPC dial context for %s already expired/canceled: %v", ServerAddress, err)
		return fmt.Errorf("gRPC dial context for %s already expired/canceled: %w", ServerAddress, err)
	}

	// Create a client connection to the gRPC server using the combined options,
	// via GRPCDialContextFunc (which defaults to grpc.NewClient).
	// See the comment on GRPCDialContextFunc for how context.Context is handled.
	target := "passthrough:///" + ServerAddress
	conn, err := GRPCDialContextFunc(target, allOpts...)
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
	storeResp, err := client.Store(ctx, storeReq)
	if err != nil {
		log.Printf("could not store value for key '%s': %v", key, err)
		return fmt.Errorf("gRPC store call failed: %w", err)
	}
	log.Printf("Store response for key '%s': %s", key, storeResp.GetMessage())
	return nil
}
