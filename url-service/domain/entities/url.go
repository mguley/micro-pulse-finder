package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusPending   = "pending"
	StatusProcessed = "processed"
	StatusFailed    = "failed"
)

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
