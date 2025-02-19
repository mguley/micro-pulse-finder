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
	Mongo MongoConfig // MongoDB configuration.
}

// MongoConfig holds configuration settings for MongoDB.
type MongoConfig struct {
	Host       string // Host is the hostname of the MongoDB server.
	Port       string // Port is the port number of the MongoDB server.
	User       string // User is the username used to connect to the MongoDB server.
	Pass       string // Pass is the password used to connect to the MongoDB server.
	DB         string // DB is the name of the MongoDB database.
	Collection string // Collection is the name of the MongoDB collection.
}

// loadConfig loads configuration falling back to default values.
func loadConfig() *Config {
	return &Config{
		Mongo: loadMongoConfig(),
	}
}

// loadMongoConfig loads MongoDB configuration.
func loadMongoConfig() MongoConfig {
	mongo := MongoConfig{
		Host:       getEnv("MONGO_HOST", "localhost"),
		Port:       getEnv("MONGO_PORT", ""),
		User:       getEnv("MONGO_USER", ""),
		Pass:       getEnv("MONGO_PASS", ""),
		DB:         getEnv("MONGO_DB", ""),
		Collection: getEnv("MONGO_COLLECTION", ""),
	}

	checkRequiredVars("MONGO", map[string]string{
		"MONGO_HOST":       mongo.Host,
		"MONGO_PORT":       mongo.Port,
		"MONGO_USER":       mongo.User,
		"MONGO_PASS":       mongo.Pass,
		"MONGO_DB":         mongo.DB,
		"MONGO_COLLECTION": mongo.Collection,
	})

	return mongo
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
