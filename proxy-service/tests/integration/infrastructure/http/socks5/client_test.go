package socks5

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_HTTPRequest tests if the SOCKS5 client can successfully make an HTTP request.
func TestClient_HTTPRequest(t *testing.T) {
	container := SetupTestContainer()
	client := container.Socks5Client.Get()

	httpClient, err := client.Create()
	require.NoError(t, err, "Failed to create HTTP client")

	// Perform the request
	url := "https://httpbin.org/get"
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	require.NoError(t, err, "Failed to create HTTP request")
	response, err := httpClient.Do(request)
	require.NoError(t, err, "HTTP request through SOCKS5 proxy failed")
	require.NotNil(t, response, "Response should not be nil")

	defer func() {
		if err = response.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Read response
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err, "Failed to read response body")
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "200 OK", response.Status)
	require.NotEmpty(t, body, "Response body should not be empty")
}

// TestClient_UserAgent verifies that the generated User-Agent is being used.
func TestClient_UserAgent(t *testing.T) {
	container := SetupTestContainer()
	client := container.Socks5Client.Get()

	httpClient, err := client.Create()
	require.NoError(t, err, "Failed to create HTTP client")

	url := "https://httpbin.org/user-agent"
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	require.NoError(t, err, "Failed to create HTTP request")

	// Perform the request.
	response, err := httpClient.Do(request)
	require.NoError(t, err, "HTTP request through SOCKS5 proxy failed")
	require.NotNil(t, response, "Response should not be nil")

	defer func() {
		if err = response.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Read response
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err, "Failed to read response body")
	assert.Contains(t, string(body), "Mozilla/5.0", "Expected User-Agent to contain Mozilla/5.0")
}

// TestClient_CircuitRotation ensures that each new client instance results in a different exit IP.
func TestClient_CircuitRotation(t *testing.T) {
	// ipResponse represents the JSON response from httpbin.org/ip
	type ipResponse struct {
		Origin string `json:"origin"`
	}

	container := SetupTestContainer()
	url := "https://httpbin.org/ip"
	ipAddresses := make(map[string]bool)
	var collectedIPs []string

	// Create three separate client instances and fetch their IP addresses.
	for i := 0; i < 3; i++ {
		client := container.Socks5Client.Get()
		httpClient, err := client.Create()
		require.NoError(t, err, "Failed to create HTTP client")

		request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
		require.NoError(t, err, "Failed to create HTTP request")

		response, err := httpClient.Do(request)
		require.NoError(t, err, "HTTP request through SOCKS5 proxy failed")
		require.NotNil(t, response, "Response should not be nil")

		// Read response
		body, err := io.ReadAll(response.Body)
		require.NoError(t, err, "Failed to read response body")

		err = response.Body.Close()
		require.NoError(t, err, "Failed to close response body")

		// Parse JSON response
		var ipData ipResponse
		err = json.Unmarshal(body, &ipData)
		require.NoError(t, err, "Failed to parse JSON response")

		require.NotEmpty(t, ipData.Origin, "IP address should not be empty")
		ipAddresses[ipData.Origin] = true
		collectedIPs = append(collectedIPs, ipData.Origin)
	}

	// Ensure all fetched IPs are unique
	assert.Equal(t, 3, len(ipAddresses), "Expected each client to have a unique IP address")
	// Debug output: print collected IP addresses
	t.Logf("Collected IP addresses: %v", collectedIPs)
}
