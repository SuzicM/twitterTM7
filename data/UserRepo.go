package data

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type UserRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

// NoSQL: Constructor which reads db configuration from environment
func New(ctx context.Context, logger *log.Logger) (*UserRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &UserRepo{
		cli:    client,
		logger: logger,
	}, nil
}

func (ur *UserRepo) Disconnect(ctx context.Context) error {
	err := ur.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (ur *UserRepo) Ping() {
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

func (ur *UserRepo) getCollection() *mongo.Collection {
	patientDatabase := ur.cli.Database("mongoDemo")
	patientsCollection := patientDatabase.Collection("patients")
	return patientsCollection
}

// NoSQL: Returns all products
func (ur *UserRepo) GetAll() (Users, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := ur.getCollection()

	var users Users
	patientsCursor, err := usersCollection.Find(ctx, bson.M{})
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	if err = patientsCursor.All(ctx, &users); err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	return users, nil
}

// NoSQL: Returns Product by id
func (ur *UserRepo) Get(id string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := ur.getCollection()

	var user User
	objID, _ := primitive.ObjectIDFromHex(id)
	err := usersCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepo) GetByUsername(name string) (Users, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := ur.getCollection()

	var users Users
	usersCursor, err := usersCollection.Find(ctx, bson.M{"username": name})
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	if err = usersCursor.All(ctx, &users); err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	return users, nil
}

func (ur *UserRepo) GetAllTweetsByUser(username string) (Tweets, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := ur.getCollection()

	var tweets Tweets
	tweetsCursor, err := usersCollection.Find(ctx, bson.M{"usernametw": username})
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	if err = tweetsCursor.All(ctx, &tweets); err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	return tweets, nil
}

func (ur *UserRepo) GetUserProfile(id string) (*User, Tweets, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := ur.getCollection()

	var user User
	objID, _ := primitive.ObjectIDFromHex(id)
	err := usersCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		ur.logger.Println(err)
		return nil, nil, err
	}

	tweets, err := ur.GetAllTweetsByUser(user.Username)
	if err != nil {
		ur.logger.Println(err)
		return nil, nil, err
	}

	return &user, tweets, nil
}

func (ur *UserRepo) Post(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	usersCollection := ur.getCollection()
	tempPass := user.Password

	user.Password, _ = HashPassword(tempPass)

	result, err := usersCollection.InsertOne(ctx, &user)
	if err != nil {
		ur.logger.Println(err)
		return err
	}
	ur.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (ur *UserRepo) PostTweet(tweet *Tweet) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	usersCollection := ur.getCollection()

	result, err := usersCollection.InsertOne(ctx, &tweet)
	if err != nil {
		ur.logger.Println(err)
		return err
	}
	ur.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (ur *UserRepo) Put(id string, user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	usersCollection := ur.getCollection()

	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{
		"name":      user.Name,
		"surname":   user.Surname,
		"username":  user.Username,
		"password":  user.Password,
		"age":       user.Age,
		"gender":    user.Gender,
		"residance": user.Residance,
	}}
	result, err := usersCollection.UpdateOne(ctx, filter, update)
	ur.logger.Printf("Documents matched: %v\n", result.MatchedCount)
	ur.logger.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		ur.logger.Println(err)
		return err
	}
	return nil
}

func (ur *UserRepo) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	usersCollection := ur.getCollection()

	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}
	result, err := usersCollection.DeleteOne(ctx, filter)
	if err != nil {
		ur.logger.Println(err)
		return err
	}
	ur.logger.Printf("Documents deleted: %v\n", result.DeletedCount)
	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func IsAlnumOrHyphen(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

func ValidatePassword(s string) bool {
	pass := 0
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			pass++
		case unicode.IsUpper(c):
			pass++
		case unicode.IsPunct(c):
			pass++
		case unicode.IsLower(c):
			pass++
		case unicode.IsLetter(c) || c == ' ':
			pass++
		default:
			return false
		}
	}
	return pass == len(s)
}

func ValidateName(user *User) bool {
	reg, _ := regexp.Compile("^[a-zA-Z]+$")
	match := reg.MatchString(user.Name)
	return match
}

func ValidateLastName(user *User) bool {
	reg, _ := regexp.Compile("^[a-zA-Z]+$")
	match := reg.MatchString(user.Surname)
	return match
}

func ValidateGender(user *User) bool {
	reg, _ := regexp.Compile("^[a-zA-Z]+$")
	match := reg.MatchString(user.Gender)
	return match
}

func ValidateResidance(user *User) bool {
	reg, _ := regexp.Compile("^[a-z]+([a-zA-Z0-9]+)$")
	match := reg.MatchString(user.Gender)
	return match
}

func ValidateAge(user *User) bool {
	reg, _ := regexp.Compile("^[0-9]+$")
	match := reg.MatchString(user.Age)
	return match
}

func ValidateUsername(user *User) bool {
	return IsAlnumOrHyphen(user.Username)
}

func (ur *UserRepo) ValidateUser(user *User) bool {
	if !ValidateAge(user) {
		return false
	}
	if !ValidateUsername(user) {
		return false
	}
	if !ValidateName(user) {
		return false
	}
	if !ValidateLastName(user) {
		return false
	}
	if !ValidateGender(user) {
		return false
	}
	if !ValidateResidance(user) {
		return false
	}
	if !ValidatePassword(user.Password) {
		return false
	}
	return true
}
