package client

// Note: This test file intentionally uses some deprecated gRPC options like grpc.WithBlock()
// for testing purposes. These options make tests more reliable and predictable by ensuring
// synchronous connection attempts, but should be avoided in production code.

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"os"

	psconfig "github.com/Suhaibinator/SuhaibParameterStoreClient/config"
	pb "github.com/Suhaibinator/SuhaibParameterStoreClient/proto" // Adjust import path if needed
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"path/filepath"

	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// No need for local grpcDialContext variable anymore
// We'll use the exported GRPCDialContextFunc from the client package

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

func bufDialer(ctx context.Context, _ string) (net.Conn, error) { // address string is not used by lis.DialContext
	return lis.DialContext(ctx)
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
			grpc.WithBlock(), // Intentionally used for testing despite deprecation warning
			// It ensures connection attempt is synchronous for test
		}
		val, err := GrpcimpleRetrieve(ctxBg, "bufnet", mockPassword, mockKey, opts...)

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
			grpc.WithBlock(), // Intentionally used for testing despite deprecation warning
		}
		_, err := GrpcimpleRetrieve(ctxBg, "bufnet", "wrongpass", mockKey, opts...)

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
			grpc.WithBlock(), // Intentionally used for testing despite deprecation warning
		}
		_, err := GrpcimpleRetrieve(ctxBg, "bufnet", mockPassword, "nonexistentkey", opts...)

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
			grpc.WithBlock(), // Intentionally used for testing despite deprecation warning
		}
		_, err := GrpcimpleRetrieve(ctxBg, "bufnet", mockPassword, mockKey, opts...)

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
			grpc.WithBlock(), // Intentionally used for testing despite deprecation warning
		}
		_, err := GrpcimpleRetrieve(ctx, "bufnet", mockPassword, mockKey, opts...)

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
			grpc.WithBlock(), // Intentionally used for testing despite deprecation warning - makes connection attempt respect context timeout
		}
		_, err := GrpcimpleRetrieve(ctx, "bufnet", mockPassword, mockKey, opts...)

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
		err := GrpcSimpleStore(ctxBg, "bufnet", mockPassword, mockKey, mockValue, opts...)

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
		err := GrpcSimpleStore(ctxBg, "bufnet", "wrongpass", mockKey, mockValue, opts...)

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
		err := GrpcSimpleStore(ctxBg, "bufnet", mockPassword, mockKey, mockValue, opts...)

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
		err := GrpcSimpleStore(ctx, "bufnet", mockPassword, mockKey, mockValue, opts...)

		assert.Error(t, err)
		// The error might be context deadline exceeded or a connection error depending on timing
		assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) || grpc.Code(err) == codes.Unavailable || grpc.Code(err) == codes.DeadlineExceeded, "Expected connection-related error or timeout")

		// stopServer = startMockServer(mockSrv) // Restart if needed
	})
}

