package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"registration/twitterTM7/data"
	"registration/twitterTM7/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type KeyUser struct{}

type UserHandler struct {
	logger *log.Logger
	repo   *data.UserRepo
	store  *sessions.CookieStore
}

func NewUserHandler(l *log.Logger, r *data.UserRepo, s *sessions.CookieStore) *UserHandler {
	return &UserHandler{l, r, s}
}

//Gets all of the users in the database
func (p *UserHandler) GetAllUsers(rw http.ResponseWriter, h *http.Request) {
	allUsers, err := p.repo.GetAll()
	if err != nil {
		http.Error(rw, "Database exception", http.StatusInternalServerError)
		p.logger.Fatal("Database exception: ", err)
	}

	if allUsers == nil {
		return
	}

	err = allUsers.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		p.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (p *UserHandler) GetOneUser(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	user, err := p.repo.Get(id)
	if err != nil {
		http.Error(rw, "Database exception", http.StatusInternalServerError)
		p.logger.Fatal("Database exception: ", err)
	}

	if user == nil {
		http.Error(rw, "User with given id not found", http.StatusNotFound)
		p.logger.Printf("User with id: '%s' not found", id)
		return
	}

	err = user.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		p.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (p *UserHandler) PostUser(rw http.ResponseWriter, h *http.Request) {
	user := h.Context().Value(KeyUser{}).(*data.User)
	p.repo.Post(user)
	rw.WriteHeader(http.StatusCreated)
}

func (p *UserHandler) PutUser(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	user := h.Context().Value(KeyUser{}).(*data.User)

	p.repo.Put(id, user)
	rw.WriteHeader(http.StatusOK)
}

func (p *UserHandler) DeleteUser(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	err := p.repo.Delete(id)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		p.logger.Fatal("Unable to delete user.", err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (p *UserHandler) LogInUser(rw http.ResponseWriter, h *http.Request) {
	session, err := p.store.Get(h, "super-secret")
	if err != nil {
		http.Error(rw, "Unable to get cookie store", http.StatusInternalServerError)
		p.logger.Fatal("Unable to get cookie store :", err)
		return
	}

	user := h.Context().Value(KeyUser{}).(*data.SignInData)
	atoken, rtoken, err := p.repo.LogInUser(user)
	if err != nil {
		http.Error(rw, "Unable to log in", http.StatusInternalServerError)
		p.logger.Fatal("Unable to log in :", err)
		return
	}
	if (atoken == "wrong") || (rtoken == "wrong") {
		http.Error(rw, "Invalid username or password: ", http.StatusInternalServerError)
		p.logger.Fatal("Invalid username or password: ", err)
		return
	}

	session.Values["access_token"] = atoken
	session.Values["refresh_token"] = rtoken
	session.Values["auth"] = "true"
	session.Values["user"] = user.Username
	session.Options.MaxAge = 1800

	err = session.Save(h, rw)
	if err != nil {
		http.Error(rw, "Unable to save cookies", http.StatusInternalServerError)
		p.logger.Fatal("Unable to save cookies :", err)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (p *UserHandler) LogoutUser(rw http.ResponseWriter, h *http.Request) {
	session, err := p.store.Get(h, "super-secret")
	if err != nil {
		http.Error(rw, "Unable to get cookies", http.StatusInternalServerError)
		p.logger.Fatal("Unable to get cookies :", err)
		return
	}
	session.Values["access_token"] = ""
	session.Values["refresh_token"] = ""
	session.Values["auth"] = "false"
	session.Values["user"] = ""
	session.Options.MaxAge = -1
	err = session.Save(h, rw)
	if err != nil {
		http.Error(rw, "Unable to save cookies", http.StatusInternalServerError)
		p.logger.Fatal("Unable to save cookies :", err)
		return
	}
	//http.Redirect(rw, h, "/", http.StatusFound)
	rw.WriteHeader(http.StatusOK)
}

func (p *UserHandler) GetLogged(rw http.ResponseWriter, h *http.Request) {
	session, err := p.store.Get(h, "super-secret")
	if err != nil {
		http.Error(rw, "Unable to get cookies", http.StatusInternalServerError)
		p.logger.Fatal("Unable to get cookies :", err)
		return
	}

	username := p.repo.GetLoggedUser(session)

	user, _ := p.repo.GetByUsername(username)

	err = user.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		p.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

//Middleware to try and decode the incoming body. When decoded we run the validation on it just to check if everything is okay
//with the model. If anything is wrong we terminate the execution and the code won't even hit the handler functions.
//With a key we bind what we read to the context of the current request. Later we use that key to get to the read value.

func (p *UserHandler) MiddlewareUserValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		user := &data.User{}
		err := user.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			p.logger.Fatal(err)
			return
		}

		if !p.repo.ValidateUser(user) {
			p.logger.Println("Error: Some of the input values for user data are not correct")
			http.Error(rw, fmt.Sprintf("Error: Some of the input values for user data are not correct"), http.StatusBadRequest)
			return
		}

		users, err := p.repo.GetByUsername(user.Username)

		if users != nil {
			p.logger.Println("Error: username exists", err)
			http.Error(rw, fmt.Sprintf("Error: username exits, %s", err), http.StatusBadRequest)
			return
		}

		user.Password, _ = data.HashPassword(user.Password)

		ctx := context.WithValue(h.Context(), KeyUser{}, user)
		h = h.WithContext(ctx)

		next.ServeHTTP(rw, h)
	})
}

//Middleware to centralize general logging and to add the header values for all requests.

func (p *UserHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		p.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		rw.Header().Add("Content-Type", "application/json")

		next.ServeHTTP(rw, h)
	})
}

func (p *UserHandler) DeserializeUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var access_token string
		cookie, err := ctx.Cookie("access_token")

		authorizationHeader := ctx.Request.Header.Get("Authorization")
		fields := strings.Fields(authorizationHeader)

		if len(fields) != 0 && fields[0] == "Bearer" {
			access_token = fields[1]
		} else if err == nil {
			access_token = cookie
		}

		if access_token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "You are not logged in"})
			return
		}

		config, _ := utils.LoadConfig(".")
		sub, err := utils.ValidateToken(access_token, config.AccessTokenPublicKey)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": err.Error()})
			return
		}

		user, err := p.repo.GetByUsername(fmt.Sprint(sub))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "The user belonging to this token no logger exists"})
			return
		}

		ctx.Set("currentUser", user)
		ctx.Next()
	}
}

func (s *UserHandler) MiddlewareDataDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		user := &data.SignInData{}
		err := user.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			s.logger.Fatal(err)
			return
		}
		ctx := context.WithValue(h.Context(), KeyUser{}, user)
		h = h.WithContext(ctx)
		next.ServeHTTP(rw, h)
	})
}
