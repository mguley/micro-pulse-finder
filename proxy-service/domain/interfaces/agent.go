package interfaces

// Agent defines the contract for generating User-Agent strings.
type Agent interface {
	// Generate returns a user agent string that simulates a browser and a device.
	Generate() (userAgent string)
}
