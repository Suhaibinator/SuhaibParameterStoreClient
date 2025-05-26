# SuhaibParameterStoreClient

A Go client library for interacting with a parameter store service via gRPC. It provides a clean, intuitive API to store and retrieve key-value pairs with optional mTLS support and configuration helpers.

[![Go CI](https://github.com/Suhaibinator/SuhaibParameterStoreClient/actions/workflows/go-ci.yml/badge.svg?event=workflow_run)](https://github.com/Suhaibinator/SuhaibParameterStoreClient/actions/workflows/go-ci.yml)

## Features

*   **Simple API:** Intuitive client with direct `Retrieve()` and `Store()` methods
*   **gRPC Support:** High-performance communication with parameter store service
*   **mTLS Support:** Built-in support for mutual TLS authentication
*   **Configuration Helper:** Smart configuration loading with parameter store priority and environment variable fallback
*   **Flexible Options:** Use functional options pattern for clean configuration
*   **Context Support:** Full support for timeouts and cancellation
*   **Well Tested:** Comprehensive test coverage with mocks

## Installation

```bash
go get github.com/Suhaibinator/SuhaibParameterStoreClient
```

## Quick Start

### Basic Client Usage

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/Suhaibinator/SuhaibParameterStoreClient/client"
)

func main() {
    // Create a client with default settings
    psClient, err := client.NewClient("localhost", 50051)
    if err != nil {
        log.Fatal(err)
    }

    // Store a value
    err = psClient.Store("my-key", "my-secret", "my-value")
    if err != nil {
        log.Fatal(err)
    }

    // Retrieve a value
    value, err := psClient.Retrieve("my-key", "my-secret")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Retrieved value: %s\n", value)
}
```

### Client with Options

```go
// Create a client with custom timeout
psClient, err := client.NewClient("localhost", 50051,
    client.WithTimeout(10 * time.Second),
)

// Create a client with mTLS
psClient, err := client.NewClient("localhost", 50051,
    client.WithMTLS(
        client.CertificateSource{FilePath: "/path/to/client.crt"},
        client.CertificateSource{FilePath: "/path/to/client.key"},
        client.CertificateSource{FilePath: "/path/to/ca.crt"},
    ),
)

// You can also use certificate bytes directly
psClient, err := client.NewClient("localhost", 50051,
    client.WithMTLS(
        client.CertificateSource{Bytes: clientCertBytes},
        client.CertificateSource{Bytes: clientKeyBytes},
        client.CertificateSource{Bytes: caCertBytes},
    ),
)
```

## Configuration Helper

The configuration helper provides a convenient way to load configuration values with automatic fallback:

1. First, it checks if a value is already set
2. Then tries to retrieve from the parameter store
3. Falls back to environment variables if parameter store fails
4. Panics if no value is found (fail-fast approach)

### Basic Configuration Usage

```go
package main

import (
    "fmt"
    "log"
    "os"
    "time"

    "github.com/Suhaibinator/SuhaibParameterStoreClient/client"
    "github.com/Suhaibinator/SuhaibParameterStoreClient/config"
)

func main() {
    // Set up a fallback environment variable
    os.Setenv("MY_APP_API_KEY", "env-api-key-123")
    
    // Create the parameter store client
    psClient, err := client.NewClient("localhost", 50051,
        client.WithTimeout(3 * time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Define your configuration
    appConfig := struct {
        APIKey config.ParameterStoreConfig
        DBPassword config.ParameterStoreConfig
    }{
        APIKey: config.ParameterStoreConfig{
            ParameterStoreKey:      "my-app/api-key",
            ParameterStoreSecret:   "secret-password",
            EnvironmentVariableKey: "MY_APP_API_KEY",
        },
        DBPassword: config.ParameterStoreConfig{
            ParameterStoreKey:      "my-app/db-password",
            ParameterStoreSecret:   "secret-password",
            EnvironmentVariableKey: "MY_APP_DB_PASSWORD",
        },
    }

    // Initialize configuration values
    appConfig.APIKey.Init(psClient)
    appConfig.DBPassword.Init(psClient)

    // Use the values
    fmt.Printf("API Key: %s\n", appConfig.APIKey.ParameterStoreValue)
    fmt.Printf("DB Password: %s\n", appConfig.DBPassword.ParameterStoreValue)
}
```

### Configuration with mTLS

```go
// Create client with mTLS
psClient, err := client.NewClient("localhost", 50051,
    client.WithTimeout(3 * time.Second),
    client.WithMTLS(
        client.CertificateSource{FilePath: "/path/to/client.crt"},
        client.CertificateSource{FilePath: "/path/to/client.key"},
        client.CertificateSource{FilePath: "/path/to/ca.crt"},
    ),
)

// Use it with configuration helper - mTLS is handled automatically
appConfig.SecureAPIKey.Init(psClient)
```

## Direct gRPC Function Usage

For advanced use cases, you can use the lower-level gRPC functions directly:

```go
import (
    "context"
    "github.com/Suhaibinator/SuhaibParameterStoreClient/client"
)

// Direct gRPC retrieve
ctx := context.Background()
value, err := client.GrpcSimpleRetrieve(ctx, "localhost:50051", "password", "key")

// Direct gRPC store
err = client.GrpcSimpleStore(ctx, "localhost:50051", "password", "key", "value")
```

## REST API Client

The library also includes a REST API client:

```go
// Create REST API client
apiClient := client.NewAPIClient("https://your-server.com/api", "your-password")

// Store a value
err := apiClient.Store("key", "value")

// Retrieve a value
value, err := apiClient.Retrieve("key")

// Create REST API client with mTLS
apiClient, err := client.NewAPIClientWithMTLS(
    "https://your-server.com/api",
    "your-password",
    "/path/to/client.crt",
    "/path/to/client.key",
    "/path/to/ca.crt",
)
```

## Testing

To run the tests:

```bash
go test ./...
```

For verbose output:

```bash
go test -v ./...
```

## License

This project is licensed under the terms of the [LICENSE](LICENSE) file.