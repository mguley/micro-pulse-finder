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
	Nats            NatsConfig      // NATS configuration.
	TLS             TLSConfig       // TLS configuration.
	InboundMessage  InboundMessage  // Inbound message service configuration.
	OutboundMessage OutboundMessage // Outbound message service configuration.
	Env             string          // Environment type (e.g., dev, prod).
}

// OutboundMessage holds configuration settings for outbound message service.
type OutboundMessage struct {
	BatchSize int // BatchSize is the max. number of concurrent URL processing goroutines.
}

// InboundMessage holds configuration settings for inbound message service.
type InboundMessage struct {
	BatchSize  int    // BatchSize is the max. number of concurrent URL processing goroutines.
	QueueGroup string // QueueGroup is the NATS queue group for load balancing.
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
		Nats:            loadNatsConfig(),
		TLS:             loadTLSConfig(),
		InboundMessage:  loadInboundMessageConfig(),
		OutboundMessage: loadOutboundMessageConfig(),
		Env:             getEnv("ENV", "dev"),
	}
}

// loadInboundMessageConfig loads inbound message service configuration.
func loadInboundMessageConfig() InboundMessage {
	inboundMessage := InboundMessage{
		BatchSize:  getEnvAsInt("INBOUND_MESSAGE_BATCH_SIZE", 0),
		QueueGroup: getEnv("INBOUND_MESSAGE_QUEUE_GROUP", ""),
	}

	checkRequiredVars("INBOUND_MESSAGE", map[string]string{
		"INBOUND_MESSAGE_BATCH_SIZE": string(rune(inboundMessage.BatchSize)),
	})
	return inboundMessage
}

// loadOutboundMessageConfig loads outbound message service configuration.
func loadOutboundMessageConfig() OutboundMessage {
	outboundMessage := OutboundMessage{
		BatchSize: getEnvAsInt("OUTBOUND_MESSAGE_BATCH_SIZE", 0),
	}

	checkRequiredVars("OUTBOUND_MESSAGE_BATCH_SIZE", map[string]string{
		"OUTBOUND_MESSAGE_BATCH_SIZE": string(rune(outboundMessage.BatchSize)),
	})
	return outboundMessage
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
