package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateCertPair(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	assert.NoError(t, err)
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return
}

func createDummyCertFilesForREST(t *testing.T, makeCACertInvalid bool) (certFile, keyFile, caFile string, cleanup func()) {
	t.Helper()
	certPEM, keyPEM := generateCertPair(t)
	caPEM := certPEM
	if makeCACertInvalid {
		caPEM = []byte("invalid")
	}
	dir, err := os.MkdirTemp("", "rest-mtls-test-certs")
	assert.NoError(t, err)
	certFile = filepath.Join(dir, "client.crt")
	keyFile = filepath.Join(dir, "client.key")
	caFile = filepath.Join(dir, "ca.crt")
	assert.NoError(t, os.WriteFile(certFile, certPEM, 0600))
	assert.NoError(t, os.WriteFile(keyFile, keyPEM, 0600))
	assert.NoError(t, os.WriteFile(caFile, caPEM, 0600))
	cleanup = func() { os.RemoveAll(dir) }
	return
}

func TestNewAPIClientWithMTLS_Success(t *testing.T) {
	cert, key, ca, cleanup := createDummyCertFilesForREST(t, false)
	defer cleanup()
	client, err := NewAPIClientWithMTLS("https://example.com", "secret", cert, key, ca)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	tlsConf, ok := client.httpClient.Transport.(*http.Transport)
	assert.True(t, ok)
	_ = tlsConf
}

func TestNewAPIClientWithMTLS_InvalidCA(t *testing.T) {
	cert, key, ca, cleanup := createDummyCertFilesForREST(t, true)
	defer cleanup()
	_, err := NewAPIClientWithMTLS("https://example.com", "secret", cert, key, ca)
	assert.Error(t, err)
}