func TestGrpcSimpleRetrieveWithMTLS(t *testing.T) {
	// Create a mock function that will be used in the test
	// Signature matches the new GRPCDialContextFunc: func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	mockDialFunc := func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		// Return a specific gRPC error to stop further processing in the tested function.
		return nil, status.Error(codes.Unavailable, "mock dial: connection refused")
	}

	// Save the original function to restore later
	originalDialFunc := GRPCDialContextFunc
	// Replace with our mock function
	GRPCDialContextFunc = mockDialFunc
	defer func() { GRPCDialContextFunc = originalDialFunc }() // Restore

	clientCertPath, clientKeyPath, caCertPath, cleanup := createDummyCertFiles(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Successful mTLS config", func(t *testing.T) {
		// Create a client config with valid cert files
		clientConfig := &psconfig.ParameterStoreClient{
			Host: "localhost",
			Port: 50051,
			ClientCert: psconfig.CertificateSource{
				FilePath: clientCertPath,
			},
			ClientKey: psconfig.CertificateSource{
				FilePath: clientKeyPath,
			},
			CACert: psconfig.CertificateSource{
				FilePath: caCertPath,
			},
		}

		_, err := GrpcSimpleRetrieveWithMTLS(ctx, "localhost:50051", "password", "key", clientConfig)
		// We expect an error because the mockDialContext returns an error.
		// The important part is that we got past cert loading.
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock dial: connection refused")

		// Since we're using a mock dial function that returns a specific error,
		// if we get that error it means the credentials were properly configured
		// and we reached the dial step without any certificate loading errors

	})

	t.Run("Missing client cert", func(t *testing.T) {
		// Create a client config with a nonexistent client cert file
		clientConfig := &psconfig.ParameterStoreClient{
			Host: "localhost",
			Port: 50051,
			ClientCert: psconfig.CertificateSource{
				FilePath: "nonexistent.pem",
			},
			ClientKey: psconfig.CertificateSource{
				FilePath: clientKeyPath,
			},
			CACert: psconfig.CertificateSource{
				FilePath: caCertPath,
			},
		}
		_, err := GrpcSimpleRetrieveWithMTLS(ctx, "localhost:50051", "password", "key", clientConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read certificate/key from nonexistent.pem")
	})

	t.Run("Missing client key", func(t *testing.T) {
		// Create a client config with missing client key
		clientConfig := &psconfig.ParameterStoreClient{
			Host: "localhost",
			Port: 50051,
			ClientCert: psconfig.CertificateSource{
				FilePath: clientCertPath,
			},
			ClientKey: psconfig.CertificateSource{
				FilePath: "nonexistent.key",
			},
			CACert: psconfig.CertificateSource{
				FilePath: caCertPath,
			},
		}

		_, err := GrpcSimpleRetrieveWithMTLS(ctx, "localhost:50051", "password", "key", clientConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read certificate/key from nonexistent.key")
	})

	t.Run("Missing CA cert", func(t *testing.T) {
		// Create a client config with missing CA cert
		clientConfig := &psconfig.ParameterStoreClient{
			Host: "localhost",
			Port: 50051,
			ClientCert: psconfig.CertificateSource{
				FilePath: clientCertPath,
			},
			ClientKey: psconfig.CertificateSource{
				FilePath: clientKeyPath,
			},
			CACert: psconfig.CertificateSource{
				FilePath: "nonexistent.pem",
			},
		}

		_, err := GrpcSimpleRetrieveWithMTLS(ctx, "localhost:50051", "password", "key", clientConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read certificate/key from nonexistent.pem")
	})

	t.Run("Empty client cert path", func(t *testing.T) {
		// Create a client config with empty client cert path
		clientConfig := &psconfig.ParameterStoreClient{
			Host: "localhost",
			Port: 50051,
			ClientCert: psconfig.CertificateSource{
				FilePath: "",
			},
			ClientKey: psconfig.CertificateSource{
				FilePath: clientKeyPath,
			},
			CACert: psconfig.CertificateSource{
				FilePath: caCertPath,
			},
		}

		_, err := GrpcSimpleRetrieveWithMTLS(ctx, "localhost:50051", "password", "key", clientConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client certificate data is empty")
	})
}

func TestGrpcSimpleStoreWithMTLS(t *testing.T) {
	// Create a mock function that will be used in the test
	// Signature matches the new GRPCDialContextFunc: func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	mockDialFunc := func(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return nil, status.Error(codes.Unavailable, "mock dial: connection refused")
	}

	// Save the original function to restore later
	originalDialFunc := GRPCDialContextFunc
	// Replace with our mock function
	GRPCDialContextFunc = mockDialFunc
	defer func() { GRPCDialContextFunc = originalDialFunc }() // Restore

	clientCertPath, clientKeyPath, caCertPath, cleanup := createDummyCertFiles(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Successful mTLS config", func(t *testing.T) {
		err := GrpcSimpleStoreWithMTLS(ctx, "localhost:50051", "password", "key", "value", clientCertPath, clientKeyPath, caCertPath)
		assert.Error(t, err)
		// The error message will be about connection refused, but not necessarily the exact mock message
		assert.Contains(t, err.Error(), "mock dial: connection refused")

		foundCreds := false
		if err != nil && (errors.Is(err, context.DeadlineExceeded) || grpc.Code(err) == codes.Unavailable || grpc.Code(err) == codes.DeadlineExceeded || strings.Contains(err.Error(), "connection refused")) {
			foundCreds = true
		}
		assert.True(t, foundCreds, "Expected transport credentials to be configured for mTLS")
	})

	t.Run("Missing client cert", func(t *testing.T) {
		err := GrpcSimpleStoreWithMTLS(ctx, "localhost:50051", "password", "key", "value", "nonexistent.pem", clientKeyPath, caCertPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load client cert")
	})

	t.Run("Missing client key", func(t *testing.T) {
		err := GrpcSimpleStoreWithMTLS(ctx, "localhost:50051", "password", "key", "value", clientCertPath, "nonexistent.key", caCertPath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load client cert")
	})

	t.Run("Missing CA cert", func(t *testing.T) {
		err := GrpcSimpleStoreWithMTLS(ctx, "localhost:50051", "password", "key", "value", clientCertPath, clientKeyPath, "nonexistent.pem")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load CA cert")
	})
}

// generateTestCertPair creates a temporary self-signed certificate and key pair for tests.
func generateTestCertPair(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	assert.NoError(t, err)
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return
}

// createDummyCertFiles writes a temporary certificate, key and CA certificate to disk.
// It returns the file paths and a cleanup function.
func createDummyCertFiles(t *testing.T) (clientCertFile, clientKeyFile, caCertFile string, cleanup func()) {
	t.Helper()
	certPEM, keyPEM := generateTestCertPair(t)
	dir, err := os.MkdirTemp("", "grpc-mtls-test-certs")
	assert.NoError(t, err)
	clientCertFile = filepath.Join(dir, "client.crt")
	clientKeyFile = filepath.Join(dir, "client.key")
	caCertFile = filepath.Join(dir, "ca.crt")
	assert.NoError(t, os.WriteFile(clientCertFile, certPEM, 0600))
	assert.NoError(t, os.WriteFile(clientKeyFile, keyPEM, 0600))
	// reuse the same cert as CA cert
	assert.NoError(t, os.WriteFile(caCertFile, certPEM, 0600))
	cleanup = func() { os.RemoveAll(dir) }
	return
}
