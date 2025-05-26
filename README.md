# SuhaibParameterStoreClient

A Go client library for interacting with a parameter store service via gRPC. It provides functions to store and retrieve key-value pairs and includes configuration helpers to fetch values, prioritizing the parameter store over environment variables.

[![Go CI](https://github.com/Suhaibinator/SuhaibParameterStoreClient/actions/workflows/go-ci.yml/badge.svg?event=workflow_run)](https://github.com/Suhaibinator/SuhaibParameterStoreClient/actions/workflows/go-ci.yml)

## Features

*   **gRPC Client:** Functions (`GrpcSimpleStore`, `GrpcSimpleRetrieve`) to interact with a gRPC-based parameter store service.
*   **Configuration Helper:** A `ParameterStoreClient` for connection details and a `ParameterStoreConfig` struct with an `Init` method to retrieve configuration values, checking the parameter store first and falling back to environment variables.
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
    retrievedValue, err := client.GrpcSimpleRetrieve(ctx, serverAddr, password, key)
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

        client := &config.ParameterStoreClient{
                Host:    paramStoreHost,
                Port:    paramStorePort,
                Timeout: paramStoreTimeout,
        }

	// Define the configuration structure
	appConfig := struct {
		APIKey config.ParameterStoreConfig
		// Add other config fields as needed
	}{
		APIKey: config.ParameterStoreConfig{
			ParameterStoreKey:     "my-app/api-key",       // Key in the parameter store
			ParameterStoreSecret:  "your-secret-password", // Password for the store
			EnvironmentVariableKey: "MY_APP_API_KEY_ENV",   // Fallback environment variable
			// ParameterStoreValue can be pre-filled to skip retrieval
		},
	}

        // Initialize the APIKey field using the client
        // This will try the parameter store first, then the environment variable.
        // It will panic if neither provides a value.
        appConfig.APIKey.Init(client)

	fmt.Printf("Initialized API Key: %s\n", appConfig.APIKey.ParameterStoreValue)

	// Use the initialized value
	// ... your application logic using appConfig.APIKey.ParameterStoreValue ...
}

```

## Using mTLS for Secure Communication

This client library supports Mutual TLS (mTLS) for establishing secure, authenticated connections to both gRPC and REST-based parameter store services. When mTLS is configured, the client and server authenticate each other using X.509 certificates.

### gRPC Client with mTLS

To communicate with a gRPC server requiring mTLS, use the dedicated `GrpcSimpleRetrieveWithMTLS` and `GrpcSimpleStoreWithMTLS` functions. These functions require paths to the client's certificate, the client's private key, and the CA certificate that signed the server's certificate.

**Functions:**

*   `client.GrpcSimpleRetrieveWithMTLS(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, clientCertFile string, clientKeyFile string, caCertFile string, opts ...grpc.DialOption) (val string, err error)`
*   `client.GrpcSimpleStoreWithMTLS(ctx context.Context, ServerAddress string, AuthenticationPassword string, key string, value string, clientCertFile string, clientKeyFile string, caCertFile string, opts ...grpc.DialOption) (err error)`

**Parameters for mTLS:**

*   `clientCertFile (string)`: Path to the client's PEM-encoded certificate file.
*   `clientKeyFile (string)`: Path to the client's PEM-encoded private key file.
*   `caCertFile (string)`: Path to the PEM-encoded CA certificate file for verifying the server.

**Example Snippet:**

```go
// Assuming clientCertPath, clientKeyPath, caCertPath are defined
retrievedValue, err := client.GrpcSimpleRetrieveWithMTLS(ctx, serverAddr, password, key, clientCertPath, clientKeyPath, caCertPath)
if err != nil {
    log.Fatalf("Failed to retrieve value with mTLS: %v", err)
}
// ...
err = client.GrpcSimpleStoreWithMTLS(ctx, serverAddr, password, key, value, clientCertPath, clientKeyPath, caCertPath)
if err != nil {
    log.Fatalf("Failed to store value with mTLS: %v", err)
}
```

### REST Client with mTLS

For RESTful communication with a server requiring mTLS, use the `NewAPIClientWithMTLS` function to create an `APIClient` instance. This client will be pre-configured with an HTTP transport that handles mTLS.

**Function:**

*   `client.NewAPIClientWithMTLS(baseURL, authenticationPassword, clientCertFile, clientKeyFile, caCertFile string) (*APIClient, error)`

**Parameters for mTLS:**

*   `clientCertFile (string)`: Path to the client's PEM-encoded certificate file.
*   `clientKeyFile (string)`: Path to the client's PEM-encoded private key file.
*   `caCertFile (string)`: Path to the PEM-encoded CA certificate file for verifying the server.

**Example Snippet:**

```go
// Assuming clientCertPath, clientKeyPath, caCertPath are defined
apiClient, err := client.NewAPIClientWithMTLS("https://your-server.com/api", password, clientCertPath, clientKeyPath, caCertPath)
if err != nil {
    log.Fatalf("Failed to create mTLS API client: %v", err)
}
// Use apiClient to .Store() or .Retrieve()
// value, err := apiClient.Retrieve("some-key")
```

### Configuration Helper with mTLS

To use mTLS, populate the certificate paths on `ParameterStoreClient`. When these fields are set, `Init` will connect to the parameter store using mTLS.

```go
client := &config.ParameterStoreClient{
    Host:          paramStoreHost,
    Port:          paramStorePort,
    Timeout:       paramStoreTimeout,
    ClientCertFile: "/path/to/client.crt",
    ClientKeyFile:  "/path/to/client.key",
    CACertFile:     "/path/to/ca.crt",
}

appConfig := struct {
    SecureSetting config.ParameterStoreConfig
}{
    SecureSetting: config.ParameterStoreConfig{
        ParameterStoreKey:     "my-app/secure-setting",
        ParameterStoreSecret:  "your-secret-password",
        EnvironmentVariableKey: "MY_APP_SECURE_SETTING_ENV",
    },
}

// Init will use gRPC with mTLS because the client has certificate paths
appConfig.SecureSetting.Init(client)

fmt.Printf("Initialized Secure Setting: %s\n", appConfig.SecureSetting.ParameterStoreValue)
```

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
