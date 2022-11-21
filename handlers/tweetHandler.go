package handlers

import (
	_ "context"
	"encoding/json"
	"log"
	"net/http"
	"twitterTM7/data"

	"github.com/gorilla/mux"
)

type KeyProduct struct{}

type tweetHandler struct {
	logger *log.Logger
	// NoSQL: injecting user repository
	repo *data.TweetRepo
}

// Injecting the logger makes this code much more testable.
func NewTweetsHandler(l *log.Logger, r *data.TweetRepo) *tweetHandler {
	return &tweetHandler{l, r}
}

func (s *tweetHandler) GetAllTweetIds(rw http.ResponseWriter, h *http.Request) {
	tweetIds, err := s.repo.GetDistinctIds("user_id", "tweets_by_user")
	if err != nil {
		s.logger.Print("Database exception: ", err)
	}

	if tweetIds == nil {
		return
	}

	s.logger.Println(tweetIds)

	e := json.NewEncoder(rw)
	err = e.Encode(tweetIds)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		s.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (s *tweetHandler) GetTweetsByUser(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	userId := vars["id"]

	tweetsByUser, err := s.repo.GetTweetsByUser(userId)
	if err != nil {
		s.logger.Print("Database exception: ", err)
	}

	if tweetsByUser == nil {
		return
	}

	err = tweetsByUser.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		s.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (s *tweetHandler) CreateTweetForUser(rw http.ResponseWriter, h *http.Request) {
	userTweet := h.Context().Value(KeyProduct{}).(*data.TweetByUser)
	err := s.repo.InsertTweetByUser(userTweet)
	if err != nil {
		s.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}
