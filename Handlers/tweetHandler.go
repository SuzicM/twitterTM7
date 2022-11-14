package Handlers

import (
	_ "context"
	_ "fmt"
	"log"
	"net/http"
	"twitterTM7/Data"

	"github.com/gorilla/mux"
)

type KeyTweet struct{}

type TweetHandler struct {
	logger *log.Logger
	repo   *Data.TweetRepo
}

func NewTweetHandler(l *log.Logger, r *Data.TweetRepo) *TweetHandler {
	return &TweetHandler{l, r}
}

// Gets all of the users in the database
func (p *TweetHandler) GetAllTweets(rw http.ResponseWriter, h *http.Request) {
	allTweets, err := p.repo.GetAll()
	if err != nil {
		http.Error(rw, "Database exception", http.StatusInternalServerError)
		p.logger.Fatal("Database exception: ", err)
	}

	err = allTweets.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		p.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (p *TweetHandler) GetOneTweet(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	tweet, err := p.repo.Get(id)
	if err != nil {
		http.Error(rw, "Database exception", http.StatusInternalServerError)
		p.logger.Fatal("Database exception: ", err)
	}

	if tweet == nil {
		http.Error(rw, "User with given id not found", http.StatusNotFound)
		p.logger.Printf("User with id: '%s' not found", id)
		return
	}

	err = tweet.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		p.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (p *TweetHandler) PostTweet(rw http.ResponseWriter, h *http.Request) {
	user := h.Context().Value(KeyTweet{}).(*Data.Tweet)
	p.repo.Post(user)
	rw.WriteHeader(http.StatusCreated)
}
