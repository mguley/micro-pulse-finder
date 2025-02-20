package interfaces

import (
	"context"
	"url-service/domain/entities"

	"go.mongodb.org/mongo-driver/bson"
)

// UrlRepository defines the contract for interacting with URL entities in the persistence layer.
type UrlRepository interface {
	// Save persists a new URL entity into the data source.
	Save(ctx context.Context, url *entities.Url) (err error)

	// FetchBatch retrieves a batch of URLs matching the given filter.
	FetchBatch(ctx context.Context, filter bson.M, limit int) (list []*entities.Url, err error)

	// UpdateFields updates URL entity in the MongoDB collection by its ID using dynamic update fields.
	UpdateFields(ctx context.Context, id string, updateFields bson.M) (err error)

	// BulkUpdateFields updates multiple entities in the MongoDB collection by their IDs using dynamic update fields.
	BulkUpdateFields(ctx context.Context, ids []string, updateFields bson.M) (err error)
}
