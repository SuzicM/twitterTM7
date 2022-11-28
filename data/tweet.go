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
	Username   string `json:"username"`
	TweetTitle string `json:"title"`
	TweetBody  string `json:"body"`
	CreatedOn  gocql.UUID
}

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

func (o *TweetByUsername) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
