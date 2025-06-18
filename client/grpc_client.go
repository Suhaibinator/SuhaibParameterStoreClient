package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	pb "github.com/Suhaibinator/SuhaibParameterStoreClient/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCDialContextFunc is a function variable that can be replaced for testing.
// It defaults to the standard grpc.NewClient function.
var GRPCDialContextFunc = grpc.NewClient

// Default timeout for gRPC operations if no context deadline is set.
const defaultGrpcTimeout = 5 * time.Second

// dial sets up a connection to the given gRPC server using the supplied options.
// It wraps GRPCDialContextFunc and provides consistent error handling across
// client helpers.
func dial(ctx context.Context, serverAddr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if err := ctx.Err(); err != nil {
		log.Printf("gRPC dial context for %s already expired/canceled: %v", serverAddr, err)
		return nil, fmt.Errorf("gRPC dial context for %s already expired/canceled: %w", serverAddr, err)
	}

	target := "passthrough:///" + serverAddr
	conn, err := GRPCDialContextFunc(target, opts...)
	if err != nil {
		log.Printf("did not connect to %s: %v", serverAddr, err)
		return nil, fmt.Errorf("failed to connect to gRPC server at %s: %w", serverAddr, err)
	}

	return conn, nil
}

// GrpcSimpleRetrieve retrieves a value from the parameter store using gRPC.
// It accepts a context for timeout/cancellation control and optional grpc.DialOptions.
func GrpcSimpleRetrieve(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
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

	// Establish connection using the helper dial function.
	var conn *grpc.ClientConn
	conn, err = dial(ctx, ServerAddress, allOpts...)
	if err != nil {
		return "", err
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
func GrpcSimpleStore(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, value string, opts ...grpc.DialOption) (err error) {
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

	// Establish connection using the helper dial function.
	var conn *grpc.ClientConn
	conn, err = dial(ctx, ServerAddress, allOpts...)
	if err != nil {
		return err
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

// GrpcSimpleRetrieveWithTLS retrieves a value from the parameter store using gRPC with TLS.
// It accepts a context, server details, credentials, a key, and TLS configuration.
func GrpcSimpleRetrieveWithTLS(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, tlsConfig *TLSConfig, opts ...grpc.DialOption) (val string, err error) {
	// Ensure context has a deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultGrpcTimeout)
		defer cancel()
	}

	// Get TLS configuration
	tlsConf, err := tlsConfig.GetTLSConfig()
	if err != nil {
		log.Printf("Failed to get TLS configuration: %v", err)
		return "", fmt.Errorf("failed to get TLS configuration: %w", err)
	}

	creds := credentials.NewTLS(tlsConf)

	// Default options: TLS credentials.
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}
	// Append any additional options passed by the caller.
	allOpts := append(defaultOpts, opts...)

	// Establish connection using the helper dial function.
	var conn *grpc.ClientConn
	conn, err = dial(ctx, ServerAddress, allOpts...)
	if err != nil {
		return "", err
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

// GrpcSimpleStoreWithTLS stores a key-value pair using gRPC with TLS.
// It accepts a context, server details, credentials, key-value pair, and TLS configuration.
func GrpcSimpleStoreWithTLS(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, value string, tlsConfig *TLSConfig, opts ...grpc.DialOption) (err error) {
	// Ensure context has a deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultGrpcTimeout)
		defer cancel()
	}

	// Get TLS configuration
	tlsConf, err := tlsConfig.GetTLSConfig()
	if err != nil {
		log.Printf("Failed to get TLS configuration: %v", err)
		return fmt.Errorf("failed to get TLS configuration: %w", err)
	}

	creds := credentials.NewTLS(tlsConf)

	// Default options: TLS credentials.
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}
	// Append any additional options passed by the caller.
	allOpts := append(defaultOpts, opts...)

	// Establish connection using the helper dial function.
	var conn *grpc.ClientConn
	conn, err = dial(ctx, ServerAddress, allOpts...)
	if err != nil {
		return err
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

// GrpcSimpleRetrieveWithPrebuiltTLS retrieves a value from the parameter store using gRPC with pre-built TLS config.
// This version accepts a pre-built *tls.Config to avoid password prompts within timeout context.
func GrpcSimpleRetrieveWithPrebuiltTLS(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, tlsConf *tls.Config, opts ...grpc.DialOption) (val string, err error) {
	// Ensure context has a deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultGrpcTimeout)
		defer cancel()
	}

	creds := credentials.NewTLS(tlsConf)

	// Default options: TLS credentials.
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}
	// Append any additional options passed by the caller.
	allOpts := append(defaultOpts, opts...)

	// Establish connection using the helper dial function.
	var conn *grpc.ClientConn
	conn, err = dial(ctx, ServerAddress, allOpts...)
	if err != nil {
		return "", err
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

// GrpcSimpleStoreWithPrebuiltTLS stores a key-value pair using gRPC with pre-built TLS config.
// This version accepts a pre-built *tls.Config to avoid password prompts within timeout context.
func GrpcSimpleStoreWithPrebuiltTLS(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, value string, tlsConf *tls.Config, opts ...grpc.DialOption) (err error) {
	// Ensure context has a deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultGrpcTimeout)
		defer cancel()
	}

	creds := credentials.NewTLS(tlsConf)

	// Default options: TLS credentials.
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}
	// Append any additional options passed by the caller.
	allOpts := append(defaultOpts, opts...)

	// Establish connection using the helper dial function.
	var conn *grpc.ClientConn
	conn, err = dial(ctx, ServerAddress, allOpts...)
	if err != nil {
		return err
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
