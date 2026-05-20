package mongodb

import "go.mongodb.org/mongo-driver/v2/bson"

type BaseEntity interface {
	GetID() bson.ObjectID
	GetCollectionName() string
}
