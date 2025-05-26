package config

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

// Store original functions to restore them after tests
var (
	originalOsGetenvFunc func(key string) string
	mu                   sync.Mutex // Mutex to protect access to global function variables
)

func setupTest() {
	mu.Lock()
	// Store the original functions assigned to our package variables
	originalOsGetenvFunc = osGetenvFunc
	mu.Unlock()
}

func teardownTest() {
	mu.Lock()
	// Restore the original functions
	osGetenvFunc = originalOsGetenvFunc
	mu.Unlock()
}

// MockClient implements a mock parameter store client for testing
type MockClient struct {
	RetrieveFunc func(key, secret string) (string, error)
}

func (m *MockClient) Retrieve(key, secret string) (string, error) {
	if m.RetrieveFunc != nil {
		return m.RetrieveFunc(key, secret)
	}
	return "", errors.New("mock retrieve not implemented")
}

// --- Tests ---

func TestParameterStoreConfig_Init(t *testing.T) {
	setupTest()
	defer teardownTest()

	// Define common test parameters
	testKey := "test-key"
	testSecret := "test-secret"
	testEnvKey := "TEST_ENV_VAR"
	testEnvValue := "env-value"

	tests := []struct {
		name                 string
		initialConfig        ParameterStoreConfig
		prePopulatedValue    string
		useEmptyValue        bool
		mockRetrieve         func(key, secret string) (string, error)
		mockGetenv           func(key string) string
		expectedValue        string
		expectPanic          bool
		expectedPanicMessage string
	}{
		{
			name: "Value already present",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			prePopulatedValue: "pre-filled-value",
			mockRetrieve: func(key, secret string) (string, error) {
				t.Errorf("RetrieveFunc should not be called when value is pre-filled")
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
			useEmptyValue: true,
			mockRetrieve: func(key, secret string) (string, error) {
				t.Errorf("RetrieveFunc should not be called when use empty value flag is true")
				return "", errors.New("should not be called")
			},
			mockGetenv: func(key string) string {
				t.Errorf("osGetenvFunc should not be called when use empty value flag is true")
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
			mockRetrieve: func(key, secret string) (string, error) {
				assert.Equal(t, testKey, key)
				assert.Equal(t, testSecret, secret)
				return "param-store-value", nil
			},
			mockGetenv: func(key string) string {
				t.Errorf("osGetenvFunc should not be called when param store succeeds")
				return ""
			},
			expectedValue: "param-store-value",
			expectPanic:   false,
		},
		{
			name: "Success from Parameter Store (with mTLS)",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			mockRetrieve: func(key, secret string) (string, error) {
				assert.Equal(t, testKey, key)
				assert.Equal(t, testSecret, secret)
				return "param-store-value-mtls", nil
			},
			mockGetenv: func(key string) string {
				t.Errorf("osGetenvFunc should not be called when param store succeeds")
				return ""
			},
			expectedValue: "param-store-value-mtls",
			expectPanic:   false,
		},
		{
			name: "Parameter Store Timeout (No mTLS), Success from Env Var",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			mockRetrieve: func(key, secret string) (string, error) {
				return "", context.DeadlineExceeded
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
			mockRetrieve: func(key, secret string) (string, error) {
				return "", context.DeadlineExceeded
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
			mockRetrieve: func(key, secret string) (string, error) {
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
			name: "Parameter Store Error (non-timeout, mTLS), Success from Env Var",
			initialConfig: ParameterStoreConfig{
				ParameterStoreKey:      testKey,
				ParameterStoreSecret:   testSecret,
				EnvironmentVariableKey: testEnvKey,
			},
			mockRetrieve: func(key, secret string) (string, error) {
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
			mockRetrieve: func(key, secret string) (string, error) {
				return "", nil // No error, but empty value
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
			mockRetrieve: func(key, secret string) (string, error) {
				return "", nil // No error, but empty value
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
			mockRetrieve: func(key, secret string) (string, error) {
				return "", errors.New("param-store-failed")
			},
			mockGetenv: func(key string) string {
				return "" // Env var also not set
			},
			expectedValue:        "",
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
			mockRetrieve: func(key, secret string) (string, error) {
				return "", errors.New("param-store-mtls-failed")
			},
			mockGetenv: func(key string) string {
				return "" // Env var also not set
			},
			expectedValue:        "",
			expectPanic:          true,
			expectedPanicMessage: fmt.Sprintf("Failed to retrieve value for parameter store key '%s' (checked env var '%s'). Neither parameter store nor environment variable provided a value.", testKey, testEnvKey),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client if we need to test retrieval
			var mockClient ParameterStoreRetriever
			if tt.mockRetrieve != nil {
				mockClient = &MockClient{
					RetrieveFunc: tt.mockRetrieve,
				}
			}

			// Set the mock functions
			if tt.mockGetenv != nil {
				osGetenvFunc = tt.mockGetenv
			}

			// Set the pre-populated value if any
			config := tt.initialConfig
			config.ParameterStoreValue = tt.prePopulatedValue
			config.ParameterStoreUseEmptyValue = tt.useEmptyValue

			// Call Init and check for panic
			if tt.expectPanic {
				assert.PanicsWithError(t, tt.expectedPanicMessage, func() {
					config.Init(mockClient)
				})
			} else {
				assert.NotPanics(t, func() {
					config.Init(mockClient)
				})
				assert.Equal(t, tt.expectedValue, config.ParameterStoreValue)
			}
		})
	}
}

func TestParameterStoreConfig_Init_WithNilClient(t *testing.T) {
	setupTest()
	defer teardownTest()

	testKey := "test-key"
	testSecret := "test-secret"
	testEnvKey := "TEST_ENV_VAR"
	testEnvValue := "env-value"

	// Mock osGetenvFunc to return testEnvValue
	osGetenvFunc = func(key string) string {
		assert.Equal(t, testEnvKey, key)
		return testEnvValue
	}

	config := ParameterStoreConfig{
		ParameterStoreKey:      testKey,
		ParameterStoreSecret:   testSecret,
		EnvironmentVariableKey: testEnvKey,
	}

	// Call Init with nil client - should fall back to env var
	assert.NotPanics(t, func() {
		config.Init(nil)
	})
	assert.Equal(t, testEnvValue, config.ParameterStoreValue)
}