package service

import (
  "strings"
  "encoding/json"
  "word-it-out/game/types"
  "github.com/gorilla/sessions"
  "fmt"
)

// create enums for correct, found, missed
const (
  CORRECT = "correct"
  FOUND = "found"
  MISSED = "missed"
  GAME_LENGTH = 6
)

// internal function to get key and value from map
func getKeyAndValueFromMap(letterMap map[string]string) (string, string) {
	for key, value := range letterMap {
		return key, value
	}
	return "", "" // Return empty strings if the map is empty (should not happen in this context)
}

// create a word from guess mapping
func concatenateKeys(previousGuess []map[string]string) string {
	var keys []string
	for _, letterMap := range previousGuess {
		for key := range letterMap {
			keys = append(keys, key)
		}
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

func CompareWord(guess []string, dailyWord types.Word) []map[string]string {
  // split daily word into array of letters
  dailyWordLetters := strings.Split(dailyWord.Word, "")

  // get char map of daily word
  charMap := getCharMap(dailyWord)

  var result []map[string]string
  
  // first mark the correct letters. Then mark the found letters. 
  for i, letter := range guess {
    entry := make(map[string]string)
    key := string(letter)
    // match each letter in guess to daily word
    // set result key as guess letter and value as "correct" or "found" or "missed"
    if dailyWordLetters[i] == letter {
      entry[key] = CORRECT
      charMap[rune(letter[0])]--

    } else {
      entry[key] = MISSED
    }

    result = append(result, entry)
  }

  
  // now loop result and set found letters
  // we need to iterate over the guess twice because letter can be found first and then be correct in the same guess
  for _, entry := range result {
    for key, value := range entry {
      if value == MISSED && strings.Contains(dailyWord.Word, key) && charMap[rune(key[0])] > 0 {
        entry[key] = FOUND
      }

      charMap[rune(key[0])]--
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

  // set session data
  session.Values["gamedata"] = string(jsonGameData)

  return nil
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
    game = types.Game{Guesses: [][]map[string]string{}}
  }
  return game, nil
}

func CheckWordBoundaries(guess []string, game types.Game) (types.Notification, bool) {
  // word must be exactly 5 letters long
  if len(guess) != 5 {
    return types.Notification{Type: "error", Message: "Word must be exactly 5 letters long"}, false
  }

  // if previous guess exists
  if len(game.Guesses) > 0 {
    
    // word must be different from any previous guesses
    for _, previousGuess := range game.Guesses {
      if strings.Join(guess, "") == concatenateKeys(previousGuess) {
        return types.Notification{Type: "error", Message: "Word must be different from previous guesses"}, false
      }
    }

    if previousGuess := game.Guesses[len(game.Guesses) - 1]; previousGuess != nil {
      for i, letter := range guess {
        char, status := getKeyAndValueFromMap(previousGuess[i])
        if status == CORRECT && char != letter {
          // each letter in CORRECT must be in the same position as previous guess
          return types.Notification{Type: "error", Message: fmt.Sprintf("Letter ”%s” must be in the same position as previous guess", char)}, false
        } else if status == FOUND && !strings.Contains(strings.Join(guess, ""), char) {
          // each letter in FOUND must exist in guess
          return types.Notification{Type: "error", Message: fmt.Sprintf("Letter ”%s” must exists in guess", char)}, false
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

