package handlers

import (
	"context"
	_ "context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"twitterTM7/data"

	"github.com/gorilla/mux"
)

type KeyProduct struct{}

type TweetHandler struct {
	logger *log.Logger
	// NoSQL: injecting user repository
	repo *data.TweetRepo
}

// Injecting the logger makes this code much more testable.
func NewTweetsHandler(l *log.Logger, r *data.TweetRepo) *TweetHandler {
	return &TweetHandler{l, r}
}

func (s *TweetHandler) GetAllTweetIds(rw http.ResponseWriter, h *http.Request) {
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

func (s *TweetHandler) GetAllTweetUsernames(rw http.ResponseWriter, h *http.Request) {
	tweetIds, err := s.repo.GetDistinctIds("username", "tweets_by_username")
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

func (s *TweetHandler) GetTweetsByUser(rw http.ResponseWriter, h *http.Request) {
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

func (s *TweetHandler) GetTweetsByUsername(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	username := vars["username"]

	tweetsByUsername, err := s.repo.GetTweetsByUsername(username)
	if err != nil {
		s.logger.Print("Database exception: ", err)
	}

	if tweetsByUsername == nil {
		return
	}

	err = tweetsByUsername.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		s.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (s *TweetHandler) CreateTweetForUser(rw http.ResponseWriter, h *http.Request) {
	userTweet := h.Context().Value(KeyProduct{}).(*data.TweetByUser)
	userTweetChanged := userTweet.TweetBody
	userTweetChanged = strings.ReplaceAll(userTweetChanged, "<", "i16")
	userTweetChanged = strings.ReplaceAll(userTweetChanged, ">", "i12")
	userTweet.TweetBody = userTweetChanged
	err := s.repo.InsertTweetByUser(userTweet)
	if err != nil {
		s.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

func (s *TweetHandler) CreateTweetForUsername(rw http.ResponseWriter, h *http.Request) {
	userTweet := h.Context().Value(KeyProduct{}).(*data.TweetByUsername)
	userTweetChanged := userTweet.TweetBody
	userTweetChanged = strings.ReplaceAll(userTweetChanged, "<", "i16")
	userTweetChanged = strings.ReplaceAll(userTweetChanged, ">", "i12")
	userTweet.TweetBody = userTweetChanged
	err := s.repo.InsertTweetByUsername(userTweet)
	if err != nil {
		s.logger.Print("Database exception: ", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}

func (s *TweetHandler) MiddlewareTweetsForUserDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		tweetByUser := &data.TweetByUser{}
		err := tweetByUser.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			s.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProduct{}, tweetByUser)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}

func (s *TweetHandler) MiddlewareTweetsForUsernameDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		tweetByUsername := &data.TweetByUsername{}
		err := tweetByUsername.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			s.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyProduct{}, tweetByUsername)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}

func (s *TweetHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		s.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}
