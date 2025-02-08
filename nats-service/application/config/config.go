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

// GetConfig retrieves the configuration.
func GetConfig() *Config {
	once.Do(func() {
		config = loadConfig()
	})
	return config
}

// Config holds configuration settings.
type Config struct {
	Nats NatsConfig // NATS configuration.
	TLS  TLSConfig  // TLS configuration.
	RPC  RPCConfig  // RPC configuration
	Env  string     // Environment type (e.g., dev, prod).
}

type RPCConfig struct {
	Port string // Port is the port for the Bus gRPC server.
}

// TLSConfig holds configuration settings for TLS.
type TLSConfig struct {
	Certificate string // Certificate is the path to the TLS certificate file.
	Key         string // Key is the path to the TLS key file.
}

// NatsConfig holds configuration settings for NATS.
type NatsConfig struct {
	Host string // Host is the hostname of the NATS server.
	Port string // Port is the port number of the NATS server.
}

// loadConfig loads configuration falling back to default values.
func loadConfig() *Config {
	return &Config{
		Nats: loadNatsConfig(),
		TLS:  loadTLSConfig(),
		RPC:  loadRPCConfig(),
		Env:  getEnv("ENV", "dev"),
	}
}

// loadRPCConfig loads RPC configuration.
func loadRPCConfig() RPCConfig {
	rpc := RPCConfig{
		Port: getEnv("RPC_SERVER_PORT", ""),
	}

	checkRequiredVars("RPC", map[string]string{
		"RPC_SERVER_PORT": rpc.Port,
	})
	return rpc
}

// loadTLSConfig loads TLS configuration.
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

// loadNatsConfig loads NATS configuration.
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

// getEnv fetches the value of an environment variable or returns a fallback.
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

// checkRequiredVars ensures required environment variables are set.
func checkRequiredVars(section string, vars map[string]string) {
	for key, value := range vars {
		if value == "" {
			panic(fmt.Sprintf("%s configuration error: %s is required", section, key))
		}
	}
}
