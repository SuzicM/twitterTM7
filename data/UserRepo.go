package data

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/consul/api"
)

type UserRepo struct {
	cli    *api.Client
	logger *log.Logger
}

// NoSQL: Constructor which reads db configuration from environment
func New(logger *log.Logger) (*UserRepo, error) {
	db := os.Getenv("DB")
	dbport := os.Getenv("DBPORT")

	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%s", db, dbport)
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &UserRepo{
		cli:    client,
		logger: logger,
	}, nil
}

// NoSQL: Returns all products
func (pr *UserRepo) GetAll() (Users, error) {
	kv := pr.cli.KV()
	data, _, err := kv.List(all, nil)
	if err != nil {
		return nil, err
	}

	users := Users{}
	for _, pair := range data {
		user := &User{}
		err = json.Unmarshal(pair.Value, user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// NoSQL: Returns Product by id
func (pr *UserRepo) Get(id string) (*User, error) {
	kv := pr.cli.KV()

	pair, _, err := kv.Get(constructKey(id), nil)
	if err != nil {
		return nil, err
	}
	// If pair is nil -> no object found for given id -> return nil
	if pair == nil {
		return nil, nil
	}

	product := &User{}
	err = json.Unmarshal(pair.Value, product)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (pr *UserRepo) Post(user *User) (*User, error) {
	kv := pr.cli.KV()

	user.CreatedOn = time.Now().UTC().String()
	user.UpdatedOn = time.Now().UTC().String()

	dbId, id := generateKey()
	user.ID = id

	data, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	userKeyValue := &api.KVPair{Key: dbId, Value: data}
	_, err = kv.Put(userKeyValue, nil)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (pr *UserRepo) Put(id string, user *User) (*User, error) {
	kv := pr.cli.KV()

	user.UpdatedOn = time.Now().UTC().String()

	data, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	userKeyValue := &api.KVPair{Key: constructKey(id), Value: data}
	_, err = kv.Put(userKeyValue, nil)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (pr *UserRepo) Delete(id string) error {
	kv := pr.cli.KV()

	_, err := kv.Delete(constructKey(id), nil)
	if err != nil {
		return err
	}

	return nil
}
