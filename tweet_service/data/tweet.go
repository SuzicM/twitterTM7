package data

import (
	"encoding/json"
	"io"

	"github.com/gocql/gocql"
)

type TweetByUser struct {
	UserId     gocql.UUID
	TweetTitle string `json:"title"`
	TweetBody  string `json:"body"`
	CreatedOn  gocql.UUID
}

type TweetByUsername struct {
	Username   string     `json:"username"`
	TweetTitle string     `json:"title"`
	TweetBody  string     `json:"body"`
	CreatedOn  gocql.UUID `json:"tweetid"`
}

type Like struct {
	Username string     `json:"username"`
	TweetId  gocql.UUID `json:"tweetid"`
	Liked    bool       `json:"liked"`
}

type TLikes []*Like

type Likes struct {
	NumberOfLikes int `json:"likes"`
}

type UsersLiked []string

type TweetsByUser []*TweetByUser

type TweetsByUsername []*TweetByUsername

func (o *TweetsByUser) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *TweetByUser) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

func (o *TweetsByUsername) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *UsersLiked) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *Likes) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *TweetByUsername) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

func (o *Like) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

func (o *Likes) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

func (o *TLikes) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}

func (o *UsersLiked) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
