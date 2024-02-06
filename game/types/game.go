package types

import "database/sql"

type Game struct {
  Guid string `json:"guid"`
  IsComplete bool `json:"isComplete"`
  IsWon bool `json:"isWon"`
  Guesses [][]map[string]string `json:"guesses"`
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
