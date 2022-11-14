package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
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
