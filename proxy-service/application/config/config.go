package config

import (
	"fmt"
	"os"
	"strconv"
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
	Nats         NatsConfig         // NATS configuration.
	TLS          TLSConfig          // TLS configuration.
	RPC          RPCConfig          // RPC configuration.
	Proxy        ProxyConfig        // Proxy configuration.
	Pool         PoolConfig         // Pool configuration.
	UrlProcessor UrlProcessorConfig // UrlProcessor configuration.
	Env          string             // Environment type (e.g., dev, prod).
}

// UrlProcessorConfig holds configuration settings for UrlProcessorService.
type UrlProcessorConfig struct {
	BatchSize  int    // BatchSize is the max. number of concurrent URL processing goroutines.
	QueueGroup string // QueueGroup is the NATS queue group for load balancing.
}

// ProxyConfig holds configuration settings for Proxy.
type ProxyConfig struct {
	Host            string // Host is the hostname of the proxy server.
	Port            string // Port is the port number of the proxy server.
	ControlPassword string // ControlPassword is the auth password used for the proxy's control port.
	ControlPort     string // ControlPort is the port number of the proxy's control port.
	Url             string // Url is the URL used to check the proxy's status or connectivity.
}

// PoolConfig holds configuration options for the connection pool.
type PoolConfig struct {
	MaxSize         int // MaxSize is the maximum number of connections in the pool.
	RefreshInterval int // RefreshInterval is the interval at which connections are refreshed.
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
	Host    string // Host is the hostname of the NATS server.
	Port    string // Port is the port number of the NATS server.
	RpcHost string // RpcHost is the address of the NATS gRPC server.
	RpcPort string // RpcPort is the port number of the NATS gRPC server.
}

// loadConfig loads configuration falling back to default values.
func loadConfig() *Config {
	return &Config{
		Nats:         loadNatsConfig(),
		TLS:          loadTLSConfig(),
		RPC:          loadRPCConfig(),
		Proxy:        loadProxyConfig(),
		Pool:         loadPoolConfig(),
		UrlProcessor: loadUrlProcessorConfig(),
		Env:          getEnv("ENV", "dev"),
	}
}

// loadUrlProcessorConfig loads url processor service configuration.
func loadUrlProcessorConfig() UrlProcessorConfig {
	processor := UrlProcessorConfig{
		BatchSize:  getEnvAsInt("URL_PROCESSOR_BATCH_SIZE", 0),
		QueueGroup: getEnv("URL_PROCESSOR_QUEUE_GROUP", ""),
	}

	checkRequiredVars("URL PROCESSOR", map[string]string{
		"URL_PROCESSOR_BATCH_SIZE": string(rune(processor.BatchSize)),
	})
	return processor
}

// loadPoolConfig loads Pool configuration.
func loadPoolConfig() PoolConfig {
	pool := PoolConfig{
		MaxSize:         getEnvAsInt("POOL_MAX_SIZE", 0),
		RefreshInterval: getEnvAsInt("POOL_REFRESH_INTERVAL", 0),
	}

	checkRequiredVars("POOL", map[string]string{
		"POOL_MAX_SIZE":         string(rune(pool.MaxSize)),
		"POOL_REFRESH_INTERVAL": string(rune(pool.RefreshInterval)),
	})
	return pool
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
		Port: getEnv("PROXY_RPC_SERVER_PORT", ""),
	}

	checkRequiredVars("PROXY_RPC", map[string]string{
		"PROXY_RPC_SERVER_PORT": rpc.Port,
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
		Host:    getEnv("NATS_HOST", "localhost"),
		Port:    getEnv("NATS_PORT", ""),
		RpcHost: getEnv("NATS_RPC_HOST", "localhost"),
		RpcPort: getEnv("NATS_RPC_PORT", ""),
	}

	// Ensure required values are present
	checkRequiredVars("NATS", map[string]string{
		"NATS_HOST":     nats.Host,
		"NATS_PORT":     nats.Port,
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

// getEnvAsInt fetches the value of an environment variable as an integer or returns a fallback.
func getEnvAsInt(key string, fallback int) int {
	v := getEnv(key, "")
	if value, err := strconv.Atoi(v); err == nil {
		return value
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
