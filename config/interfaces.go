package config

// ParameterStoreRetriever is an interface for retrieving values from a parameter store
type ParameterStoreRetriever interface {
	Retrieve(key, secret string) (string, error)
}