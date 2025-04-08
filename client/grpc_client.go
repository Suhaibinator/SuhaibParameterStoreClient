package client

import (
	"context"
	"fmt"
	"log"

	pb "github.com/Suhaibinator/SuhaibParameterStoreClient/proto"

	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Default timeout for gRPC operations if no context deadline is set.
const defaultGrpcTimeout = 5 * time.Second

// GrpcimpleRetrieve retrieves a value from the parameter store using gRPC.
// It accepts a context for timeout/cancellation control and optional grpc.DialOptions.
func GrpcimpleRetrieve(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
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
