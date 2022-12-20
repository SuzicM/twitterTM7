package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type Client struct {
	address string
}

func NewClient(host, port string) Client {
	return Client{
		address: fmt.Sprintf("http://%s:%s/", host, port),
	}
}

func (client Client) GetUser(username string) (Users, error) {
	reqBytes, err := json.Marshal(UserRequest{username})
	if err != nil {
		return nil, err
	}

	bodyReader := bytes.NewReader(reqBytes)
	requestURL := client.address + username + "/"
	log.Println("send to this address: " + requestURL)
	httpReq, err := http.NewRequest(http.MethodGet, requestURL, bodyReader)

	if err != nil {
		log.Println(err)
		return nil, errors.New("error getting user")
	}

	res, err := http.DefaultClient.Do(httpReq)

	if err != nil || res.StatusCode != http.StatusOK {
		log.Println(err)
		log.Println(res.StatusCode)
		return nil, errors.New("error getting user.")
	}

	var usertemp Users
	json.NewDecoder(res.Body).Decode(usertemp)

	log.Println(json.NewDecoder(res.Body).Decode(usertemp))

	/*var users Users
	for _, user :=  range *usertemp{
		users = append(users, user)
	}*/

	return usertemp, nil
}
