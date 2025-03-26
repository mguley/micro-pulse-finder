package config

import (
	"fmt"
	"os"
	"sync"
)

var (
	once   sync.Once
	config *Config
)

// GetConfig retrieves the global configuration instance.
//
// Returns:
//   - *Config: A pointer to the configuration struct containing all configuration settings.
func GetConfig() *Config {
	once.Do(func() {
		config = loadConfig()
	})
	return config
}

// Config holds all configuration settings for the application.
//
// Fields:
//   - Nats: NATS configuration settings.
//   - TLS:  TLS configuration settings.
//   - RPC:  RPC configuration settings.
//   - Env:  Environment type (e.g., dev, prod).
type Config struct {
	Nats NatsConfig
	TLS  TLSConfig
	RPC  RPCConfig
	Env  string
}

// RPCConfig holds configuration settings for the RPC server.
//
// Fields:
//   - Port: Port on which the Bus gRPC server listens.
type RPCConfig struct {
	Port string
}

// TLSConfig holds configuration settings for TLS.
//
// Fields:
//   - Certificate: Path to the TLS certificate file.
//   - Key:         Path to the TLS key file.
type TLSConfig struct {
	Certificate string
	Key         string
}

// NatsConfig holds configuration settings for the NATS server.
//
// Fields:
//   - Host: Hostname of the NATS server.
//   - Port: Port number of the NATS server.
type NatsConfig struct {
	Host string
	Port string
}

// loadConfig loads the application configuration by reading the environment variables.
//
// Returns:
//   - *Config: A pointer to the newly created configuration structure.
func loadConfig() *Config {
	return &Config{
		Nats: loadNatsConfig(),
		TLS:  loadTLSConfig(),
		RPC:  loadRPCConfig(),
		Env:  getEnv("ENV", "dev"),
	}
}

// loadRPCConfig loads RPC configuration settings from environment variables.
//
// Returns:
//   - RPCConfig: An instance of RPCConfig with the appropriate port setting.
func loadRPCConfig() RPCConfig {
	rpc := RPCConfig{
		Port: getEnv("NATS_RPC_SERVER_PORT", ""),
	}

	checkRequiredVars("NATS_RPC", map[string]string{
		"NATS_RPC_SERVER_PORT": rpc.Port,
	})
	return rpc
}

// loadTLSConfig loads TLS configuration settings from environment variables.
//
// Returns:
//   - TLSConfig: An instance of TLSConfig with paths to the certificate and key files.
func loadTLSConfig() TLSConfig {
	tls := TLSConfig{
		Certificate: getEnv("TLS_CERTIFICATE", ""),
		Key:         getEnv("TLS_KEY", ""),
	}

	checkRequiredVars("TLS", map[string]string{
		"TLS_CERTIFICATE": tls.Certificate,
		"TLS_KEY":         tls.Key,
	})
	return tls
}

// loadNatsConfig loads NATS configuration settings from environment variables.
//
// Returns:
//   - NatsConfig: An instance of NatsConfig with NATS server hostname and port.
func loadNatsConfig() NatsConfig {
	nats := NatsConfig{
		Host: getEnv("NATS_HOST", "localhost"),
		Port: getEnv("NATS_PORT", ""),
	}

	// Ensure required values are present
	checkRequiredVars("NATS", map[string]string{
		"NATS_HOST": nats.Host,
		"NATS_PORT": nats.Port,
	})
	return nats
}

// getEnv fetches the value of an environment variable.
//
// Parameters:
//   - key:      The name of the environment variable.
//   - fallback: The default value to return if the environment variable is not set.
//
// Returns:
//   - string: The value of the environment variable or the fallback.
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

// checkRequiredVars ensures that all required environment variables are set.
//
// Parameters:
//   - section: A string representing the configuration section (e.g., "NATS", "TLS", or "NATS_RPC").
//   - vars:    A map where keys are environment variable names and values are their corresponding values.
//
// Returns:
//   - None: This function panics if any required variable is missing.
func checkRequiredVars(section string, vars map[string]string) {
	for key, value := range vars {
		if value == "" {
			panic(fmt.Sprintf("%s configuration error: %s is required", section, key))
		}
	}
}
