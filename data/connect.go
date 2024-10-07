package data

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DBClient *mongo.Client

func ConnectDb() *mongo.Client {

	fmt.Println(DBClient)

	if DBClient != nil {
		return DBClient
	}

	uri := os.Getenv("DB_URI")

	if uri == "" {
		panic("NO DB URI")
	}

	client,err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	return client
}