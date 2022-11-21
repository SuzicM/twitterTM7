package data

import (
	"encoding/json"
	"github.com/gocql/gocql"
	"io"
)

type TweetByUser struct {
	UserId     gocql.UUID
	TweetTitle string
	TweetBody  string
}

type TweetsByUser []*TweetByUser

func (o *TweetsByUser) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(o)
}

func (o *TweetByUser) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(o)
}
