package entities

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrFileAlreadyExists = errors.New("file(s) already exist")
	ErrFileNotFound      = errors.New("file not found")
	ErrFileCantRemove    = errors.New("could not remove file")
	ErrNoFiles           = errors.New("no files")
)

//todo: compute hash file <filename><size> to prevent same files upload
type File struct {
	ID          primitive.ObjectID `bson:"_id"`
	UUID        string             `bson:"uuid"`
	Bucket      string             `bson:"bucket"`
	AvailableIn []string           `bson:"availableIn"`
	MimeType    string             `bson:"mimeType"`
	Extension   string             `bson:"extension"`
}
