package config

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	// Mock client package (replace with actual import path if different)
)

// --- Mocks ---

// Store original functions to restore them after tests
var originalGrpcRetrieveFunc func(ctx context.Context, ServerAddress string, AuthenticationPassword string, caCertPath string, clientCertPath string, clientKeyPath string, key string, opts ...grpc.DialOption) (val string, err error)
var originalOsGetenvFunc func(key string) string
var mu sync.Mutex // Mutex to protect access to global function variables

func setupTest() {
	mu.Lock()
	// Store the original functions assigned to our package variables
	originalGrpcRetrieveFunc = grpcRetrieveFunc
	originalOsGetenvFunc = osGetenvFunc
	mu.Unlock()
}

func teardownTest() {
	mu.Lock()
	// Restore the original functions
	grpcRetrieveFunc = originalGrpcRetrieveFunc
	osGetenvFunc = originalOsGetenvFunc
	mu.Unlock()
}

// Define types for our mock functions for clarity
type mockGrpcRetrieveFunc func(ctx context.Context, ServerAddress string, AuthenticationPassword string, caCertPath string, clientCertPath string, clientKeyPath string, key string, opts ...grpc.DialOption) (val string, err error)
type mockGetenvFunc func(key string) string

// --- Tests ---

func TestParameterStoreConfig_Init(t *testing.T) {
	setupTest()
	defer teardownTest()

	// Define common test parameters
	testHost := "localhost"
	testPort := 1234
	testTimeout := 100 * time.Millisecond
	testKey := "test-key"
	testSecret := "test-secret"
	testEnvKey := "TEST_ENV_VAR"
	testEnvValue := "value-from-env"
	testParamValue := "value-from-param-store"

	tests := []struct {
		name                 string
		initialConfig        ParameterStoreConfig
		mockRetrieve         mockGrpcRetrieveFunc // Use the type defined above
		mockGetenv           mockGetenvFunc       // Use the type defined above
		expectedValue        string
		expectPanic          bool // To test log.Fatalf
		expectedPanicMessage string
	}{
		{
			name: "Value already present",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
				ParameterStoreValue:    "pre-filled-value",
			},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				return "", errors.New("should not be called") // Should not be called
			},
			mockGetenv: func(key string) string {
				return "" // Should not be called
			},
			expectedValue: "pre-filled-value",
			expectPanic:   false,
		},
		{
			name: "Success from Parameter Store",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
				ParameterStoreValue:    "", // Initially empty
			},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				assert.Equal(t, testKey, key)
				assert.Equal(t, testSecret, AuthenticationPassword)
				// Ignore opts in mock for simplicity
				return testParamValue, nil
			},
			mockGetenv: func(key string) string {
				return "" // Should not be called
			},
			expectedValue: testParamValue,
			expectPanic:   false,
		},
		{
			name: "Parameter Store Timeout, Success from Env Var",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
				ParameterStoreValue:    "",
			},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				// Simulate timeout
				select {
				case <-ctx.Done():
					return "", context.DeadlineExceeded // Simulate timeout error
				case <-time.After(testTimeout * 2): // Ensure this takes longer than testTimeout
					return "should-not-return", nil
				}
			},
			mockGetenv: func(key string) string {
				assert.Equal(t, testEnvKey, key)
				return testEnvValue
			},
			expectedValue: testEnvValue,
			expectPanic:   false,
		},
		{
			name: "Parameter Store Error (non-timeout), Success from Env Var",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
				ParameterStoreValue:    "",
			},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				return "", errors.New("some-grpc-error")
			},
			mockGetenv: func(key string) string {
				assert.Equal(t, testEnvKey, key)
				return testEnvValue
			},
			expectedValue: testEnvValue,
			expectPanic:   false,
		},
		{
			name: "Parameter Store returns empty string, Success from Env Var",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
				ParameterStoreValue:    "",
			},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				return "", nil // Success, but empty value
			},
			mockGetenv: func(key string) string {
				assert.Equal(t, testEnvKey, key)
				return testEnvValue
			},
			expectedValue: testEnvValue,
			expectPanic:   false,
		},
		{
			name: "Parameter Store Fails, Env Var Fails (Expect Panic)",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
				ParameterStoreValue:    "",
			},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				return "", errors.New("param-store-failed")
			},
			mockGetenv: func(key string) string {
				assert.Equal(t, testEnvKey, key)
				return "" // Env var also fails (empty)
			},
			expectedValue: "", // Value doesn't matter as it should panic
			expectPanic:   true,
			// expectedPanicMessage is no longer needed, we check the error type/message below
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set mocks for this test case by assigning to package variables
			mu.Lock()
			grpcRetrieveFunc = tt.mockRetrieve
			osGetenvFunc = tt.mockGetenv
			mu.Unlock()
			// Ensure mocks are reset even if test panics (using defer in setup/teardown is better)
			// defer teardownTest() // Already handled by outer defer

			config := tt.initialConfig // Make a copy

			if tt.expectPanic {
				// Use assert.PanicsWithError to check for the specific error message
				expectedErr := fmt.Sprintf("Failed to retrieve value for parameter store key '%s' (checked env var '%s'). Neither parameter store nor environment variable provided a value.", testKey, testEnvKey)
				assert.PanicsWithError(t, expectedErr, func() {
					config.Init(testHost, testPort, testTimeout)
				}, "Expected Init to panic with specific error")
			} else {
				// Use assert.NotPanics for non-panic cases
				assert.NotPanics(t, func() {
					config.Init(testHost, testPort, testTimeout)
				}, "Expected Init not to panic")
				assert.Equal(t, tt.expectedValue, config.ParameterStoreValue, "ParameterStoreValue mismatch")
			}
		})
	}
}

// Note: Testing simpleRetrieveParameterWithTimeout and setValueIfEmpty directly
// is implicitly covered by testing Init, as Init calls setValueIfEmpty which
// calls simpleRetrieveParameterWithTimeout. If more granular tests are needed,
// they can be added similarly, mocking dependencies as required.
