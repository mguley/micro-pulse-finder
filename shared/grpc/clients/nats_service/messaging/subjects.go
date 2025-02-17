package messaging

// Subjects hold the NATS subjects used for inter-microservice communication.
const (
	// ProxyUrlRequest is the subject on which the proxy-service microservice listens for incoming URL requests.
	// Each message published to this subject should contain a valid URL that needs to be processed.
	ProxyUrlRequest = "proxy.url.request"

	// ProxyUrlResponse is the subject on which the proxy-service microservice publishes the results
	// after processing the URL (e.g., the HTTP response body, or a compressed version of it).
	// Other microservices can subscribe to this subject to receive the processed data.
	ProxyUrlResponse = "proxy.url.response"
)
