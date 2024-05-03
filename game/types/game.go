package types

import "database/sql"

type Game struct {
  Guid string `json:"guid"`
  IsComplete bool `json:"isComplete"`
  IsWon bool `json:"isWon"`
  Guesses [][][]string `json:"guesses"`
  UsedAt string `json:"date"`
  Streak int `json:"streak"`
}

type Notification struct {
  Type string `json:"type"`
  Message string `json:"message"`
}

type Word struct {
  Guid string
  Word string
  UsedAt sql.NullString
}

type Debug struct {
  Guid string `json:"guid"`
  Runes [][] rune `json:"runes"`
}

