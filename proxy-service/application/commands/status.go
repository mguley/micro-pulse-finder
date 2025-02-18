package commands

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"proxy-service/infrastructure/http/socks5"
	"time"
)

// StatusCommand checks service status by pinging the URL.
type StatusCommand struct {
	timeout    time.Duration          // timeout specifies the timeout duration for the HTTP request.
	pingUrl    string                 // pingUrl is the URL to be pinged to check the service status.
	socks5Pool *socks5.ConnectionPool // socks5Pool is the pool to obtain HTTP clients configured for SOCKS5.
}

// NewStatusCommand creates a new instance of StatusCommand.
func NewStatusCommand(timeout time.Duration, url string, pool *socks5.ConnectionPool) *StatusCommand {
	return &StatusCommand{timeout: timeout, pingUrl: url, socks5Pool: pool}
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

	ctx, cancel = context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	httpClient = c.socks5Pool.Borrow()
	defer c.socks5Pool.Return(httpClient)

	if request, err = http.NewRequestWithContext(ctx, http.MethodGet, c.pingUrl, http.NoBody); err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	if response, err = httpClient.Do(request); err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}

	if body, err = io.ReadAll(response.Body); err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	if err = response.Body.Close(); err != nil {
		return "", fmt.Errorf("close response body: %w", err)
	}

	return string(body), nil
}
