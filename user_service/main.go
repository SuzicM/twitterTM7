package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"registration/twitterTM7/data"
	"registration/twitterTM7/handlers"
	"syscall"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func main() {
	//Initialize the logger we are going to use, with prefix and datetime for every log
	//As per 12 factor app the general place for app to log is the standard output.
	//If you want to save the logs to a file run the app with the following command.
	//
	//	go run . >> output.txt
	//
	logger := log.New(os.Stdout, "[user-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[user-store] ", log.LstdFlags)

	//Reading from environment, if not set we will default it to 8080.
	//This allows flexibility in different environments (for eg. when running multiple docker api's and want to override the default port)
	port := os.Getenv("USER_SERVICE_PORT")
	if len(port) == 0 {
		port = "8080"
	}

	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the repository that uses the actual database. If the in memory counter part is to be used,
	//swap out the call to 'NewPostgreSql' to 'NewInMemory' and rerun the program.
	store, err := data.New(timeoutContext, storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer store.Disconnect(timeoutContext)

	store.Ping()

	cookies := sessions.NewCookieStore([]byte("super-secret"))
	cookies.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 15,
		HttpOnly: true,
	}
	//Initialize the handler and inject said logger
	usersHandler := handlers.NewUserHandler(logger, store, cookies)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()
	router.Use(usersHandler.MiddlewareContentTypeSet)

	getRouter := router.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/{username}/", usersHandler.GetOneUsername)

	getAllRouter := router.Methods(http.MethodGet).Subrouter()
	getAllRouter.HandleFunc("/all", usersHandler.GetAllUsers)

	getLogged := router.Methods(http.MethodGet).Subrouter()
	getLogged.HandleFunc("/current/user/", usersHandler.GetLogged)

	postRouter := router.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/register/", usersHandler.PostUser)
	postRouter.Use(usersHandler.MiddlewareUserValidation)

	logInRouter := router.Methods(http.MethodPost).Subrouter()
	logInRouter.HandleFunc("/login/", usersHandler.LogInUser)
	logInRouter.Use(usersHandler.MiddlewareDataDeserialization)

	logOutRouter := router.Methods(http.MethodGet).Subrouter()
	logOutRouter.HandleFunc("/u/logout/", usersHandler.LogoutUser)
	

	putRouter := router.Methods(http.MethodPut).Subrouter()
	putRouter.HandleFunc("/{id}/", usersHandler.PutUser)

	deleteHandler := router.Methods(http.MethodDelete).Subrouter()
	deleteHandler.HandleFunc("/user/{id}", usersHandler.DeleteUser)

	//Set cors. Generally you wouldn't like to set cors to a "*". It is a wildcard and it will match any source.
	//Normally you would set this to a set of ip's you want this api to serve. If you have an associated frontend app
	//you would put the ip of the server where the frontend is running. The only time you don't need cors is when you
	//calling the api from the same ip, or when you are using the proxy (for eg. Nginx)
	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"http://localhost:4200"}))
	cors = gorillaHandlers.CORS(gorillaHandlers.AllowCredentials())

	//Initialize the server
	server := http.Server{
		Addr:         ":" + port,        // Addr optionally specifies the TCP address for the server to listen on, in the form "host:port". If empty, ":http" (port 80) is used.
		Handler:      cors(router),      // handler to invoke, http.DefaultServeMux if nil
		IdleTimeout:  120 * time.Second, // IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled.
		ReadTimeout:  5 * time.Second,   // ReadTimeout is the maximum duration for reading the entire request, including the body. A zero or negative value means there will be no timeout.
		WriteTimeout: 5 * time.Second,   // WriteTimeout is the maximum duration before timing out writes of the response.
	}

	logger.Println("Server listening on port", port)
	//Distribute all the connections to goroutines
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Fatal(err)
		}
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT)
	signal.Notify(sigCh, syscall.SIGKILL)

	//When we receive an interrupt or kill, if we don't have any current connections the code will terminate.
	//But if we do the code will stop receiving any new connections and wait for maximum of 30 seconds to finish all current requests.
	//After that the code will terminate.
	sig := <-sigCh
	logger.Println("Received terminate, graceful shutdown", sig)

	//Try to shutdown gracefully
	if server.Shutdown(timeoutContext) != nil {
		logger.Fatal("Cannot gracefully shutdown...")
	}
}
