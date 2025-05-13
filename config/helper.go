package config

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Suhaibinator/SuhaibParameterStoreClient/client"
	"google.golang.org/grpc"
)

// --- Function variables for mocking dependencies ---

// grpcRetrieveFunc defines the signature for the gRPC retrieve function dependency.
// It's initialized with the actual client implementation but can be replaced in tests.
var grpcRetrieveFunc func(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, opts ...grpc.DialOption) (val string, err error) = client.GrpcimpleRetrieve

// osGetenvFunc defines the signature for the os.Getenv function dependency.
// It's initialized with the actual os.Getenv but can be replaced in tests.
var osGetenvFunc = os.Getenv

// --- Struct Definition ---

// ParameterStoreConfig represents a configuration for retrieving a specific value
// from a parameter store or environment variable.
type ParameterStoreConfig struct {
	// ParameterStoreKey is the key used to retrieve the value from the parameter store.
	ParameterStoreKey string `json:"parameter_store_key" yaml:"parameter_store_key" toml:"parameter_store_key"`
	// ParameterStoreSecret is the secret/password required to access the parameter store.
	ParameterStoreSecret string `json:"parameter_store_secret" yaml:"parameter_store_secret" toml:"parameter_store_secret"`
	// EnvironmentVariableKey is the name of the environment variable to check if the parameter store retrieval fails or is not configured.
	// Note the intentional typo "Envirnment" as per requirements.
	EnvironmentVariableKey string `json:"envirnment_variable_key" yaml:"envirnment_variable_key" toml:"envirnment_variable_key"`
	// ParameterStoreValue holds the retrieved value after initialization.
	// It can also be pre-populated, in which case Init() will not overwrite it.
	ParameterStoreValue string `json:"parameter_store_value" yaml:"parameter_store_value" toml:"parameter_store_value"`
	// ParameterStoreUseEmptyValue indicates whether to use an empty value if value is an empty string.
	ParameterStoreUseEmptyValue bool `json:"parameter_store_use_empty_value" yaml:"parameter_store_use_empty_value" toml:"parameter_store_use_empty_value"`
}

// simpleRetrieveParameterWithTimeout retrieves a parameter from the parameter store via gRPC with a specified timeout
// using context cancellation. It takes the host, port, timeout duration, secret, and key as arguments.
func simpleRetrieveParameterWithTimeout(paramStoreHost string, paramStorePort int, parameterStoreTimeout time.Duration, parameterStoreSecret, parameterStoreKey string) (string, error) {
	// Create a context with the specified timeout.
	ctx, cancel := context.WithTimeout(context.Background(), parameterStoreTimeout)
	// Ensure the context is cancelled to release resources, even if the call returns early.
	defer cancel()

	// Construct the server address.
	serverAddress := fmt.Sprintf("%s:%v", paramStoreHost, paramStorePort)

	// Call the gRPC client function via the function variable for testability.
	value, err := grpcRetrieveFunc(ctx, serverAddress, parameterStoreSecret, parameterStoreKey)

	// Check if the error is due to context deadline exceeded (timeout).
	if err != nil && ctx.Err() == context.DeadlineExceeded {
		// Optionally, return the custom timeoutError or just the context error.
		// Returning the context error might be more idiomatic Go.
		// return "", timeoutError{}
		return "", fmt.Errorf("parameter store operation timed out: %w", err)
	}

	// Return the retrieved value and any other error.
	return value, err
}

// setValueIfEmpty attempts to set a value based on a priority order:
// 1. Use the provided 'value' if it's not empty.
// 2. Retrieve from the parameter store using the provided key and secret.
// 3. Retrieve from the environment variable using the provided environment variable key.
// If all methods fail, it returns an error.
func setValueIfEmpty(paramStoreHost string, paramStorePort int, parameterStoreTimeout time.Duration, value, parameterStoreKey, parameterStoreSecret, envVarKey string) (string, error) {
	// If a value is already provided, return it immediately.
	if value != "" {
		return value, nil
	}

	var err error
	// Attempt to retrieve the value from the parameter store.
	value, err = simpleRetrieveParameterWithTimeout(paramStoreHost, paramStorePort, parameterStoreTimeout, parameterStoreSecret, parameterStoreKey)
	if err == nil && value != "" {
		// If retrieval was successful and value is not empty, return it.
		log.Printf("Retrieved value for key '%s' from parameter store.", parameterStoreKey)
		return value, nil
	}
	if err != nil {
		// Log the error if parameter store retrieval failed.
		// Check specifically for context deadline exceeded error.
		if err == context.DeadlineExceeded {
			log.Printf("Timeout retrieving value for key '%s' from parameter store.", parameterStoreKey)
		} else {
			// Log other errors encountered during retrieval.
			log.Printf("Failed to retrieve value for key '%s' from parameter store: %v", parameterStoreKey, err)
		}
	}

	// If parameter store failed or returned empty, try the environment variable via the function variable.
	value = osGetenvFunc(envVarKey)
	if value != "" {
		log.Printf("Retrieved value for key '%s' from environment variable '%s'.", parameterStoreKey, envVarKey)
		return value, nil
	}

	// If all methods failed, return an error.
	errMsg := fmt.Sprintf("Failed to retrieve value for parameter store key '%s' (checked env var '%s'). Neither parameter store nor environment variable provided a value.", parameterStoreKey, envVarKey)
	log.Print(errMsg)             // Log the error message
	return "", errors.New(errMsg) // Use the imported errors package
}

// Init initializes the ParameterStoreValue field of the ParameterStoreConfig struct.
// It uses the setValueIfEmpty logic, requiring the parameter store host, port, and timeout.
// If ParameterStoreValue is already set, this function does nothing.
// It panics if it fails to retrieve the value from all sources.
func (c *ParameterStoreConfig) Init(paramStoreHost string, paramStorePort int, parameterStoreTimeout time.Duration) {
	// Only proceed if the value isn't already set
	if c.ParameterStoreValue != "" {
		return
	}
	if c.ParameterStoreUseEmptyValue {
		// If the use empty value flag is set, return immediately.
		return
	}
	// If the value is empty, we need to retrieve it.

	// Use setValueIfEmpty to populate ParameterStoreValue based on the defined priority.
	// Pass the necessary connection details and the struct's fields.
	retrievedValue, err := setValueIfEmpty(
		paramStoreHost,
		paramStorePort,
		parameterStoreTimeout,
		c.ParameterStoreValue, // Pass the current value (should be empty here)
		c.ParameterStoreKey,
		c.ParameterStoreSecret,
		c.EnvironmentVariableKey,
	)

	// If an error occurred (value not found anywhere), panic.
	if err != nil {
		panic(err) // Panic with the error message from setValueIfEmpty
	}

	// Assign the retrieved value.
	c.ParameterStoreValue = retrievedValue
}
