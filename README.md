# SuhaibParameterStoreClient

A Go client library for interacting with a parameter store service via gRPC. It provides a clean, modern API to store and retrieve key-value pairs with optional TLS support including PKCS#12 certificates and configuration helpers.

[![Go CI](https://github.com/Suhaibinator/SuhaibParameterStoreClient/actions/workflows/go-ci.yml/badge.svg)](https://github.com/Suhaibinator/SuhaibParameterStoreClient/actions/workflows/go-ci.yml)

## Features

*   **Modern gRPC API:** High-performance communication with type safety
*   **TLS Support:** Flexible TLS configuration including separate cert/key files and PKCS#12 (.p12) files
*   **Secure Password Handling:** Built-in secure password callbacks for P12 decryption
*   **Configuration Helper:** Smart configuration loading with parameter store priority and environment variable fallback
*   **Functional Options Pattern:** Clean, extensible configuration API
*   **Context Support:** Full support for timeouts and cancellation
*   **Well Tested:** Comprehensive test coverage

## Installation

```bash
go get github.com/Suhaibinator/SuhaibParameterStoreClient
```

## Quick Start

### Basic Client Usage (No TLS)

```go
package main

import (
    "fmt"
    "log"

    "github.com/Suhaibinator/SuhaibParameterStoreClient/client"
)

func main() {
    // Create a client with default settings (no TLS)
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

### Client with TLS using Separate Cert/Key Files

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/Suhaibinator/SuhaibParameterStoreClient/client"
)

func main() {
    // Create TLS configuration from separate cert/key files
    tlsConfig := client.NewTLSConfigFromSeparateCerts(
        "/path/to/client.crt",
        "/path/to/client.key", 
        "/path/to/ca.crt",
    )

    // Create client with TLS
    psClient, err := client.NewClient("localhost", 50051,
        client.WithTimeout(10*time.Second),
        client.WithTLS(tlsConfig),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Use the client
    value, err := psClient.Retrieve("my-key", "my-secret")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Retrieved value: %s\n", value)
}
```

### Client with PKCS#12 Certificate

```go
package main

import (
    "fmt"
    "log"

    "github.com/Suhaibinator/SuhaibParameterStoreClient/client"
)

func main() {
    // Create TLS configuration from P12 file with terminal password prompt
    tlsConfig := client.NewTLSConfigFromP12File(
        "/path/to/cert.p12",
        client.DefaultPasswordCallbacks.TerminalPrompt,
    )

    // Create client with P12 TLS
    psClient, err := client.NewClient("localhost", 50051,
        client.WithTLS(tlsConfig),
    )
    if err != nil {
        log.Fatal(err)
    }

    value, err := psClient.Retrieve("my-key", "my-secret")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Retrieved value: %s\n", value)
}
```

### PKCS#12 with Environment Variable Password

```go
// Create P12 config with password from environment variable
tlsConfig := client.NewTLSConfigFromP12File(
    "/path/to/cert.p12",
    client.DefaultPasswordCallbacks.EnvVar("P12_PASSWORD"),
)

psClient, err := client.NewClient("localhost", 50051,
    client.WithTLS(tlsConfig),
)
```

### PKCS#12 with Custom Password Callback

```go
// Custom password callback
customPasswordFn := func() (string, error) {
    // Your custom logic here (e.g., read from secure vault)
    return getPasswordFromVault(), nil
}

tlsConfig := client.NewTLSConfigFromP12File(
    "/path/to/cert.p12", 
    customPasswordFn,
)
```

## Configuration Helper

The configuration helper provides automatic fallback from parameter store to environment variables:

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
        client.WithTimeout(3*time.Second),
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
    // This will try parameter store first, then fall back to env vars
    appConfig.APIKey.Init(psClient)
    appConfig.DBPassword.Init(psClient)

    // Use the values
    fmt.Printf("API Key: %s\n", appConfig.APIKey.ParameterStoreValue)
    fmt.Printf("DB Password: %s\n", appConfig.DBPassword.ParameterStoreValue)
}
```

## Advanced Usage

### Direct gRPC Function Usage

For lower-level control, you can use the direct gRPC functions:

```go
import (
    "context"
    "github.com/Suhaibinator/SuhaibParameterStoreClient/client"
)

// Direct gRPC retrieve (no TLS)
ctx := context.Background()
value, err := client.GrpcSimpleRetrieve(ctx, "localhost:50051", "password", "key")

// Direct gRPC store (no TLS)
err = client.GrpcSimpleStore(ctx, "localhost:50051", "password", "key", "value")

// Direct gRPC with TLS
tlsConfig := client.NewTLSConfigFromP12File("/path/to/cert.p12", passwordCallback)
value, err := client.GrpcSimpleRetrieveWithTLS(ctx, "localhost:50051", "password", "key", tlsConfig)
err = client.GrpcSimpleStoreWithTLS(ctx, "localhost:50051", "password", "key", "value", tlsConfig)
```

## TLS Configuration Options

### Separate Certificate Files
```go
tlsConfig := client.NewTLSConfigFromSeparateCerts(
    "/path/to/client.crt",
    "/path/to/client.key",
    "/path/to/ca.crt",
)
```

### Certificate Data in Memory
```go
tlsConfig := client.NewTLSConfigFromSeparateCertBytes(
    clientCertData, // []byte
    clientKeyData,  // []byte
    caCertData,     // []byte
)
```

### PKCS#12 File
```go
tlsConfig := client.NewTLSConfigFromP12File(
    "/path/to/cert.p12",
    client.DefaultPasswordCallbacks.TerminalPrompt,
)
```

### PKCS#12 Data in Memory
```go
tlsConfig := client.NewTLSConfigFromP12Bytes(
    p12Data,     // []byte
    passwordFn,  // func() (string, error)
)
```

## Password Callback Options

### Built-in Callbacks
```go
// Terminal prompt (secure input)
client.DefaultPasswordCallbacks.TerminalPrompt

// Environment variable
client.DefaultPasswordCallbacks.EnvVar("P12_PASSWORD")

// Static password (not recommended for production)
client.DefaultPasswordCallbacks.Static("my-password")
```

### Custom Callback
```go
customCallback := func() (string, error) {
    // Your custom secure password retrieval logic
    return getPasswordSecurely(), nil
}
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