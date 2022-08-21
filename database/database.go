package database

import (
	"context"
	"errors"
	"log"
	"os"
	"time"
	"github.com/ayo-ajayi/proj/model"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	client *mongo.Client
}

func Connect() *DB {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	mongo_url := os.Getenv("MONGODB_URL")
	client, err := mongo.NewClient(options.Client().ApplyURI(mongo_url))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	if err:= client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	return &DB{
		client: client,
	}
}

var db = Connect()

func CreateUser(entry model.User) (primitive.ObjectID, error) {
	collection := db.client.Database("jwt").Collection("new")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	count, err := collection.CountDocuments(ctx, bson.M{"username": entry.Username})
	if count != 0 {
		return primitive.ObjectID{}, errors.New("user already exists")
	}
	if err != nil {
		return primitive.ObjectID{}, err
	}
	res, err := collection.InsertOne(ctx, bson.M{
		"password": entry.Password,
		"phone":    entry.Phone,
		"username": entry.Username,
	})
	if err != nil {
		log.Fatal(err)
		return primitive.ObjectID{}, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

func FindUser(entry model.User) (model.User, error) {
	collection := db.client.Database("jwt").Collection("new")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	count, err := collection.CountDocuments(ctx, bson.M{"username": entry.Username})
	if count == 0 {
		return model.User{}, errors.New("user doesn't exist")
	}
	if err != nil {
		return model.User{}, err
	}
	user := model.User{}
	err = collection.FindOne(ctx, bson.M{"username": entry.Username}).Decode(&user)
	if err != nil {
		log.Fatal(err)
	}
	return user, nil
}

func CreateTodo(entry model.Todo) error {
	collection := db.client.Database("jwt").Collection("new")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := collection.InsertOne(ctx, bson.M{
		"title":   entry.Title,
		"user_id": entry.UserID,
	}); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func FindUserTodos(entry primitive.ObjectID) ([]primitive.M, error) {
	collection := db.client.Database("jwt").Collection("new")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	opts := options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}).SetProjection(bson.D{{Key: "user_id", Value: 0}, {Key: "_id", Value: 0}})
	cursor, err := collection.Find(ctx, bson.M{"user_id": entry}, opts)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
		return nil, err
	}
	
	return results, nil
}

