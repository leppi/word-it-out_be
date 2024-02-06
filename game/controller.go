package game

import (
  "os"
  "strings"
  "encoding/json"
  "net/http"
  "log"
  "github.com/gorilla/context"
  "github.com/gorilla/sessions"
  "word-it-out/game/service"
  "word-it-out/game/types"
)

type Controller struct {
  Repository *GameRepository
}

func NewController() *Controller {
  repository, err := NewGameRepository()
  if err != nil {
    log.Fatal(err)
  }

  return &Controller{Repository: repository}
}

func handleError(w http.ResponseWriter, err error) {
  http.Error(w, "Virhe :(", http.StatusInternalServerError)
  log.Println(err)
}

// define function for returnin notification json response
func handleNotificationResponse(w http.ResponseWriter, notification types.Notification) {
    // serialize Notification message
    notificationJson, err := json.Marshal(notification)
    if err != nil {
      handleError(w, err)
      return
    }

    // set response headers and return notification json
    w.Header().Set("Content-Type", "application/json")
    w.Write(notificationJson)
}

func (c *Controller) GetGame(w http.ResponseWriter, r *http.Request) {
  // fetch existing session
  session := context.Get(r, "session").(*sessions.Session)

  // fetch game data from session
  game, err := service.GetGameFromSession(session)
  if err != nil {
    handleError(w, err)
    return
  }

  // fetch daily word
  dailyWord, err := c.Repository.GetDailyWord()
  if err != nil {
    handleError(w, err)
    return
  }

  // check if game is currently active
  if service.GameIsActive(game, dailyWord) {
    // check if game is over
    isComplete, isWon := service.GameIsComplete(game)

    // set game data
    game.IsComplete = isComplete
    game.IsWon = isWon
  } else {
    // init new game
    game = types.Game{Guesses: [][]map[string]string{}, IsComplete: false, IsWon: false}

    // replace game data
    if err := service.SetGameToSession(session, game); err != nil {
      handleError(w, err)
      return
    }

    // save session
    if err := session.Save(r, w); err != nil {
      handleError(w, err)
      return
    }
  }

  // create json response
  jsonGameData, err := json.Marshal(game)
  if err != nil {
    handleError(w, err)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  w.Write(jsonGameData)

}

/*
  * PostGuess handles POST request to /guess endpoint
  * It expects a JSON array of strings in the request body
  * The array is the guess of the user
  * It compares the guess to the daily word and returns the result
  * The result is a JSON array of strings
  * The result contains the guess and the result of the comparison
  */
func (c *Controller) PostGuess(w http.ResponseWriter, r *http.Request) {
  // decode post body to json
  decoder := json.NewDecoder(r.Body)

  // create slice of strings to hold the guess
  var guess []string
  // populate guess
  err := decoder.Decode(&guess)

  // handle error
	if err != nil {
		handleError(w, err)
    return
	}

  defer r.Body.Close()

  // fetch existing session
  session := context.Get(r, "session").(*sessions.Session)

  // fetch game data from session
  game, err := service.GetGameFromSession(session)
  if err != nil {
    handleError(w, err)
    return
  }

  // if game is complete, return
  if isComplete, _ := service.GameIsComplete(game); isComplete {
    handleNotificationResponse(w, types.Notification{Type: "success", Message: "Game is complete"})
    return
  }

  if wordExists, _ := c.Repository.WordExists(strings.Join(guess, "")); !wordExists {
    handleNotificationResponse(w, types.Notification{Type: "error", Message: "Word does not exist in sanaseppo"})
    return
  }

  // check word boundaries
  clientNotification, ok := service.CheckWordBoundaries(guess, game)

  if ok {
    // fetch daily word
    dailyWord, err := c.Repository.GetDailyWord()

    if err != nil {
      handleError(w, err)
      return
    }
    
    // compare guess to daily word
    compareResult := service.CompareWord(guess, dailyWord)

    // set game data
    game.Guesses = append(game.Guesses, compareResult)
    game.Guid = dailyWord.Guid
    
    // set session data
    if err := service.SetGameToSession(session, game); err != nil {
      handleError(w, err)
      return
    }

    // save session
    if err := session.Save(r, w); err != nil {
      handleError(w, err)
      return
    }
  }

  handleNotificationResponse(w, clientNotification)
}

func (c *Controller) PostWord(w http.ResponseWriter, r *http.Request) {

  // check secret bearer token
  if r.Header.Get("Authorization") != "Bearer " + os.Getenv("ADMIN_SECRET") {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
  }

  // decode post body to json
  decoder := json.NewDecoder(r.Body)
  var words []string

  err := decoder.Decode(&words)
  if err != nil {
    handleError(w, err)
    return
  }

  defer r.Body.Close()

  log.Println(words)
  // insert words to database
  c.Repository.InsertWords(words)

  w.WriteHeader(http.StatusOK)



  
}

