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
	timeout      time.Duration  // timeout specifies the timeout duration for the HTTP request.
	pingUrl      string         // pingUrl is the URL to be pinged to check the service status.
	socks5Client *socks5.Client // socks5Client is the client to interact with SOCKS5 proxy.
}

// NewStatusCommand creates a new instance of StatusCommand.
func NewStatusCommand(timeout time.Duration, url string, client *socks5.Client) *StatusCommand {
	return &StatusCommand{timeout: timeout, pingUrl: url, socks5Client: client}
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

	if httpClient, err = c.socks5Client.Create(); err != nil {
		return "", fmt.Errorf("create socks5 client: %w", err)
	}

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
