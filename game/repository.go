package game

import (
  "log"
  "database/sql"
  "word-it-out/repository"
  "word-it-out/game/types"
)

type GameRepository struct {
  con *sql.DB
}

func NewGameRepository() (*GameRepository, error) {
  con, err := repository.NewDatabase()
  if err != nil {
    return nil, err
  }

  return &GameRepository{con}, nil
}

func (r *GameRepository) InsertWords(words []string) {
  stmt, err := r.con.Prepare("INSERT INTO words (guid, word) VALUES (uuid(), ?)")
  if err != nil {
    log.Fatal(err)
  }
  defer stmt.Close()
  for _, word := range words {
    _, err = stmt.Exec(word)
    if err != nil {
      log.Fatal(err)
    }
  }
}

func (r *GameRepository) WordExists(word string) (bool, error) {
  var found bool
  err := r.con.QueryRow("SELECT EXISTS(SELECT 1 FROM words WHERE word = ?) AS found", word).Scan(&found)

  if err != nil {
    return false, err
  }
  return found, nil
}

func (r *GameRepository) GetDailyWord() (types.Word, error) {
  var word types.Word

  // fetch word from database
  err := r.con.QueryRow("SELECT guid, word, used_at FROM words ORDER BY CASE WHEN used_at = CURRENT_DATE THEN 0 ELSE 1 END, RAND() LIMIT 1").Scan(&word.Guid, &word.Word, &word.UsedAt)

  if err != nil {
    return word, err
  }

  // if word was not used today, update used_at
  if !word.UsedAt.Valid {
    stmt, err := r.con.Prepare("UPDATE words SET used_at = CURRENT_DATE WHERE guid = ?")

    if err != nil {
      return word, err
    }

    defer stmt.Close()

    _, err = stmt.Exec(word.Guid)
    if err != nil {
      return word, err
    }

    err = r.con.QueryRow("SELECT used_at FROM words WHERE guid = ?", word.Guid).Scan(&word.UsedAt)
    if err != nil {
      return word, err
    }
  }

  return word, nil
}

