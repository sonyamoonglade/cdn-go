package dto

import (
	"animakuro/cdn/internal/entities"
)

type SaveFileDto struct {
	Name        string   `bson:"name"`
	Bucket      string   `bson:"bucket"`
	AvailableIn []string `bson:"availableIn"`
	MimeType    string   `bson:"mimeType"`
	UUID        string   `bson:"uuid"`
	Extension   string   `bson:"extension"`
}

type CreateBucketDto struct {
	Name       string               `json:"name" validate:"required"`
	Module     string               `json:"module"`
	Operations []entities.Operation `json:"operations" validate:"required"`
}
