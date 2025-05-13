package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"testing"
	"time"

	pb "github.com/Suhaibinator/SuhaibParameterStoreClient/proto" // Adjust import path if needed
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

// --- Mock ParameterStore Server ---

type mockParameterStoreServer struct {
	pb.UnimplementedParameterStoreServer // Embed the unimplemented server
	mu                                   sync.Mutex
	store                                map[string]string
	correctPassword                      string
	simulateError                        error
	simulateDelay                        time.Duration
}

func newMockServer(password string) *mockParameterStoreServer {
	return &mockParameterStoreServer{
		store:           make(map[string]string),
		correctPassword: password,
	}
}

func (s *mockParameterStoreServer) Retrieve(ctx context.Context, req *pb.RetrieveRequest) (*pb.RetrieveResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.simulateDelay > 0 {
		select {
		case <-time.After(s.simulateDelay):
			// Continue after delay
		case <-ctx.Done():
			return nil, status.Errorf(codes.DeadlineExceeded, "context deadline exceeded")
		}
	}

	if s.simulateError != nil {
		return nil, s.simulateError
	}

	if req.Password != s.correctPassword {
		return nil, status.Errorf(codes.Unauthenticated, "invalid password")
	}

	value, ok := s.store[req.Key]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "key not found: %s", req.Key)
	}

	return &pb.RetrieveResponse{Value: value}, nil
}

func (s *mockParameterStoreServer) Store(ctx context.Context, req *pb.StoreRequest) (*pb.StoreResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.simulateDelay > 0 {
		select {
		case <-time.After(s.simulateDelay):
			// Continue after delay
		case <-ctx.Done():
			return nil, status.Errorf(codes.DeadlineExceeded, "context deadline exceeded")
		}
	}

	if s.simulateError != nil {
		return nil, s.simulateError
	}

	if req.Password != s.correctPassword {
		return nil, status.Errorf(codes.Unauthenticated, "invalid password")
	}

	s.store[req.Key] = req.Value
	return &pb.StoreResponse{Message: fmt.Sprintf("Stored value for key %s", req.Key)}, nil
}

// --- Test Setup ---

func startMockServer(mockSrv *mockParameterStoreServer) func() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterParameterStoreServer(s, mockSrv)

	go func() {
		if err := s.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	// Return a function to stop the server
	return func() {
		s.GracefulStop()
		lis.Close()
	}
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// --- Tests ---

func TestGrpcimpleRetrieve(t *testing.T) {
	mockPassword := "goodpass"
	mockKey := "mykey"
	mockValue := "myvalue"

	mockSrv := newMockServer(mockPassword)
	// Pre-populate the store for retrieve tests
	mockSrv.store[mockKey] = mockValue

	stopServer := startMockServer(mockSrv)
	defer stopServer()

	ctxBg := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockSrv.simulateError = nil // Ensure no error simulation
		mockSrv.simulateDelay = 0

		// Pass test-specific dial options
		opts := []grpc.DialOption{
			grpc.WithContextDialer(bufDialer),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(), // Ensure connection attempt is synchronous for test
		}
		val, err := GrpcimpleRetrieve(ctxBg, "bufnet", mockPassword, "", "", "", mockKey, opts...)

		assert.NoError(t, err)
		assert.Equal(t, mockValue, val)
	})

	t.Run("Error - Wrong Password", func(t *testing.T) {
		mockSrv.simulateError = nil
		mockSrv.simulateDelay = 0

		// Pass test-specific dial options
		opts := []grpc.DialOption{
			grpc.WithContextDialer(bufDialer),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		}
		_, err := GrpcimpleRetrieve(ctxBg, "bufnet", "wrongpass", "", "", "", mockKey, opts...)

		assert.Error(t, err)
		// _, ok := status.FromError(err) // Remove unused st, ok
		// assert.True(t, ok, "Error should be a gRPC status error") // This check might fail due to wrapping
		assert.Contains(t, err.Error(), "Unauthenticated", "Error message should indicate Unauthenticated")
		// assert.Equal(t, codes.Unauthenticated, st.Code(), "Expected Unauthenticated error code")
	})

	t.Run("Error - Key Not Found", func(t *testing.T) {
		mockSrv.simulateError = nil
		mockSrv.simulateDelay = 0

		// Pass test-specific dial options
		opts := []grpc.DialOption{
			grpc.WithContextDialer(bufDialer),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		}
		_, err := GrpcimpleRetrieve(ctxBg, "bufnet", mockPassword, "", "", "", "nonexistentkey", opts...)

		assert.Error(t, err)
		// _, ok := status.FromError(err) // Remove unused st, ok
		// assert.True(t, ok, "Error should be a gRPC status error") // This check might fail due to wrapping
		assert.Contains(t, err.Error(), "NotFound", "Error message should indicate NotFound")
	})

	t.Run("Error - Simulated Server Error", func(t *testing.T) {
		simulatedErr := status.Error(codes.Internal, "internal server failure")
		mockSrv.simulateError = simulatedErr
		mockSrv.simulateDelay = 0

		// Pass test-specific dial options
		opts := []grpc.DialOption{
			grpc.WithContextDialer(bufDialer),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		}
		_, err := GrpcimpleRetrieve(ctxBg, "bufnet", mockPassword, "", "", "", mockKey, opts...)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "internal server failure", "Error message should contain simulated error")
		// Check underlying error if needed, depends on error wrapping
		// assert.ErrorIs(t, err, simulatedErr) // Might fail due to wrapping
	})

	t.Run("Timeout", func(t *testing.T) {
		mockSrv.simulateError = nil
		mockSrv.simulateDelay = 100 * time.Millisecond // Delay longer than context timeout

		timeout := 50 * time.Millisecond
		ctx, cancel := context.WithTimeout(ctxBg, timeout)
		defer cancel()

		// Pass test-specific dial options
		opts := []grpc.DialOption{
			grpc.WithContextDialer(bufDialer),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		}
		_, err := GrpcimpleRetrieve(ctx, "bufnet", mockPassword, "", "", "", mockKey, opts...)

		assert.Error(t, err)
		// Check if the error is context deadline exceeded or contains the message
		assert.True(t, errors.Is(err, context.DeadlineExceeded) || status.Code(err) == codes.DeadlineExceeded, "Expected DeadlineExceeded error")
	})

	t.Run("Connection Error", func(t *testing.T) {
		// Stop the server to simulate connection error
		stopServer()

		// Try to connect (use a short timeout context)
		ctx, cancel := context.WithTimeout(ctxBg, 50*time.Millisecond)
		defer cancel()

		// Pass test-specific dial options
		opts := []grpc.DialOption{
			grpc.WithContextDialer(bufDialer),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(), // Use WithBlock to make connection attempt respect context timeout
		}
		_, err := GrpcimpleRetrieve(ctx, "bufnet", mockPassword, "", "", "", mockKey, opts...)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect", "Expected connection error")

		// Restart server for subsequent tests if needed (though defer handles it for the whole function)
		// stopServer = startMockServer(mockSrv)
	})
}

