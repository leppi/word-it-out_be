package app

import (
  "log"
  "net/http"
  "github.com/gorilla/mux"
  "github.com/joho/godotenv"
  "github.com/gorilla/context"
  "github.com/gorilla/sessions"
  "os"
  "word-it-out/game"
)

func CreateSessionStore() *sessions.CookieStore {
  store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
  store.Options = &sessions.Options{
    Path:     "/",
    MaxAge:   86400 * 7, // 7 days
    HttpOnly: true,
  }

  return store
}
type App struct {
  Router *mux.Router
}

func (a *App) Initialize() {
  // Load environment variables
  err := godotenv.Load()

  // Check if environment variables are loaded
  if err != nil {
    log.Fatal("Error loading .env file")
  }

  // Initialize router
  a.Router = mux.NewRouter()
}

// Run starts the application
func (a *App) Run() {
  // Initialize app
  a.Initialize()

  // create controller
  gameController := game.NewController()

  // set up middlewares
  a.Router.Use(corsMiddleware)
  a.Router.Use(sessionMiddleware)

  // set up routes
  a.Router.HandleFunc("/word", gameController.PostWord).Methods("POST", "OPTIONS")
  a.Router.HandleFunc("/guess", gameController.PostGuess).Methods("POST", "OPTIONS")
  a.Router.HandleFunc("/game", gameController.GetGame).Methods("GET", "OPTIONS")
  a.Router.HandleFunc("/debug", gameController.GetDebug).Methods("GET")

  // check if certificate files are set
  if os.Getenv("CERT_DIR") != "" && os.Getenv("CERT_FILE") != "" && os.Getenv("KEY_FILE") != "" {
    // Start secure server
    log.Print("Starting secure server on port " + os.Getenv("PORT"))
    log.Fatal(http.ListenAndServeTLS(
      ":" + os.Getenv("PORT"),
      os.Getenv("CERT_DIR") + os.Getenv("CERT_FILE"),
      os.Getenv("CERT_DIR") + os.Getenv("KEY_FILE"),
      a.Router,
    ))
  } else {
    // Start server
    log.Print("Starting server on port " + os.Getenv("PORT"))
    log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), a.Router))
  }
}


func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", os.Getenv("CLIENT_URL"))
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
    w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

func sessionMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Fetch session
		session, err := CreateSessionStore().Get(r, os.Getenv("SESSION_NAME"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}


    // set session to request context
		context.Set(r, "session", session)

    // save session
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

    // Continue to next middleware or handler
		next.ServeHTTP(w, r)
	})
}

