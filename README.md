# SuhaibParameterStoreClient

A Go client library for interacting with a parameter store service via gRPC. It provides functions to store and retrieve key-value pairs and includes configuration helpers to fetch values, prioritizing the parameter store over environment variables.

[![Go CI](https://github.com/Suhaibinator/SuhaibParameterStoreClient/actions/workflows/go-ci.yml/badge.svg?event=workflow_run)](https://github.com/Suhaibinator/SuhaibParameterStoreClient/actions/workflows/go-ci.yml)

## Features

*   **gRPC Client:** Functions (`GrpcSimpleStore`, `GrpcimpleRetrieve`) to interact with a gRPC-based parameter store service.
*   **Configuration Helper:** A `ParameterStoreConfig` struct and `Init` method to easily retrieve configuration values, checking the parameter store first and falling back to environment variables.
*   **Context Handling:** gRPC client functions utilize `context.Context` for timeouts and cancellation.
*   **Testable:** Includes mocks and tests for both the gRPC client and the configuration helper.

## Installation

```bash
go get github.com/Suhaibinator/SuhaibParameterStoreClient
```

*(Note: Replace the import path if your repository location is different)*

## Usage

### Direct gRPC Client Usage

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Suhaibinator/SuhaibParameterStoreClient/client"
)

func main() {
	serverAddr := "localhost:50051" // Replace with your gRPC server address
	password := "your-secret-password"
	key := "my-config-key"
	value := "my-config-value"

	// Use a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Store a value
	err := client.GrpcSimpleStore(ctx, serverAddr, password, key, value)
	if err != nil {
		log.Fatalf("Failed to store value: %v", err)
	}
	fmt.Printf("Successfully stored value for key '%s'\n", key)

	// Retrieve a value
	retrievedValue, err := client.GrpcimpleRetrieve(ctx, serverAddr, password, key)
	if err != nil {
		log.Fatalf("Failed to retrieve value: %v", err)
	}
	fmt.Printf("Successfully retrieved value for key '%s': %s\n", key, retrievedValue)
}
```

### Configuration Helper Usage

```go
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Suhaibinator/SuhaibParameterStoreClient/config"
)

func main() {
	// Example: Set an environment variable as a fallback
	os.Setenv("MY_APP_API_KEY_ENV", "env-api-key-123")
	defer os.Unsetenv("MY_APP_API_KEY_ENV") // Clean up env var

	paramStoreHost := "localhost" // Replace with your parameter store host
	paramStorePort := 50051       // Replace with your parameter store port
	paramStoreTimeout := 3 * time.Second

	// Define the configuration structure
	appConfig := struct {
		APIKey config.ParameterStoreConfig
		// Add other config fields as needed
	}{
		APIKey: config.ParameterStoreConfig{
			ParameterStoreKey:     "my-app/api-key",       // Key in the parameter store
			ParameterStoreSecret:  "your-secret-password", // Password for the store
			EnvirnmentVariableKey: "MY_APP_API_KEY_ENV",   // Fallback environment variable
			// ParameterStoreValue can be pre-filled to skip retrieval
		},
	}

	// Initialize the APIKey field
	// This will try the parameter store first, then the environment variable.
	// It will panic if neither provides a value.
	appConfig.APIKey.Init(paramStoreHost, paramStorePort, paramStoreTimeout)

	fmt.Printf("Initialized API Key: %s\n", appConfig.APIKey.ParameterStoreValue)

	// Use the initialized value
	// ... your application logic using appConfig.APIKey.ParameterStoreValue ...
}

```
*(Note the intentional typo `EnvirnmentVariableKey` in the struct field tag as required by the original code).*

## Testing

To run the tests for this library:

```bash
go test ./...
```

This command will execute all tests in the `client` and `config` packages.

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues.

## License

This project is licensed under the terms of the [LICENSE](LICENSE) file.
