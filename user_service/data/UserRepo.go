package data

import (
	"context"
	"crypto/rand"

	//"crypto/tls"
	"fmt"
	"log"
	"os"
	"registration/twitterTM7/utils"
	"strconv"
	"time"

	mail "gopkg.in/mail.v2"

	"github.com/gorilla/sessions"
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
	userDatabase := ur.cli.Database("user_database")
	usersCollection := userDatabase.Collection("users")
	return usersCollection
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

func (ur *UserRepo) GetByRegisterCode(code int) (Users, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := ur.getCollection()

	var users Users
	usersCursor, err := usersCollection.Find(ctx, bson.M{"code": code})
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

func (ur *UserRepo) Post(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	usersCollection := ur.getCollection()
	//tempPass := user.Password
	RandomCrypto, _ := rand.Prime(rand.Reader, 42)
	conversionInt := RandomCrypto.String()
	user.RegisterCode, _ = strconv.Atoi(conversionInt)

	//user.Password, _ = HashPassword(user.Password)

	result, err := usersCollection.InsertOne(ctx, &user)
	if err != nil {
		ur.logger.Println(err)
		return err
	}
	ur.SendEmail(user.RegisterCode, user.Email)
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
		"code":      user.RegisterCode,
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

func (ur *UserRepo) NegateCodePut(code int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	usersCollection := ur.getCollection()
	usersFound, err := ur.GetByRegisterCode(code)
	if err != nil {
		ur.logger.Println(err)
		return err
	}
	for _, user := range usersFound {
		user.RegisterCode = 0
		filter := bson.M{"code": code}
		update := bson.M{"$set": bson.M{
			"name":      user.Name,
			"surname":   user.Surname,
			"username":  user.Username,
			"password":  user.Password,
			"age":       user.Age,
			"gender":    user.Gender,
			"residance": user.Residance,
			"code":      user.RegisterCode,
		}}
		result, err := usersCollection.UpdateOne(ctx, filter, update)
		ur.logger.Printf("Documents matched: %v\n", result.MatchedCount)
		ur.logger.Printf("Documents updated: %v\n", result.ModifiedCount)

		if err != nil {
			ur.logger.Println(err)
			return err
		}
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

func (ur *UserRepo) LogInUser(user *SignInData) (string, string, error) {
	logged, err := ur.GetByUsername(user.Username)
	if err != nil {
		ur.logger.Println(err)
		return "", "", err
	}

	key := os.Getenv("ACCESS_TOKEN_PRIVATE_KEY")
	refresh := os.Getenv("REFRESH_TOKEN_PRIVATE_KEY")
	var access_token string
	var refresh_token string

	for _, password := range logged {
		if !CheckPasswordHash(user.Password, password.Password) {
			return "wrong", "wrong", err
		}
		access_token, err = utils.CreateToken(time.Hour, user.Username, key)
		if err != nil {
			ur.logger.Println(err)
			return "", "", err
		}

		refresh_token, err = utils.CreateToken(time.Hour, user.Username, refresh)
		if err != nil {
			ur.logger.Println(err)
			return "", "", err
		}

	}

	return access_token, refresh_token, nil
}

func (ur *UserRepo) GetLoggedUser(s *sessions.Session) string {
	val := s.Values["user"]
	if val == nil {
		return "Empty"
	}
	user := val.(string)
	return user
}

func (ur *UserRepo) SendEmail(code int, email string) error {
	m := mail.NewMessage()
	m.SetHeader("From", "tim7projekat@gmail.com")
	m.SetHeader("To", "markodmaki1999@gmail.com")
	m.SetHeader("Subject", "Registration Confirmation")
	m.SetBody("text/plain", "Hello this is your code: "+strconv.Itoa(code)+"\nGo to the localhost:4200/confirm to confirm your account.")

	d := mail.NewDialer("smtp.gmail.com", 587, "tim7projekat@gmail.com", "qxqcanewprbhydkv")

	if err := d.DialAndSend(m); err != nil {
		panic(err)
		return err
	}
	return nil
}
