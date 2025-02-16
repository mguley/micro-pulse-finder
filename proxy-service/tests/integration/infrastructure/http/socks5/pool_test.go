package socks5

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const urlGet = "https://httpbin.org/get"

// TestConnectionPool_BorrowReturn verifies that a connection can be borrowed from the pool,
// used to make an HTTP request, and then returned to the pool for reuse.
func TestConnectionPool_BorrowReturn(t *testing.T) {
	container := SetupTestContainer()
	pool := container.ConnectionPool.Get()

	// Borrow a connection from the pool.
	client := pool.Borrow()
	require.NotNil(t, client, "Borrowed client should not be nil")

	// Use the borrowed client to make a simple HTTP request.
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, urlGet, http.NoBody)
	require.NoError(t, err, "Failed to create HTTP request")

	response, err := client.Do(request)
	require.NoError(t, err, "HTTP request through borrowed connection failed")
	require.Equal(t, http.StatusOK, response.StatusCode, "Expected HTTP status 200")
	require.NotNil(t, response, "Response should not be nil")

	err = response.Body.Close()
	require.NoError(t, err, "Failed to close response body")

	// Return the client to the pool.
	pool.Return(client)
	pool.Shutdown()
}

// TestConnectionPool_ConcurrentBorrow tests concurrent borrowing and returning of connections
// from the pool by spawning multiple goroutines that each performs an HTTP GET request.
func TestConnectionPool_ConcurrentBorrow(t *testing.T) {
	container := SetupTestContainer()
	pool := container.ConnectionPool.Get()

	type ipResponse struct {
		Origin string `json:"origin"`
	}

	var (
		mu           sync.Mutex
		workers      = 7
		urlIp        = "https://httpbin.org/ip"
		collectedIPs []string
		errorsChan   = make(chan error, workers)
		wg           sync.WaitGroup
	)

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errorsChan <- fmt.Errorf("panic: %v", r)
				}
			}()

			// Workload
			client := pool.Borrow()
			defer pool.Return(client)
			require.NotNil(t, client, "Borrowed client should not be nil")

			request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, urlIp, http.NoBody)
			require.NoError(t, err, "Failed to create HTTP request")

			response, err := client.Do(request)
			require.NoError(t, err, "HTTP request through borrowed connection failed")
			require.Equal(t, http.StatusOK, response.StatusCode, "Expected HTTP status 200")
			require.NotNil(t, response, "Response should not be nil")

			// Read response
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err, "Failed to read response body")

			// Parse JSON response
			var ipData ipResponse
			err = json.Unmarshal(body, &ipData)
			require.NoError(t, err, "Failed to parse JSON response")
			require.NotEmpty(t, ipData.Origin, "IP address should not be empty")

			mu.Lock()
			collectedIPs = append(collectedIPs, ipData.Origin)
			mu.Unlock()

			err = response.Body.Close()
			require.NoError(t, err, "Failed to close response body")
		}()
	}

	wg.Wait()
	close(errorsChan)
	pool.Shutdown()

	for err := range errorsChan {
		require.NoError(t, err, "Error during concurrent borrow and return")
	}

	require.Equal(t, workers, len(collectedIPs), "Expected each client to have a unique IP address")
	t.Logf("Collected IP addresses: %v", collectedIPs)
}

// TestConnectionPool_Shutdown ensures that after Shutdown, the pool stops its refresh routine
// and no further connections can be borrowed from the pool.
func TestConnectionPool_Shutdown(t *testing.T) {
	container := SetupTestContainer()
	pool := container.ConnectionPool.Get()

	// Borrow one connection to simulate usage.
	client := pool.Borrow()
	require.NotNil(t, client, "Borrowed client should not be nil")

	// Return the connection.
	pool.Return(client)

	// Shutdown the pool.
	pool.Shutdown()

	// After shutdown, attempting to borrow should return nil since the channel is closed.
	client = pool.Borrow()
	require.Nil(t, client, "Expected borrowed client to be nil after shutdown")

	// Intentionally called to make sure that Shutdown is executed only once.
	pool.Shutdown()
}

// TestConnectionPool_Refresh verifies that the pool refreshes idle connections periodically.
func TestConnectionPool_Refresh(t *testing.T) {
	container := SetupTestContainer()
	pool := container.ConnectionPool.Get()

	// Borrow and return a connection to ensure it's part of the idle pool.
	client := pool.Borrow()
	require.NotNil(t, client, "Borrowed client should not be nil")
	pool.Return(client)

	// Wait for a duration longer than the refresh interval.
	config := container.Config.Get()
	refreshInterval := time.Duration(config.Pool.RefreshInterval) * time.Second
	refreshDuration := refreshInterval + time.Duration(2)*time.Second
	t.Logf("Refresh interval: %v, sleeping: %v", refreshInterval, refreshDuration)
	time.Sleep(refreshDuration)

	// Borrow a connection and make sure it still works.
	refreshedClient := pool.Borrow()
	require.NotNil(t, refreshedClient, "Borrowed client after refresh should not be nil")

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, urlGet, http.NoBody)
	require.NoError(t, err, "Failed to create HTTP request")

	response, err := refreshedClient.Do(request)
	require.NoError(t, err, "HTTP request through refreshed connection failed")
	require.Equal(t, http.StatusOK, response.StatusCode, "Expected HTTP status 200")
	require.NotNil(t, response, "Response should not be nil")

	err = response.Body.Close()
	require.NoError(t, err, "Failed to close response body")

	pool.Return(refreshedClient)
	pool.Shutdown()
}
