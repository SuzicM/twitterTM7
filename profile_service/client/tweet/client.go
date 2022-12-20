package tweet

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
		address: fmt.Sprintf("http://%s:%s/username/", host, port),
	}
}

func (client Client) GetTweet(username string) (TweetsByUsername, error) {
	reqBytes, err := json.Marshal(TweetRequest{username})
	if err != nil {
		return nil, err
	}

	bodyReader := bytes.NewReader(reqBytes)
	requestURL := client.address + username + "/"
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

	var tweettemp TweetsByUsername
	json.NewDecoder(res.Body).Decode(tweettemp)

	/*var tweets TweetsByUsername
	for _, tweet :=  range *tweettemp{
		tweets = append(tweets, tweet)
	}*/

	return tweettemp, nil
}
