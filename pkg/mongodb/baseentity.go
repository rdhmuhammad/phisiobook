package mongodb

import "go.mongodb.org/mongo-driver/bson/primitive"

type BaseEntity interface {
	GetID() primitive.ObjectID
	GetCollectionName() string
}
