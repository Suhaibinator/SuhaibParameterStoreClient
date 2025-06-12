# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go client library for interacting with a parameter store service via gRPC and REST APIs. It provides:

- **gRPC Client**: High-performance communication with parameter store service
- **REST Client**: HTTP-based API client for parameter store operations
- **Configuration Helper**: Smart configuration loading with parameter store priority and environment variable fallback
- **mTLS Support**: Built-in support for mutual TLS authentication

## Commands

### Build and Test
```bash
# Build the project
go build -v ./...

# Run all tests
go test -v ./...

# Run tests for specific package
go test -v ./client
go test -v ./config
```

### Dependencies
```bash
# Update dependencies
go mod tidy

# Download dependencies
go mod download
```

## Architecture

### Core Components

1. **client/** - Main client implementations
   - `client.go`: Main Client struct with functional options pattern
   - `grpc_client.go`: Low-level gRPC functions (GrpcSimpleStore, GrpcSimpleRetrieve)
   - `rest_client.go`: REST API client implementation

2. **config/** - Configuration management
   - `helper.go`: ParameterStoreConfig struct with Init() method
   - `interfaces.go`: ParameterStoreRetriever interface
   - Priority order: pre-set value → parameter store → environment variable → panic

3. **proto/** - Generated gRPC code
   - Generated from `parameter_store_interface.proto`

### Key Design Patterns

- **Functional Options**: Client configuration uses `WithTimeout()`, `WithMTLS()` options
- **Interface-based**: Config helper uses `ParameterStoreRetriever` interface for testability
- **Fail-fast**: Configuration loading panics if no value found (by design)
- **mTLS Support**: Automatic mTLS when certificates are provided via `CertificateSource`

### Client Usage Patterns

```go
// Basic client
client, err := client.NewClient("localhost", 50051)

// Client with options
client, err := client.NewClient("localhost", 50051,
    client.WithTimeout(10 * time.Second),
    client.WithMTLS(clientCert, clientKey, caCert),
)

// Configuration helper
config.APIKey.Init(psClient) // Retrieves from store or env var
```

## Testing

Tests use testify/mock for mocking dependencies. The config package uses function variables (`osGetenvFunc`) for mocking os.Getenv in tests.