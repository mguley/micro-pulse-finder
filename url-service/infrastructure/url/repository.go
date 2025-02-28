package url

import (
	"context"
	"fmt"
	"log/slog"
	"url-service/domain/entities"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository provides a MongoDB-based implementation for managing URL entities.
type Repository struct {
	client     *mongo.Client     // client is the MongoDB client.
	collection *mongo.Collection // collection is the MongoDB collection.
	logger     *slog.Logger
}

// NewRepository creates a new instance of Repository.
func NewRepository(mongoClient *mongo.Client, collection *mongo.Collection, logger *slog.Logger) *Repository {
	return &Repository{client: mongoClient, collection: collection, logger: logger}
}

// Save persists a new URL entity into the MongoDB collection.
func (r *Repository) Save(ctx context.Context, url *entities.Url) (err error) {
	if url.Id.IsZero() {
		url.Id = primitive.NewObjectID()
	}

	var insertResult *mongo.InsertOneResult
	if insertResult, err = r.collection.InsertOne(ctx, url); err != nil {
		r.logger.Error("Failed to execute an insert command", "error", err)
		return fmt.Errorf("insert one: %w", err)
	}
	r.logger.Info("Inserted URL", "address", url.Address, "insertedID", insertResult.InsertedID)
	return nil
}

// FetchBatch retrieves a batch of URLs matching the given filter.
// The filter parameter is of type bson.M, allowing dynamic filtering.
func (r *Repository) FetchBatch(ctx context.Context, filter bson.M, limit int) (list []*entities.Url, err error) {
	var (
		opts   = options.Find().SetLimit(int64(limit))
		cursor *mongo.Cursor
	)

	if cursor, err = r.collection.Find(ctx, filter, opts); err != nil {
		r.logger.Error("Failed to execute a find command", "error", err)
		return nil, fmt.Errorf("find by filter: %w", err)
	}
	defer func() {
		if closeErr := cursor.Close(ctx); closeErr != nil {
			r.logger.Error("Failed to close cursor", "error", closeErr)
		}
	}()

	if err = cursor.All(ctx, &list); err != nil {
		r.logger.Error("Failed to execute cursor's command", "error", err)
		return nil, fmt.Errorf("decode URL documents: %w", err)
	}
	return list, nil
}

// UpdateFields updates URL entity in the MongoDB collection by its ID using dynamic update fields.
// The updateFields parameter is a bson.M map that specifies the fields to update.
func (r *Repository) UpdateFields(ctx context.Context, id string, updateFields bson.M) (err error) {
	var (
		objectId     primitive.ObjectID
		update       = bson.M{"$set": updateFields}
		updateResult *mongo.UpdateResult
	)

	if objectId, err = primitive.ObjectIDFromHex(id); err != nil {
		r.logger.Error("Failed to parse object ID", "error", err)
		return fmt.Errorf("ID format: %w", err)
	}
	if updateResult, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectId}, update); err != nil {
		r.logger.Error("Failed to execute an update command", "objectId", objectId, "error", err)
		return fmt.Errorf("update for ID %s: %w", id, err)
	}

	if updateResult.MatchedCount == 0 {
		r.logger.Error("No rows were updated", "id", id)
		return fmt.Errorf("ID %s not found", id)
	}
	return nil
}

// BulkUpdateFields updates multiple URL entities in the MongoDB collection by their IDs using dynamic update fields.
// The updateFields parameter allows updating any set of fields provided in a bson.M map.
func (r *Repository) BulkUpdateFields(ctx context.Context, ids []string, updateFields bson.M) (err error) {
	var objectIds []primitive.ObjectID
	if objectIds, err = r.parseObjectIDs(ids); err != nil {
		return err
	}

	var (
		filter       = bson.M{"_id": bson.M{"$in": objectIds}}
		update       = bson.M{"$set": updateFields}
		updateResult *mongo.UpdateResult
	)

	if updateResult, err = r.collection.UpdateMany(ctx, filter, update); err != nil {
		r.logger.Error("Failed to execute an update command", "filter", filter, "error", err)
		return fmt.Errorf("bulk update: %w", err)
	}

	if updateResult.MatchedCount == 0 {
		r.logger.Error("No rows were updated", "ids", ids)
		return fmt.Errorf("no documents found for ids: %s", ids)
	}

	return nil
}

// parseObjectIDs converts a slice of string IDs to a slice of MongoDB ObjectIDs.
func (r *Repository) parseObjectIDs(ids []string) (list []primitive.ObjectID, err error) {
	list = make([]primitive.ObjectID, 0, len(ids))

	var objectId primitive.ObjectID
	for _, id := range ids {
		if objectId, err = primitive.ObjectIDFromHex(id); err != nil {
			r.logger.Error("Failed to parse object ID", "objectId", objectId, "error", err)
			return nil, fmt.Errorf("ID format: %w", err)
		}
		list = append(list, objectId)
	}

	return list, nil
}
