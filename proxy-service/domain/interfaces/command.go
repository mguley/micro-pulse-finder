package interfaces

// Command defines the contract for executing a command.
type Command interface {
	// Execute runs the command and returns an error if it fails.
	Execute() (err error)
}
