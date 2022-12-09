package cdn

import (
	"context"

	"animakuro/cdn/internal/cdn/dto"
	"animakuro/cdn/internal/entities"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const (
	BucketCollection = "bucket"
	FileCollection   = "file"
)

type Repository interface {
	GetBucket(ctx context.Context, name string) (*entities.Bucket, error)
	GetAllBuckets(ctx context.Context) ([]*entities.Bucket, error)
	SaveBucket(ctx context.Context, dto dto.CreateBucketDto) (*entities.Bucket, error)

	GetFile(ctx context.Context, bucket string, uuid string) (*entities.File, error)
	SaveFile(ctx context.Context, dto dto.SaveFileDto) (bool, error)

	// Actually deletes the file from database forever
	DeleteFile(ctx context.Context, bucket string, uuid string) (bool, error)
	// Makes file ready to be deleted.
	// Marked file no longer can be accessed via GetFile
	MarkAsDeletable(ctx context.Context, bucket string, mongoID primitive.ObjectID) error
}

type cdnRepo struct {
	db     *mongo.Database
	logger *zap.SugaredLogger
}

func NewRepository(logger *zap.SugaredLogger, dbname string, client *mongo.Client) *cdnRepo {
	return &cdnRepo{
		logger: logger,
		db:     client.Database(dbname),
	}
}

func (r *cdnRepo) SaveBucket(ctx context.Context, dto dto.CreateBucketDto) (*entities.Bucket, error) {
	res, err := r.db.Collection(BucketCollection).InsertOne(ctx, dto)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, nil
		}

		return nil, err
	}

	return &entities.Bucket{
		ID:         res.InsertedID.(primitive.ObjectID),
		Name:       dto.Name,
		Operations: dto.Operations,
		Module:     dto.Module,
	}, nil
}

func (r *cdnRepo) GetBucket(ctx context.Context, name string) (*entities.Bucket, error) {

	var b entities.Bucket

	q := bson.D{{"name", name}}
	res := r.db.Collection(BucketCollection).FindOne(ctx, q)

	if err := res.Decode(&b); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return &b, nil
}

func (r *cdnRepo) GetFile(ctx context.Context, bucket string, uuid string) (*entities.File, error) {

	var f entities.File

	q := bson.D{{"bucket", bucket}, {"uuid", uuid}}
	res := r.db.Collection(FileCollection).FindOne(ctx, q)

	if err := res.Decode(&f); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return &f, nil
}

func (r *cdnRepo) SaveFile(ctx context.Context, dto dto.SaveFileDto) (bool, error) {
	_, err := r.db.Collection(FileCollection).InsertOne(ctx, dto)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *cdnRepo) GetAllBuckets(ctx context.Context) ([]*entities.Bucket, error) {

	c, err := r.db.Collection(BucketCollection).Find(ctx, bson.D{{}})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	var buckets []*entities.Bucket

	for c.Next(ctx) {
		var b entities.Bucket
		err := c.Decode(&b)
		if err != nil {
			return nil, err
		}
		buckets = append(buckets, &b)
	}

	return buckets, nil
}

func (r *cdnRepo) DeleteFile(ctx context.Context, bucket string, uuid string) (bool, error) {

	q := bson.D{{"bucket", bucket}, {"uuid", uuid}}

	res, err := r.db.Collection(FileCollection).DeleteOne(ctx, q)
	if err != nil {
		return false, err
	}

	if res.DeletedCount == 0 {
		return false, nil
	}

	return true, nil
}

func (r *cdnRepo) MarkAsDeletable(ctx context.Context, bucket string, mongoID primitive.ObjectID) error {

	q := bson.D{{"_id", mongoID}, {"is_deletable", false}}

	// Update
	update := bson.D{{"$set", bson.D{{"is_deletable", true}}}}

	_, err := r.db.Collection(FileCollection).UpdateOne(ctx, q, update)
	if err != nil {
		return err
	}

	return nil
}
