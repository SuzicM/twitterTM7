package data

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
)

const (
	users = "users/%s"
	all   = "users"
)

func generateKey() (string, string) {
	id := uuid.New().String()
	return fmt.Sprintf(users, id), id
}

func constructKey(id string) string {
	return fmt.Sprintf(users, id)
}

func decodeBodyGroup(r io.Reader) (*Tweet, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt *Tweet
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}
	return rt, nil
}
