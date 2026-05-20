package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CacheChat struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	CreatedAt time.Time     `bson:"created_at,omitempty" json:"created_at"`
	FromID    string        `bson:"from_id" json:"from_id"`
	Read      bool          `bson:"read" json:"read"`
	ReadAt    time.Time     `bson:"read_at" json:"read_at"`
	ToID      string        `bson:"to_id" json:"to_id"`
	Message   string        `bson:"message" json:"message"`
	RoomID    string        `bson:"room_id" json:"room_id"`
}

func (receiver CacheChat) GetID() bson.ObjectID {
	return receiver.ID
}

func (receiver CacheChat) GetCollectionName() string {
	return "cache_chat"
}
