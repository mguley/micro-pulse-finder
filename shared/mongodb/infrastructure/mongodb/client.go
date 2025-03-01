package mongodb

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client provides functionality to interact with MongoDB.
type Client struct {
	uri    string        // uri is the MongoDB connection URI.
	client *mongo.Client // client is the underlying MongoDB client.
	mu     sync.RWMutex  // mu protects client for safe concurrent access.
	logger *slog.Logger  // logger for structured logging.
}

// NewClient creates a new instance of Client.
func NewClient(uri string, logger *slog.Logger) *Client {
	return &Client{uri: uri, logger: logger}
}

// Connect establishes connection to MongoDB.
func (c *Client) Connect() (mongoClient *mongo.Client, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		c.logger.Info("MongoDB client already connected")
		return c.client, nil
	}

	c.logger.Info("Attempting to connect to MongoDB", "uri", c.uri)
	connectCtx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(c.uri)
	if mongoClient, err = mongo.Connect(connectCtx, clientOptions); err != nil {
		c.logger.Error("Failed to connect to MongoDB", "error", err)
		return nil, fmt.Errorf("could not connect to MongoDB: %w", err)
	}
	if err = mongoClient.Ping(connectCtx, nil); err != nil {
		c.logger.Error("MongoDB ping failed", "error", err)
		return nil, fmt.Errorf("could not ping MongoDB: %w", err)
	}

	c.client = mongoClient
	c.logger.Info("Successfully connected to MongoDB")
	return c.client, nil
}

// Close gracefully closes the connection to MongoDB.
func (c *Client) Close() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		c.logger.Info("MongoDB client already disconnected")
		return nil
	}

	c.logger.Info("Disconnecting MongoDB client")
	disconnectCtx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	if err = c.client.Disconnect(disconnectCtx); err != nil {
		c.logger.Error("Failed to disconnect MongoDB client", "error", err)
		return fmt.Errorf("could not disconnect MongoDB client: %w", err)
	}

	c.client = nil
	c.logger.Info("Successfully disconnected MongoDB client")
	return nil
}

// IsConnected checks if the MongoDB client is currently connected.
func (c *Client) IsConnected() (isConnected bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.client == nil {
		c.logger.Info("MongoDB client not initialized")
		return false
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	if err := c.client.Ping(pingCtx, nil); err != nil {
		c.logger.Error("MongoDB client ping failed", "error", err)
		return false
	}

	c.logger.Info("MongoDB client ping succeeded")
	return true
}
