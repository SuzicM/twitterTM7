package data

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type ProfileRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

// NoSQL: Constructor which reads db configuration from environment
func New(ctx context.Context, logger *log.Logger) (*ProfileRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &ProfileRepo{
		cli:    client,
		logger: logger,
	}, nil
}

func (ur *ProfileRepo) Disconnect(ctx context.Context) error {
	err := ur.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (ur *ProfileRepo) Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := ur.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		ur.logger.Println(err)
	}

	// Print available databases
	databases, err := ur.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		ur.logger.Println(err)
	}
	fmt.Println(databases)
}

func (ur *ProfileRepo) getCollection() *mongo.Collection {
	userDatabase := ur.cli.Database("user_database")
	usersCollection := userDatabase.Collection("users")
	return usersCollection
}

// NoSQL: Returns all products
func (ur *ProfileRepo) GetAll() (Profiles, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	profilesCollection := ur.getCollection()

	var profiles Profiles
	profilesCursor, err := profilesCollection.Find(ctx, bson.M{})
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	if err = profilesCursor.All(ctx, &profiles); err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	return profiles, nil
}

// NoSQL: Returns Product by id
func (ur *ProfileRepo) Get(id string) (*Profile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	profilesCollection := ur.getCollection()

	var profile Profile
	objID, _ := primitive.ObjectIDFromHex(id)
	err := profilesCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&profile)
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	return &profile, nil
}

func (ur *ProfileRepo) GetByUsername(name string) (Profiles, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	profilesCollection := ur.getCollection()

	var profiles Profiles
	profilesCursor, err := profilesCollection.Find(ctx, bson.M{"username": name})
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	if err = profilesCursor.All(ctx, &profiles); err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	return profiles, nil
}

func (ur *ProfileRepo) Post(profile *Profile) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	profilesCollection := ur.getCollection()

	result, err := profilesCollection.InsertOne(ctx, &profile)
	if err != nil {
		ur.logger.Println(err)
		return err
	}
	ur.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (ur *ProfileRepo) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	profilesCollection := ur.getCollection()

	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}
	result, err := profilesCollection.DeleteOne(ctx, filter)
	if err != nil {
		ur.logger.Println(err)
		return err
	}
	ur.logger.Printf("Documents deleted: %v\n", result.DeletedCount)
	return nil
}
