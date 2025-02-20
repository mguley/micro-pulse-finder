package entities

import (
	"fmt"
	"shared/mongodb/application/config"
	"sync"
)

var (
	mongo     *Mongo
	mongoOnce sync.Once
)

// GetMongo retrieves MongoDB configuration.
func GetMongo() *Mongo {
	mongoOnce.Do(func() {
		cfg := config.GetConfig()
		mongo = &Mongo{
			Host:       cfg.Mongo.Host,
			Port:       cfg.Mongo.Port,
			User:       cfg.Mongo.User,
			Pass:       cfg.Mongo.Pass,
			DB:         cfg.Mongo.DB,
			Collection: cfg.Mongo.Collection,
		}
	})
	return mongo
}

// Mongo represents the details of a MongoDB configuration.
type Mongo struct {
	Host       string // Host is the hostname of the MongoDB server.
	Port       string // Port is the port number of the MongoDB server.
	User       string // User is the username used to connect to the MongoDB server.
	Pass       string // Pass is the password used to connect to the MongoDB server.
	DB         string // DB is the name of the MongoDB database.
	Collection string // Collection is the name of the MongoDB collection.
}

// Address returns the full address of the MongoDB server.
func (m *Mongo) Address() (uri string, err error) {
	if m.Host == "" || m.Port == "" || m.User == "" || m.Pass == "" || m.DB == "" || m.Collection == "" {
		return "", fmt.Errorf("invalid address: some of the parameters are missing")
	}
	return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin", m.User, m.Pass, m.Host, m.Port, m.DB), nil
}
