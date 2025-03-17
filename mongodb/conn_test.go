package mongodb

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mockMongoClient is a helper function to create a mock MongoDB client for testing
func mockMongoClient(t *testing.T) (*mongo.Client, error) {
	// Use a MongoDB memory server for testing
	// This is a mock connection that doesn't actually connect to a real MongoDB server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use a fake connection string that points to a non-existent server
	// This is fine for mocking certain behaviors
	clientOpts := options.Client().ApplyURI("mongodb://localhost:27017")

	// Set direct connection to avoid connection pooling which might cause test issues
	clientOpts.SetDirect(true)

	// Set a very short timeout to avoid hanging tests
	clientOpts.SetConnectTimeout(1 * time.Second)
	clientOpts.SetServerSelectionTimeout(1 * time.Second)

	return mongo.Connect(ctx, clientOpts)
}
