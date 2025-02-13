package socks5

import "net/http"

// RoundTripWithUserAgent is an HTTP transport wrapper that modifies requests by setting a custom
// User-Agent header before forwarding them to the underlying RoundTripper.
type RoundTripWithUserAgent struct {
	roundTripper http.RoundTripper // roundTripper is the underlying round tripper to use for requests.
	agent        string            // agent is the User-Agent value to set in HTTP requests.
}

// RoundTrip executes a single HTTP transaction, adding the User-Agent header to the request.
func (r *RoundTripWithUserAgent) RoundTrip(request *http.Request) (response *http.Response, err error) {
	newRequest := request.Clone(request.Context())
	newRequest.Header.Set("User-Agent", r.agent)

	return r.roundTripper.RoundTrip(newRequest)
}
