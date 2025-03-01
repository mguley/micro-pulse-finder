package socks5

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"proxy-service/domain/entities"
	"proxy-service/domain/interfaces"
	"time"

	"golang.org/x/net/proxy"
)

// Client provides functionality to interact with a SOCKS5 proxy.
type Client struct {
	userAgent interfaces.Agent // userAgent is responsible for generating User-Agent headers.
	timeout   time.Duration    // timeout specifies the timeout duration for the HTTP client.
	network   string           // network specifies the network type (e.g., "tcp").
	logger    *slog.Logger
}

// NewClient creates a new instance of Client.
func NewClient(userAgent interfaces.Agent, timeout time.Duration, logger *slog.Logger) *Client {
	return &Client{
		userAgent: userAgent,
		timeout:   timeout,
		network:   "tcp",
		logger:    logger,
	}
}

// Create initializes HTTP client configured to route traffic through a SOCKS5 proxy with authentication.
func (c *Client) Create() (client *http.Client, err error) {
	var (
		address string
		dialer  proxy.Dialer
		auth    = c.generateCredentials()
	)

	c.logger.Info("Creating SOCKS5 HTTP client")

	// Retrieve the SOCKS5 proxy address from the configuration.
	if address, err = entities.GetProxy().Address(); err != nil {
		c.logger.Error("Could not get proxy address", "error", err)
		return nil, fmt.Errorf("could not get proxy address: %w", err)
	}
	c.logger.Debug("Obtained proxy address", "address", address)

	// Create a SOCKS5 dialer using the proxy address.
	if dialer, err = proxy.SOCKS5(c.network, address, auth, proxy.Direct); err != nil {
		c.logger.Error("Could not create SOCKS5 dialer", "error", err)
		return nil, fmt.Errorf("could not create socks5 dialer: %w", err)
	}

	// Define a context-aware dialer that uses the SOCKS5 proxy.
	dialContext := func(ctx context.Context, network, address string) (conn net.Conn, err error) {
		return dialer.Dial(network, address)
	}

	// Create an HTTP client with custom transport that supports the User-Agent and SOCKS5 proxy.
	client = &http.Client{
		Transport: &RoundTripWithUserAgent{
			roundTripper: &http.Transport{DialContext: dialContext},
			agent:        c.userAgent.Generate(),
			logger:       c.logger,
		},
		Timeout: c.timeout,
	}

	c.logger.Info("Successfully created SOCKS5 HTTP client")
	return client, nil
}

// generateCredentials creates a random username and password for authentication.
func (c *Client) generateCredentials() *proxy.Auth {
	username := make([]byte, 8)
	password := make([]byte, 16)

	_, _ = rand.Read(username)
	_, _ = rand.Read(password)

	c.logger.Debug("Generated proxy credentials", "username", hex.EncodeToString(username))
	c.logger.Debug("Generated proxy credentials", "password", hex.EncodeToString(password))
	return &proxy.Auth{
		User:     hex.EncodeToString(username),
		Password: hex.EncodeToString(password),
	}
}
