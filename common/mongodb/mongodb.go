package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"

	"github.com/syamsv/Advtrix/config"
)

var client *mongo.Client
var db *mongo.Database
var log *zap.Logger

func Init() {
	log = zap.L().Named("mongodb")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(options.Client().ApplyURI(config.MONGODB_URI))
	if err != nil {
		log.Fatal("failed to connect", zap.Error(err))
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal("failed to ping", zap.Error(err))
	}

	db = client.Database(config.MONGODB_NAME)
	log.Info("connected", zap.String("uri", config.MONGODB_URI), zap.String("database", config.MONGODB_NAME))
}

func Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Disconnect(ctx); err != nil {
		log.Error("error disconnecting", zap.Error(err))
	}
	log.Info("disconnected")
}

func Collection(name string) *mongo.Collection {
	return db.Collection(name)
}

func InsertOne(ctx context.Context, collection string, doc any) (*mongo.InsertOneResult, error) {
	return db.Collection(collection).InsertOne(ctx, doc)
}

func InsertMany(ctx context.Context, collection string, docs []any) (*mongo.InsertManyResult, error) {
	return db.Collection(collection).InsertMany(ctx, docs)
}

func FindOne(ctx context.Context, collection string, filter bson.M, result any) error {
	return db.Collection(collection).FindOne(ctx, filter).Decode(result)
}

func FindMany(ctx context.Context, collection string, filter bson.M, opts ...options.Lister[options.FindOptions]) (*mongo.Cursor, error) {
	return db.Collection(collection).Find(ctx, filter, opts...)
}

func UpdateOne(ctx context.Context, collection string, filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	return db.Collection(collection).UpdateOne(ctx, filter, update)
}

func UpdateMany(ctx context.Context, collection string, filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	return db.Collection(collection).UpdateMany(ctx, filter, update)
}

func DeleteOne(ctx context.Context, collection string, filter bson.M) (*mongo.DeleteResult, error) {
	return db.Collection(collection).DeleteOne(ctx, filter)
}

func DeleteMany(ctx context.Context, collection string, filter bson.M) (*mongo.DeleteResult, error) {
	return db.Collection(collection).DeleteMany(ctx, filter)
}

func CountDocuments(ctx context.Context, collection string, filter bson.M) (int64, error) {
	return db.Collection(collection).CountDocuments(ctx, filter)
}
