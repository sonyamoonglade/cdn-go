package entities

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrFileAlreadyExists  = errors.New("file(s) already exist")
	ErrFileNotFound       = errors.New("file not found")
	ErrNoFiles            = errors.New("no files")
	ErrFileCantDelete     = errors.New("could not delete file")
	ErrFileAlreadyDeleted = errors.New("file has already deleted")
)

//todo: compute hash file <filename><size> to prevent same files upload
type File struct {
	ID primitive.ObjectID `bson:"_id"`
	// Folder where the file lives in a disk
	UUID string `bson:"uuid"`
	// List of cdn hosts where file is available
	AvailableIn []string `bson:"availableIn"`
	// Marked with `isDeletable` file no longer can be accessed via Get
	IsDeletable bool   `bson:"is_deletable"`
	Bucket      string `bson:"bucket"`
	MimeType    string `bson:"mimeType"`
	Extension   string `bson:"extension"`
}
