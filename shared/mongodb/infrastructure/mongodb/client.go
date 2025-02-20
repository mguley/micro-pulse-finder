package mongodb

import (
	"context"
	"fmt"
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
}

// NewClient creates a new instance of Client.
func NewClient(uri string) *Client {
	return &Client{uri: uri}
}

// Connect establishes connection to MongoDB.
func (c *Client) Connect() (mongoClient *mongo.Client, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		return c.client, nil
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(c.uri)
	if mongoClient, err = mongo.Connect(connectCtx, clientOptions); err != nil {
		return nil, fmt.Errorf("could not connect to MongoDB: %w", err)
	}
	if err = mongoClient.Ping(connectCtx, nil); err != nil {
		return nil, fmt.Errorf("could not ping MongoDB: %w", err)
	}

	c.client = mongoClient
	return c.client, nil
}

// Close gracefully closes the connection to MongoDB.
func (c *Client) Close() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return nil
	}

	disconnectCtx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	if err = c.client.Disconnect(disconnectCtx); err != nil {
		return fmt.Errorf("could not disconnect MongoDB client: %w", err)
	}

	c.client = nil
	return nil
}

// IsConnected checks if the MongoDB client is currently connected.
func (c *Client) IsConnected() (isConnected bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.client == nil {
		return false
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	return (c.client.Ping(pingCtx, nil)) == nil
}
