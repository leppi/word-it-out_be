package service

import (
  "strings"
  "encoding/json"
  "word-it-out/game/types"
  "github.com/gorilla/sessions"
  "fmt"
  "time"
  "github.com/google/uuid"
)

// create enums for correct, found, missed
const (
  CORRECT = "correct"
  FOUND = "found"
  MISSED = "missed"
  GAME_LENGTH = 6
)

func generateUUID() string {
  id, err := uuid.NewRandom()
  if err != nil {
    fmt.Println("Error generating UUID:", err)
    return ""
  }
  return id.String()
}


// internal function to get key and value from map
func getKeyAndValueFromMap(letterMap []string) (string, string) {
  return letterMap[0], letterMap[1]
}

// create a word from guess mapping
func concatenateKeys(previousGuess [][]string) string {
	var keys []string
	for _, letterMap := range previousGuess {
    keys = append(keys, letterMap[0])
	}
	return strings.Join(keys, "")
}

func getCharMap(word types.Word) map[rune]int {
  charMap := make(map[rune]int)
  for _, char := range word.Word {
    charMap[char]++
  }
  return charMap
}

func CompareWord(guess []string, dailyWord types.Word) [][]string {
  // split daily word into array of letters
  dailyWordRunes := []rune(dailyWord.Word)

  // get char map of daily word
  charMap := getCharMap(dailyWord)
  var result [][]string

  // first mark the correct letters. Then mark the found letters. 
  for i, letter := range guess {
    entry := make([]string, 2)
    key := string(letter)
    runeKey := []rune(key)[0]

    // match each letter in guess to daily word
    // set result key as guess letter and value as "correct" or "found" or "missed"
    if dailyWordRunes[i] == runeKey {
      entry = []string{key, CORRECT}
      charMap[runeKey]--

    } else {
      entry = []string{key, MISSED}
    }

    result = append(result, entry)
  }

  
  // now loop result and set found letters
  // we need to iterate over the guess twice because letter can be found first and then be correct in the same guess
  for i, entry := range result {
    key := entry[0]
    value := entry[1]
    // make sure runeKey is in utf8 format
    runeKey := []rune(key)[0]
    if value == MISSED && strings.ContainsRune(dailyWord.Word, runeKey) && charMap[runeKey] > 0 {
      result[i][1] = FOUND
      charMap[runeKey]--
    }
  }

  return result
}

func GameIsActive(game types.Game, dailyWord types.Word) bool {
  return game.Guid == dailyWord.Guid
}

func SetGameToSession(session *sessions.Session, game types.Game) error {
  // create json data
  jsonGameData, err := json.Marshal(game)
  if err != nil {
    return err
  }

  // create session id if it doesn't exist
  if session.Values["id"] == nil {
    session.Values["id"] = generateUUID()
  }

  // set session data
  session.Values["gamedata"] = string(jsonGameData)

  return nil
}

func GameIsTooOld(game types.Game, dailyWord types.Word) bool {
  // validate date
  if !dailyWord.UsedAt.Valid || game.UsedAt == "" {
    return true
  }
  // check if date string day difference is over 1
  layout := "2006-01-02"
  gameDate, err := time.Parse(layout, game.UsedAt)
	if err != nil {
		fmt.Println("Virhe pelipäivämäärän muuntamisessa:", err)
		return true
	}

  dailyWordDate, err := time.Parse(layout, dailyWord.UsedAt.String)
	if err != nil {
		fmt.Println("Virhe päivittäisen sanan päivämäärän muuntamisessa:", err)
		return true
	}

  // calculate date difference
  dateDifference := dailyWordDate.Sub(gameDate).Hours() / 24

  // if game is won, value 1 is acceptable
  if game.IsWon {
    return dateDifference > 1
  }

  // if game is not won, check if date difference is over 0
	return dateDifference > 0
}

func CreateDebugData(game types.Game) types.Debug {
  // create debug data
  var debugData types.Debug
  debugData.Guid = game.Guid
  debugData.Runes = [][]rune{}
  for _, guess := range game.Guesses {
    var runes []rune
    for _, letterMap := range guess {
      key, _ := getKeyAndValueFromMap(letterMap)
      runes = append(runes, []rune(key)[0])
    }
    debugData.Runes = append(debugData.Runes, runes)
  }

  return debugData
}

func GetGameFromSession(session *sessions.Session) (types.Game, error) {
  // create game struct
  var game types.Game
  // fetch game data from session
  if gamedataStr, ok := session.Values["gamedata"].(string); ok {
    // convert json string to bytes
    gamedataBytes := []byte(gamedataStr)
    // unmarshal json bytes to game struct
    if err := json.Unmarshal(gamedataBytes, &game); err != nil {
      return game, err
    }
  } else {
    // init empty game
    game = types.Game{ Guesses: [][][]string{}, IsComplete: false, IsWon: false, Streak: 0}
  }
  return game, nil
}

func CheckWordBoundaries(guess []string, game types.Game) (types.Notification, bool) {
  // word must be exactly 5 letters long
  if len(guess) != 5 {
    return types.Notification{Type: "error", Message: "Sanan pitää olla 5 merkkiä pitkä"}, false
  }

  // if previous guess exists
  if len(game.Guesses) > 0 {
    
    // word must be different from any previous guesses
    for _, previousGuess := range game.Guesses {
      if strings.Join(guess, "") == concatenateKeys(previousGuess) {
        return types.Notification{Type: "error", Message: fmt.Sprintf("Olet jo arvannut sanan ”%s”", strings.Join(guess, ""))}, false
      }
    }

    if previousGuess := game.Guesses[len(game.Guesses) - 1]; previousGuess != nil {
      for i, letter := range guess {
        char, status := getKeyAndValueFromMap(previousGuess[i])
        if status == CORRECT && char != letter {
          // each letter in CORRECT must be in the same position as previous guess
          return types.Notification{Type: "error", Message: fmt.Sprintf("Kirjaimen ”%s” on oltava oikealla paikallaan", char)}, false
        } else if status == FOUND && !strings.Contains(strings.Join(guess, ""), char) {
          // each letter in FOUND must exist in guess
          return types.Notification{Type: "error", Message: fmt.Sprintf("Kirjain ”%s” täytyy esiintyä arvauksessa", char)}, false
        }
      }
    }
  }

  return types.Notification{}, true
}

// return true if game is complete
// return true if game is won
func GameIsComplete(game types.Game) (bool, bool) {
  var won bool = true
  // game is won if last guess is correct
  if len(game.Guesses) > 0 {
    if previousGuess := game.Guesses[len(game.Guesses) - 1]; previousGuess != nil {
      for i, _ := range previousGuess {
        _, status := getKeyAndValueFromMap(previousGuess[i])
        won = won && status == CORRECT
      }

      // if won is still true, game is won
      if won {
        // return complete and won game
        return true, true
      }
    }
  }
  // game is complete if there are 6 guesses,
  return len(game.Guesses) == GAME_LENGTH, false
}

