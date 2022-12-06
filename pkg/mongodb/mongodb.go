package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Mongo struct {
	client *mongo.Client
}

// New creates instance of mongodb client and pings primary server
func New(ctx context.Context, uri string) (*Mongo, error) {
	// Uses pooling internally
	opts := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return &Mongo{client: client}, nil
}

func (m *Mongo) CloseConnection(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

func (m *Mongo) Client() *mongo.Client {
	return m.client
}
