package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type CacheChat struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	ToID    string             `bson:"to_id" json:"to_id"`
	Message string             `bson:"message" json:"message"`
	RoomID  string             `bson:"room_id" json:"room_id"`
}

func (receiver CacheChat) GetID() primitive.ObjectID {
	return receiver.ID
}

func (receiver CacheChat) GetCollectionName() string {
	return "cache_chat"
}
