package Data

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/consul/api"
)

type TweetRepo struct {
	cli    *api.Client
	logger *log.Logger
}

func New(logger *log.Logger) (*TweetRepo, error) {
	db := os.Getenv("DB")
	dbport := os.Getenv("DBPORT")

	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:%s", db, dbport)
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &TweetRepo{
		cli:    client,
		logger: logger,
	}, nil
}

// NoSQL: Returns all products
func (pr *TweetRepo) GetAll() (Tweets, error) {
	kv := pr.cli.KV()
	data, _, err := kv.List(all, nil)
	if err != nil {
		return nil, err
	}

	users := Tweets{}
	for _, pair := range data {
		user := &Tweet{}
		err = json.Unmarshal(pair.Value, user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// NoSQL: Returns Product by id
func (pr *TweetRepo) Get(id string) (*Tweet, error) {
	kv := pr.cli.KV()

	pair, _, err := kv.Get(constructKey(id), nil)
	if err != nil {
		return nil, err
	}
	// If pair is nil -> no object found for given id -> return nil
	if pair == nil {
		return nil, nil
	}

	product := &Tweet{}
	err = json.Unmarshal(pair.Value, product)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (pr *TweetRepo) Post(user *Tweet) (*Tweet, error) {
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
