package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Conn struct {
	Client   *mongo.Client
	database string
}

func NewConn(
	ctx context.Context,
	uri string,
	database string,
) (*Conn, error) {
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)

	if err != nil {
		return nil, err
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &Conn{
		Client:   client,
		database: database,
	}, nil
}

// Database returns a handle to the specified database
func (conn *Conn) Database(name string) *mongo.Database {
	if name == "" {
		name = conn.database
	}
	return conn.Client.Database(name)
}

// Close disconnects from MongoDB
func (conn *Conn) Close(ctx context.Context) error {
	return conn.Client.Disconnect(ctx)
}

// ListDatabases returns a list of database names
func (conn *Conn) ListDatabases(ctx context.Context) ([]string, error) {
	dbs, err := conn.Client.ListDatabaseNames(ctx, struct{}{})
	if err != nil {
		return nil, err
	}
	return dbs, nil
}
