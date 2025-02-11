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
	Nats  NatsConfig  // NATS configuration.
	TLS   TLSConfig   // TLS configuration.
	RPC   RPCConfig   // RPC configuration.
	Proxy ProxyConfig // Proxy configuration.
	Env   string      // Environment type (e.g., dev, prod).
}

// ProxyConfig holds configuration settings for Proxy.
type ProxyConfig struct {
	Host            string // Host is the hostname of the proxy server.
	Port            string // Port is the port number of the proxy server.
	ControlPassword string // ControlPassword is the auth password used for the proxy's control port.
	ControlPort     string // ControlPort is the port number of the proxy's control port.
	Url             string // Url is the URL used to check the proxy's status or connectivity.
}

// RPCConfig holds configuration settings for RPC.
type RPCConfig struct {
	Port string // Port is the port for the Proxy gRPC server.
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
		Nats:  loadNatsConfig(),
		TLS:   loadTLSConfig(),
		RPC:   loadRPCConfig(),
		Proxy: loadProxyConfig(),
		Env:   getEnv("ENV", "dev"),
	}
}

// loadProxyConfig loads Proxy configuration.
func loadProxyConfig() ProxyConfig {
	proxy := ProxyConfig{
		Host:            getEnv("PROXY_HOST", ""),
		Port:            getEnv("PROXY_PORT", ""),
		ControlPassword: getEnv("PROXY_CONTROL_PASSWORD", ""),
		ControlPort:     getEnv("PROXY_CONTROL_PORT", ""),
		Url:             getEnv("PROXY_URL", ""),
	}

	checkRequiredVars("PROXY", map[string]string{
		"PROXY_HOST":             proxy.Host,
		"PROXY_PORT":             proxy.Port,
		"PROXY_CONTROL_PASSWORD": proxy.ControlPassword,
		"PROXY_CONTROL_PORT":     proxy.ControlPort,
		"PROXY_URL":              proxy.Url,
	})
	return proxy
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
