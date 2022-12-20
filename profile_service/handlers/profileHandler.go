package handlers

import (
	"log"
	"net/http"

	"registration/twitterTM7/client/tweet"
	"registration/twitterTM7/client/user"
	"registration/twitterTM7/data"

	"github.com/gorilla/mux"
)

type KeyProfile struct{}

type ProfileHandler struct {
	logger      *log.Logger
	repo        *data.ProfileRepo
	tweetClient tweet.Client
	userClient  user.Client
}

func NewProfileHandler(d *data.ProfileRepo, tc tweet.Client, uc user.Client, l *log.Logger) *ProfileHandler {
	return &ProfileHandler{
		repo:        d,
		tweetClient: tc,
		userClient:  uc,
		logger:      l,
	}
}

func (handler *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username, ok := vars["username"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := handler.userClient.GetUser(username)
	if err != nil {
		http.Error(w, "Unable to get user from that username ", http.StatusInternalServerError)
		handler.logger.Fatal("Unable to get user from that username :", err)
		return
	}

	tweet, err := handler.tweetClient.GetTweet(username)
	if err != nil {
		http.Error(w, "Unable to get tweets from that username ", http.StatusInternalServerError)
		handler.logger.Fatal("Unable to get tweets from that username :", err)
		return
	}

	profile, err := handler.repo.GetByUsername(username)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if profile == nil {
		var tempProfile *data.Profile
		tempProfile.Tweets = tweet
		tempProfile.User = user

		err = handler.repo.Post(tempProfile)
		if err != nil {
			http.Error(w, "Unable to save profile ", http.StatusInternalServerError)
			handler.logger.Fatal("Unable to save profile :", err)
			return
		}
		profile, err = handler.repo.GetByUsername(username)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	jsonResponse(profile, w)
}

/*func (handler *ProfileHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	orders, err := handler.service.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonResponse(orders, w)
}*/

func (p *ProfileHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		p.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}
