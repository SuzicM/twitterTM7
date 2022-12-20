package data

import (
	"encoding/json"
	"io"

	"github.com/gocql/gocql"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"registration/twitterTM7/client/tweet"
	"registration/twitterTM7/client/user"
)

// Defining the main struct for our API
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name" validate:"required"`
	Surname   string             `bson:"surname" json:"surname" validate:"required"`
	Username  string             `bson:"username" json:"username" validate:"required"`
	Password  string             `bson:"password" json:"password" validate:"required"`
	Age       string             `bson:"age" json:"age" validate:"required"`
	Gender    string             `bson:"gender" json:"gender" validate:"required"`
	Residance string             `bson:"residance" json:"residance" validate:"required"`
}

type Tweet struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TweetID  gocql.UUID         `bson:"tweetid" json:"tweetid"`
	Username string             `bson:"usernametw" json:"usernametw"`
	Body     string             `bson:"body" json:"body"`
}

type TweetByUsername struct {
	Username   string `json:"username"`
	TweetTitle string `json:"title"`
	TweetBody  string `json:"body"`
	CreatedOn  gocql.UUID
}

type Tweets tweet.TweetsByUsername

type Profile struct {
	User   user.Users
	Tweets tweet.TweetsByUsername
}

type Users []*User

type Profiles []*Profile

func (p *Tweets) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func (p *Tweet) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func (p *Tweet) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(p)
}

func (p *Users) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func (p *User) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func (p *User) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(p)
}

func (p *Profile) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}
