package codes

// Response codes returned by the proxy server.
const (
	SuccessResponse               = "250" // Indicates the command was successful.
	AuthenticationInvalidPassword = "515" // Indicates the provided password is invalid.
	AuthenticationRequired        = "514" // Indicates that authentication is required.
)
