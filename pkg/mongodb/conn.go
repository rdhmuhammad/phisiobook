package mongodb

import (
	"fmt"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Conn struct {
	client *mongo.Client
	access Connection
}

type Connection struct {
	Username string
	Password string
	Host     string
	Port     string
	Database string
}

func (c *Conn) GetClient() *mongo.Client {
	return c.client
}

func (c *Conn) GetDatabaseName() string {
	return c.access.Database
}

func NewConnection(connection Connection) *Conn {
	level := options.Logger().SetComponentLevel(options.LogComponentConnection, options.LogLevelInfo)
	conStr := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/?authSource=admin",
		connection.Username, connection.Password, connection.Host, connection.Port)
	client, err := mongo.Connect(options.Client().ApplyURI(conStr).SetLoggerOptions(level))
	if err != nil {
		panic(fmt.Errorf("MongoDB Connection Error => %s", err.Error()))
	}

	return &Conn{client: client, access: connection}
}
