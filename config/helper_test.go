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
)

// --- Mocks ---

// Store original functions to restore them after tests
var (
	originalGrpcSimpleRetrieveFunc         func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error)
	originalGrpcSimpleRetrieveWithMTLSFunc func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientConfig *ParameterStoreClient, opts ...grpc.DialOption) (val string, err error)
	originalOsGetenvFunc                   func(key string) string
	mu                                     sync.Mutex // Mutex to protect access to global function variables
)

func setupTest() {
	mu.Lock()
	// Store the original functions assigned to our package variables
	originalGrpcSimpleRetrieveFunc = GrpcSimpleRetrieveFunc
	originalGrpcSimpleRetrieveWithMTLSFunc = GrpcSimpleRetrieveWithMTLSFunc
	originalOsGetenvFunc = osGetenvFunc
	mu.Unlock()
}

func teardownTest() {
	mu.Lock()
	// Restore the original functions
	GrpcSimpleRetrieveFunc = originalGrpcSimpleRetrieveFunc
	GrpcSimpleRetrieveWithMTLSFunc = originalGrpcSimpleRetrieveWithMTLSFunc
	osGetenvFunc = originalOsGetenvFunc
	mu.Unlock()
}

// Define types for our mock functions for clarity
type mockGrpcRetrieveFuncType func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error)
type mockGrpcRetrieveWithMTLSFuncType func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientConfig *ParameterStoreClient, opts ...grpc.DialOption) (val string, err error)
type mockGetenvFuncType func(key string) string

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
	testMtlsParamValue := "value-from-mtls-param-store"
	dummyClientCert := "client.crt"
	dummyClientKey := "client.key"
	dummyCACert := "ca.crt"

	tests := []struct {
		name                 string
		initialConfig        ParameterStoreConfig
		prePopulatedValue    string
		useEmptyValue        bool
		mockRetrieve         mockGrpcRetrieveFuncType
		mockRetrieveMTLS     mockGrpcRetrieveWithMTLSFuncType
		mockGetenv           mockGetenvFuncType
		expectedValue        string
		expectPanic          bool
		expectedPanicMessage string
		client               ParameterStoreClient
	}{
		{
			name: "Value already present",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			client:            ParameterStoreClient{Host: testHost, Port: testPort, Timeout: testTimeout},
			prePopulatedValue: "pre-filled-value",
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				t.Errorf("grpcSimpleRetrieveFunc should not be called when value is pre-filled")
				return "", errors.New("should not be called")
			},
			mockRetrieveMTLS: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientConfig *ParameterStoreClient, opts ...grpc.DialOption) (val string, err error) {
				t.Errorf("GrpcSimpleRetrieveWithMTLSFunc should not be called when value is pre-filled")
				return "", errors.New("should not be called")
			},
			mockGetenv: func(key string) string {
				t.Errorf("osGetenvFunc should not be called when value is pre-filled")
				return ""
			},
			expectedValue: "pre-filled-value",
			expectPanic:   false,
		},
		{
			name: "Use Empty Value flag is true and value is empty",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			client:        ParameterStoreClient{Host: testHost, Port: testPort, Timeout: testTimeout},
			useEmptyValue: true,
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				t.Errorf("grpcSimpleRetrieveFunc should not be called when UseEmptyValue is true and value is initially empty")
				return "", errors.New("should not be called")
			},
			mockRetrieveMTLS: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientConfig *ParameterStoreClient, opts ...grpc.DialOption) (val string, err error) {
				t.Errorf("GrpcSimpleRetrieveWithMTLSFunc should not be called when UseEmptyValue is true and value is initially empty")
				return "", errors.New("should not be called")
			},
			mockGetenv: func(key string) string {
				t.Errorf("osGetenvFunc should not be called when UseEmptyValue is true and value is initially empty")
				return ""
			},
			expectedValue: "",
			expectPanic:   false,
		},
		{
			name: "Success from Parameter Store (No mTLS)",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			client: ParameterStoreClient{Host: testHost, Port: testPort, Timeout: testTimeout},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				assert.Equal(t, testKey, key)
				assert.Equal(t, testSecret, AuthenticationPassword)
				assert.Equal(t, fmt.Sprintf("%s:%d", testHost, testPort), ServerAddress)
				return testParamValue, nil
			},
			mockRetrieveMTLS: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientConfig *ParameterStoreClient, opts ...grpc.DialOption) (val string, err error) {
				t.Errorf("GrpcSimpleRetrieveWithMTLSFunc should not be called when mTLS certs are not provided")
				return "", errors.New("mTLS func called unexpectedly")
			},
			mockGetenv: func(key string) string {
				t.Errorf("osGetenvFunc should not be called")
				return ""
			},
			expectedValue: testParamValue,
			expectPanic:   false,
		},
		{
			name: "Success from Parameter Store (with mTLS)",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			client: ParameterStoreClient{
				Host:    testHost,
				Port:    testPort,
				Timeout: testTimeout,
				ClientCert: CertificateSource{
					FilePath: dummyClientCert,
				},
				ClientKey: CertificateSource{
					FilePath: dummyClientKey,
				},
				CACert: CertificateSource{
					FilePath: dummyCACert,
				},
			},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				t.Errorf("GrpcSimpleRetrieveFunc should not be called when mTLS certs are provided")
				return "", errors.New("non-mTLS func called unexpectedly")
			},
			mockRetrieveMTLS: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, keyParam string, clientConfig *ParameterStoreClient, opts ...grpc.DialOption) (val string, err error) {
				assert.Equal(t, testKey, keyParam)
				assert.Equal(t, testSecret, AuthenticationPassword)
				assert.Equal(t, fmt.Sprintf("%s:%d", testHost, testPort), ServerAddress)
				assert.Equal(t, dummyClientCert, clientConfig.ClientCert.FilePath)
				assert.Equal(t, dummyClientKey, clientConfig.ClientKey.FilePath)
				assert.Equal(t, dummyCACert, clientConfig.CACert.FilePath)
				return testMtlsParamValue, nil
			},
			mockGetenv: func(key string) string {
				t.Errorf("osGetenvFunc should not be called")
				return ""
			},
			expectedValue: testMtlsParamValue,
			expectPanic:   false,
		},
		{
			name: "Parameter Store Timeout (No mTLS), Success from Env Var",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			client: ParameterStoreClient{Host: testHost, Port: testPort, Timeout: testTimeout},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				select {
				case <-ctx.Done():
					return "", context.DeadlineExceeded
				case <-time.After(testTimeout * 2):
					t.Errorf("gRPC call should have timed out")
					return "should-not-return", nil
				}
			},
			mockRetrieveMTLS: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientConfig *ParameterStoreClient, opts ...grpc.DialOption) (val string, err error) {
				t.Errorf("GrpcSimpleRetrieveWithMTLSFunc should not be called")
				return "", errors.New("mTLS func called unexpectedly")
			},
			mockGetenv: func(key string) string {
				assert.Equal(t, testEnvKey, key)
				return testEnvValue
			},
			expectedValue: testEnvValue,
			expectPanic:   false,
		},
		{
			name: "Parameter Store Timeout (mTLS), Success from Env Var",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			client: ParameterStoreClient{
				Host:    testHost,
				Port:    testPort,
				Timeout: testTimeout,
				ClientCert: CertificateSource{
					FilePath: dummyClientCert,
				},
				ClientKey: CertificateSource{
					FilePath: dummyClientKey,
				},
				CACert: CertificateSource{
					FilePath: dummyCACert,
				},
			},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				t.Errorf("GrpcSimpleRetrieveFunc should not be called")
				return "", errors.New("non-mTLS func called unexpectedly")
			},
			mockRetrieveMTLS: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientConfig *ParameterStoreClient, opts ...grpc.DialOption) (val string, err error) {
				assert.Equal(t, dummyClientCert, clientConfig.ClientCert.FilePath)
				select {
				case <-ctx.Done():
					return "", context.DeadlineExceeded
				case <-time.After(testTimeout * 2):
					t.Errorf("gRPC mTLS call should have timed out")
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
			name: "Parameter Store Error (non-timeout, No mTLS), Success from Env Var",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			client: ParameterStoreClient{Host: testHost, Port: testPort, Timeout: testTimeout},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				return "", errors.New("some-grpc-error")
			},
			mockRetrieveMTLS: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientConfig *ParameterStoreClient, opts ...grpc.DialOption) (val string, err error) {
				t.Errorf("GrpcSimpleRetrieveWithMTLSFunc should not be called")
				return "", errors.New("mTLS func called unexpectedly")
			},
			mockGetenv: func(key string) string {
				assert.Equal(t, testEnvKey, key)
				return testEnvValue
			},
			expectedValue: testEnvValue,
			expectPanic:   false,
		},
		{
			name: "Parameter Store Error (non-timeout, mTLS), Success from Env Var",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			client: ParameterStoreClient{
				Host:    testHost,
				Port:    testPort,
				Timeout: testTimeout,
				ClientCert: CertificateSource{
					FilePath: dummyClientCert,
				},
				ClientKey: CertificateSource{
					FilePath: dummyClientKey,
				},
				CACert: CertificateSource{
					FilePath: dummyCACert,
				},
			},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				t.Errorf("GrpcSimpleRetrieveFunc should not be called")
				return "", errors.New("non-mTLS func called unexpectedly")
			},
			mockRetrieveMTLS: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientConfig *ParameterStoreClient, opts ...grpc.DialOption) (val string, err error) {
				assert.Equal(t, dummyClientCert, clientConfig.ClientCert.FilePath)
				return "", errors.New("some-grpc-mtls-error")
			},
			mockGetenv: func(key string) string {
				assert.Equal(t, testEnvKey, key)
				return testEnvValue
			},
			expectedValue: testEnvValue,
			expectPanic:   false,
		},
		{
			name: "Parameter Store returns empty string (No mTLS), Success from Env Var",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			client: ParameterStoreClient{Host: testHost, Port: testPort, Timeout: testTimeout},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				return "", nil // Success, but empty value
			},
			mockRetrieveMTLS: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientConfig *ParameterStoreClient, opts ...grpc.DialOption) (val string, err error) {
				t.Errorf("GrpcSimpleRetrieveWithMTLSFunc should not be called")
				return "", errors.New("mTLS func called unexpectedly")
			},
			mockGetenv: func(key string) string {
				assert.Equal(t, testEnvKey, key)
				return testEnvValue
			},
			expectedValue: testEnvValue,
			expectPanic:   false,
		},
		{
			name: "Parameter Store returns empty string (mTLS), Success from Env Var",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			client: ParameterStoreClient{
				Host:    testHost,
				Port:    testPort,
				Timeout: testTimeout,
				ClientCert: CertificateSource{
					FilePath: dummyClientCert,
				},
				ClientKey: CertificateSource{
					FilePath: dummyClientKey,
				},
				CACert: CertificateSource{
					FilePath: dummyCACert,
				},
			},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				t.Errorf("GrpcSimpleRetrieveFunc should not be called")
				return "", errors.New("non-mTLS func called unexpectedly")
			},
			mockRetrieveMTLS: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientConfig *ParameterStoreClient, opts ...grpc.DialOption) (val string, err error) {
				assert.Equal(t, dummyClientCert, clientConfig.ClientCert.FilePath)
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
			name: "Parameter Store Fails (No mTLS), Env Var Fails (Expect Panic)",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			client: ParameterStoreClient{Host: testHost, Port: testPort, Timeout: testTimeout},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				return "", errors.New("param-store-failed")
			},
			mockRetrieveMTLS: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientConfig *ParameterStoreClient, opts ...grpc.DialOption) (val string, err error) {
				t.Errorf("GrpcSimpleRetrieveWithMTLSFunc should not be called")
				return "", errors.New("mTLS func called unexpectedly")
			},
			mockGetenv: func(key string) string {
				assert.Equal(t, testEnvKey, key)
				return "" // Env var also fails (empty)
			},
			expectPanic:          true,
			expectedPanicMessage: fmt.Sprintf("Failed to retrieve value for parameter store key '%s' (checked env var '%s'). Neither parameter store nor environment variable provided a value.", testKey, testEnvKey),
		},
		{
			name: "Parameter Store Fails (mTLS), Env Var Fails (Expect Panic)",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			client: ParameterStoreClient{
				Host:    testHost,
				Port:    testPort,
				Timeout: testTimeout,
				ClientCert: CertificateSource{
					FilePath: dummyClientCert,
				},
				ClientKey: CertificateSource{
					FilePath: dummyClientKey,
				},
				CACert: CertificateSource{
					FilePath: dummyCACert,
				},
			},
			mockRetrieve: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) {
				t.Errorf("GrpcSimpleRetrieveFunc should not be called")
				return "", errors.New("non-mTLS func called unexpectedly")
			},
			mockRetrieveMTLS: func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientConfig *ParameterStoreClient, opts ...grpc.DialOption) (val string, err error) {
				assert.Equal(t, dummyClientCert, clientConfig.ClientCert.FilePath)
				return "", errors.New("param-store-mtls-failed")
			},
			mockGetenv: func(key string) string {
				assert.Equal(t, testEnvKey, key)
				return "" // Env var also fails (empty)
			},
			expectPanic:          true,
			expectedPanicMessage: fmt.Sprintf("Failed to retrieve value for parameter store key '%s' (checked env var '%s'). Neither parameter store nor environment variable provided a value.", testKey, testEnvKey),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set mocks for this test case by assigning to package variables
			mu.Lock()
			GrpcSimpleRetrieveFunc = tt.mockRetrieve
			GrpcSimpleRetrieveWithMTLSFunc = tt.mockRetrieveMTLS
			osGetenvFunc = tt.mockGetenv
			mu.Unlock()

			config := tt.initialConfig // Make a copy
			if tt.prePopulatedValue != "" {
				config.ParameterStoreValue = tt.prePopulatedValue
			}
			config.ParameterStoreUseEmptyValue = tt.useEmptyValue

			if tt.expectPanic {
				assert.PanicsWithError(t, tt.expectedPanicMessage, func() {
					config.Init(&tt.client)
				}, "Expected Init to panic with specific error")
			} else {
				assert.NotPanics(t, func() {
					config.Init(&tt.client)
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
