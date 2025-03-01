package commands

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"proxy-service/infrastructure/http/socks5"
	"time"
)

// StatusCommand checks service status by pinging the URL.
type StatusCommand struct {
	timeout    time.Duration          // timeout specifies the timeout duration for the HTTP request.
	pingUrl    string                 // pingUrl is the URL to be pinged to check the service status.
	socks5Pool *socks5.ConnectionPool // socks5Pool is the pool to obtain HTTP clients configured for SOCKS5.
	logger     *slog.Logger           // logger for structured logging.
}

// NewStatusCommand creates a new instance of StatusCommand.
func NewStatusCommand(
	timeout time.Duration,
	url string,
	pool *socks5.ConnectionPool,
	logger *slog.Logger,
) *StatusCommand {
	return &StatusCommand{
		timeout:    timeout,
		pingUrl:    url,
		socks5Pool: pool,
		logger:     logger,
	}
}

// Execute performs the status check by sending the HTTP request.
func (c *StatusCommand) Execute() (status string, err error) {
	var (
		httpClient *http.Client
		ctx        context.Context
		cancel     context.CancelFunc
		request    *http.Request
		response   *http.Response
		body       []byte
	)

	c.logger.Info("Initiating status check", "url", c.pingUrl, "timeout", c.timeout)
	ctx, cancel = context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	httpClient = c.socks5Pool.Borrow()
	defer c.socks5Pool.Return(httpClient)

	if request, err = http.NewRequestWithContext(ctx, http.MethodGet, c.pingUrl, http.NoBody); err != nil {
		c.logger.Error("Error creating HTTP request", "error", err)
		return "", fmt.Errorf("create request: %w", err)
	}
	c.logger.Debug("HTTP request created", "method", http.MethodGet, "url", c.pingUrl)

	if response, err = httpClient.Do(request); err != nil {
		c.logger.Error("Error executing HTTP request", "error", err)
		return "", fmt.Errorf("do request: %w", err)
	}
	c.logger.Info("HTTP response received", "statusCode", response.StatusCode)

	if body, err = io.ReadAll(response.Body); err != nil {
		c.logger.Error("Error reading HTTP response body", "error", err)
		return "", fmt.Errorf("read body: %w", err)
	}
	c.logger.Debug("HTTP response body read", "bodyLength", len(body))

	if err = response.Body.Close(); err != nil {
		c.logger.Error("Error closing HTTP response body", "error", err)
		return "", fmt.Errorf("close response body: %w", err)
	}

	c.logger.Info("Status check completed", "result", string(body))
	return string(body), nil
}
