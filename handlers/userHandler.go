package handlers

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"log"
	"net/http"
	"registration/twitterTM7/data"
)

type KeyUser struct{}

type UserHandler struct {
	logger *log.Logger
	repo   *data.UserRepo
}

func NewUserHandler(l *log.Logger, r *data.UserRepo) *UserHandler {
	return &UserHandler{l, r}
}

// Gets all of the users in the database
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

func (p *UserHandler) GetUsername(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	username := vars["username"]

	user, err := p.repo.Get(username)
	if err != nil {
		http.Error(rw, "Database exception", http.StatusInternalServerError)
		p.logger.Fatal("Database exception: ", err)
	}
	if user == nil {
		http.Error(rw, "User with given username not found", http.StatusNotFound)
		p.logger.Printf("User with username: '%s' not found", username)
		return
	}
	err = user.ToJSON(rw)
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

		users, err := p.repo.GetByUsername(user.Username)

		if users != nil {
			p.logger.Println("Error: username exists", err)
			http.Error(rw, fmt.Sprintf("Error: username exits, %s", err), http.StatusBadRequest)
			return
		}

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

func (p *UserHandler) LoginUser(rw http.ResponseWriter, h *http.Request) {
	name := h.FormValue("name")
	pass := h.FormValue("password")
	redirectTarget := "/"
	if name != "" && pass != "" {
		/*users, err := p.repo.GetByUsername(name)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			p.logger.Fatal("Unable to delete user.", err)
			return
		}
		*/

		setSession(name, rw)
		redirectTarget = "/internal"
	}
	http.Redirect(rw, h, redirectTarget, 302)
}

func (p *UserHandler) LogoutUser(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, "/", 302)
}

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func (p *UserHandler) getUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

func setSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}
