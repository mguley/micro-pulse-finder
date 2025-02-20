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
	Env  string     // Environment type (e.g., dev, prod).
}

// NatsConfig holds configuration settings for NATS.
type NatsConfig struct {
	RpcHost string // RpcHost is the address of the NATS gRPC server.
	RpcPort string // RpcPort is the port number of the NATS gRPC server.
}

// TLSConfig holds configuration settings for TLS.
type TLSConfig struct {
	Certificate string // Certificate is the path to the TLS certificate file.
	Key         string // Key is the path to the TLS key file.
}

// loadConfig loads configuration falling back to default values.
func loadConfig() *Config {
	return &Config{
		Nats: loadNatsConfig(),
		TLS:  loadTLSConfig(),
		Env:  getEnv("ENV", "dev"),
	}
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
		RpcHost: getEnv("NATS_RPC_HOST", "localhost"),
		RpcPort: getEnv("NATS_RPC_PORT", ""),
	}

	checkRequiredVars("NATS", map[string]string{
		"NATS_RPC_HOST": nats.RpcHost,
		"NATS_RPC_PORT": nats.RpcPort,
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
