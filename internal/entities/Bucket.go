package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Bucket struct {
	ID         primitive.ObjectID `bson:"_id"`
	Name       string             `bson:"name"`
	Operations []Operation        `bson:"operations"`
	Module     string             `bson:"module"`
}

type Operation struct {
	// Post or Get or Delete
	Name string `json:"operation" validate:"required" bson:"name"`
	// Private or Public
	Type string `json:"type" validate:"required" bson:"type"`
	// List of jwt signing keys
	Keys []string `json:"keys" validate:"required" bson:"keys"`
}