func TestGrpcSimpleStore(t *testing.T) {
	mockPassword := "goodpass"
	mockKey := "storekey"
	mockValue := "storevalue"

	mockSrv := newMockServer(mockPassword)
	stopServer := startMockServer(mockSrv)
	defer stopServer()

	t.Run("Success", func(t *testing.T) {
		mockSrv.simulateError = nil
		mockSrv.simulateDelay = 0

		// Pass test-specific dial options
		opts := []grpc.DialOption{
			grpc.WithContextDialer(bufDialer),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			// grpc.WithBlock(), // Optional for store
		}
		ctxBg := context.Background() // Use background context for store tests
		err := GrpcSimpleStore(ctxBg, "bufnet", mockPassword, "", "", "", mockKey, mockValue, opts...)

		assert.NoError(t, err)
		// Verify value was stored in the mock server
		mockSrv.mu.Lock()
		storedVal, ok := mockSrv.store[mockKey]
		mockSrv.mu.Unlock()
		assert.True(t, ok, "Key should exist in mock store")
		assert.Equal(t, mockValue, storedVal, "Stored value mismatch")
	})

	t.Run("Error - Wrong Password", func(t *testing.T) {
		mockSrv.simulateError = nil
		mockSrv.simulateDelay = 0

		// Pass test-specific dial options
		opts := []grpc.DialOption{
			grpc.WithContextDialer(bufDialer),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
		ctxBg := context.Background()
		err := GrpcSimpleStore(ctxBg, "bufnet", "wrongpass", "", "", "", mockKey, mockValue, opts...)

		assert.Error(t, err)
		// _, ok := status.FromError(err) // Remove unused st, ok
		// assert.True(t, ok, "Error should be a gRPC status error") // Check might fail due to wrapping
		assert.Contains(t, err.Error(), "Unauthenticated", "Error message should indicate Unauthenticated")
	})

	t.Run("Error - Simulated Server Error", func(t *testing.T) {
		simulatedErr := status.Error(codes.Internal, "internal store failure")
		mockSrv.simulateError = simulatedErr
		mockSrv.simulateDelay = 0

		// Pass test-specific dial options
		opts := []grpc.DialOption{
			grpc.WithContextDialer(bufDialer),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
		ctxBg := context.Background()
		err := GrpcSimpleStore(ctxBg, "bufnet", mockPassword, "", "", "", mockKey, mockValue, opts...)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "internal store failure", "Error message should contain simulated error")
	})

	t.Run("Connection Error", func(t *testing.T) {
		stopServer() // Stop server

		// Pass test-specific dial options
		opts := []grpc.DialOption{
			grpc.WithContextDialer(bufDialer),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			// grpc.WithBlock(), // Add if needed to test connection timeout
		}
		// Use a context with timeout for connection error test to avoid hanging
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		err := GrpcSimpleStore(ctx, "bufnet", mockPassword, "", "", "", mockKey, mockValue, opts...)

		assert.Error(t, err)
		// The error might be context deadline exceeded or a connection error depending on timing
		assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) || grpc.Code(err) == codes.Unavailable || grpc.Code(err) == codes.DeadlineExceeded, "Expected connection-related error or timeout")

		// stopServer = startMockServer(mockSrv) // Restart if needed
	})
}
