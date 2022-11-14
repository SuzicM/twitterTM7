package data

import (
	"fmt"

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
