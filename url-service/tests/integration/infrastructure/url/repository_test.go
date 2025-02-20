package url

import (
	"context"
	"testing"
	"time"
	"url-service/domain/entities"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestRepository_Save verifies that URL entity is successfully saved.
func TestRepository_Save(t *testing.T) {
	container := SetupTestContainer(t)
	repository := container.MongoRepository.Get()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()

	// Create a new URL entity.
	now := time.Now()
	urlEntity := &entities.Url{
		Address:   "https://example.com",
		Status:    entities.StatusPending,
		Source:    "test",
		Processed: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Save entity.
	err := repository.Save(ctx, urlEntity)
	require.NoError(t, err, "Failed to save URL entity")

	// Verify insertion by fetching the document using a filter on the address.
	filter := bson.M{"address": urlEntity.Address}
	urls, err := repository.FetchBatch(ctx, filter, 1)
	require.NoError(t, err, "Failed to fetch URLs")
	require.NotEmpty(t, urls, "Expected at least one URL to be returned")
	require.Equal(t, urlEntity.Address, urls[0].Address, "Expected matching address")
}

// TestRepository_FetchBatch verifies that filtering works as expected.
func TestRepository_FetchBatch(t *testing.T) {
	container := SetupTestContainer(t)
	repository := container.MongoRepository.Get()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()

	// Insert two URL entities with a specific source value.
	now := time.Now()
	urlsToInsert := []*entities.Url{
		{
			Address:   "https://example.com/1",
			Status:    entities.StatusPending,
			Source:    "integration_test",
			Processed: now,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Address:   "https://example.com/2",
			Status:    entities.StatusPending,
			Source:    "integration_test",
			Processed: now,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, url := range urlsToInsert {
		err := repository.Save(ctx, url)
		require.NoError(t, err, "Failed to save URL entity")
	}

	// Use FetchBatch to retrieve documents filtered by source.
	filter := bson.M{"source": "integration_test"}
	fetched, err := repository.FetchBatch(ctx, filter, 100)
	require.NoError(t, err, "Failed to fetch URLs")
	require.GreaterOrEqual(t, len(fetched), 2, "Expected to find 2 URLs but found %d", len(fetched))
}

// TestRepository_UpdateFields verifies that updating specific fields works as expected.
func TestRepository_UpdateFields(t *testing.T) {
	container := SetupTestContainer(t)
	repository := container.MongoRepository.Get()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()

	// Insert a new URL entity.
	now := time.Now()
	urlEntity := &entities.Url{
		Address:   "https://example.org",
		Status:    entities.StatusPending,
		Source:    "update_test",
		Processed: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := repository.Save(ctx, urlEntity)
	require.NoError(t, err, "Failed to save URL entity")

	// Update the URL entity by changing its status and updating the timestamp.
	newStatus := entities.StatusProcessed
	updateTime := time.Now()
	updateFields := bson.M{
		"status":     newStatus,
		"updated_at": updateTime,
	}
	err = repository.UpdateFields(ctx, urlEntity.Id.Hex(), updateFields)
	require.NoError(t, err, "Failed to update URL fields")

	// Fetch the updated entity and verify the changes.
	filter := bson.M{"_id": urlEntity.Id}
	fetched, err := repository.FetchBatch(ctx, filter, 100)
	require.NoError(t, err, "Failed to fetch URLs")
	require.Len(t, fetched, 1, "Expected to find one URL but found %d", len(fetched))
	require.Equal(t, newStatus, fetched[0].Status, "Expected status to be updated")
	require.WithinDuration(t, updateTime, fetched[0].UpdatedAt, time.Second, "Expected updated_at to be updated")
}

// TestRepository_BulkUpdateFields verifies that multiple URL entities can be updated at once.
func TestRepository_BulkUpdateFields(t *testing.T) {
	container := SetupTestContainer(t)
	repository := container.MongoRepository.Get()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()

	// Insert multiple URL entities.
	now := time.Now()
	urlEntities := []*entities.Url{
		{
			Address:   "https://bulk.example.com/1",
			Status:    entities.StatusPending,
			Source:    "bulk_update_test",
			Processed: now,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Address:   "https://bulk.example.com/2",
			Status:    entities.StatusPending,
			Source:    "bulk_update_test",
			Processed: now,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	var ids = make([]string, 0, len(urlEntities))
	for _, url := range urlEntities {
		err := repository.Save(ctx, url)
		require.NoError(t, err, "Failed to save URL entity")
		ids = append(ids, url.Id.Hex())
	}

	// Update both entities' status in a bulk operation.
	newStatus := entities.StatusFailed
	updateTime := time.Now()
	updateFields := bson.M{
		"status":     newStatus,
		"updated_at": updateTime,
	}
	err := repository.BulkUpdateFields(ctx, ids, updateFields)
	require.NoError(t, err, "Failed to bulk update URL fields")

	// Convert the slice of string IDs to ObjectIDs.
	var objectIDs = make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)
		require.NoError(t, err, "Invalid ObjectID format")
		objectIDs = append(objectIDs, oid)
	}

	// Verify that both documents have been updated.
	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	updatedDocs, err := repository.FetchBatch(ctx, filter, 100)
	require.NoError(t, err, "Failed to fetch URLs")
	require.Len(t, updatedDocs, len(ids), "Expected to find %d URLs but found %d", len(ids), len(updatedDocs))
	for _, doc := range updatedDocs {
		require.Equal(t, newStatus, doc.Status, "Expected status to be updated")
		require.WithinDuration(t, updateTime, doc.UpdatedAt, time.Second, "Expected updated_at to be updated")
	}
}
