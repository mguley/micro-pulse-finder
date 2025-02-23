package entities

import (
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// StatusPending represents URL that is pending processing.
	StatusPending = "pending"
	// StatusProcessed represents URL that has been successfully processed.
	StatusProcessed = "processed"
	// StatusFailed represents URL that failed processing.
	StatusFailed = "failed"
)

// urlEntityPool is the on-demand pool for Url entities.
var urlEntityPool = urlPool()

// Url represents the URL entity.
type Url struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`      // Id is the unique identifier of the URL.
	Address   string             `bson:"address" json:"address"`       // Address is the URL address to be processed.
	Status    string             `bson:"status" json:"status"`         // Status is the processing status of the URL.
	Source    string             `bson:"source" json:"source"`         // Source is the source who created the record.
	Processed time.Time          `bson:"processed" json:"processed"`   // Processed is the time when URL was processed.
	CreatedAt time.Time          `bson:"created_at" json:"created_at"` // CreatedAt is the time when URL was created.
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"` // UpdatedAt is the time when URL was updated.
}

// String implements the fmt.Stringer interface, providing a readable string representation of the Url entity.
func (e *Url) String() (result string) {
	return fmt.Sprintf("Id: %s, Address: %s, Status: %s, Source: %s", e.Id, e.Address, e.Status, e.Source)
}

// urlPool returns a function that provides access to a *sync.Pool for Url entities.
// It uses sync.Once to ensure the pool is created only once.
func urlPool() func() *sync.Pool {
	var (
		once sync.Once
		pool *sync.Pool
	)
	return func() *sync.Pool {
		once.Do(func() {
			pool = &sync.Pool{
				New: func() interface{} {
					return &Url{}
				},
			}
		})
		return pool
	}
}

// GetUrl retrieves Url entity from the pool.
func GetUrl() *Url {
	return urlEntityPool().Get().(*Url)
}

// Reset resets the Url fields to their zero values, preparing it for reuse.
func (e *Url) Reset() *Url {
	e.Id = primitive.NilObjectID
	e.Address = ""
	e.Status = ""
	e.Source = ""
	e.Processed = time.Time{}
	e.CreatedAt = time.Time{}
	e.UpdatedAt = time.Time{}
	return e
}

// Release puts the Url back into the pool after resetting its state.
func (e *Url) Release() {
	urlEntityPool().Put(e.Reset())
}
