package game

import (
	"database/sql"
	"errors"
	"log"
	"word-it-out/game/types"
	"word-it-out/repository"
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

	tx, err := r.con.Begin()
	if err != nil {
		return word, err
	}
	defer tx.Rollback()

	// if today's word already exists, always use it
	err = tx.QueryRow("SELECT guid, word, used_at FROM words WHERE used_at = CURRENT_DATE LIMIT 1").Scan(&word.Guid, &word.Word, &word.UsedAt)
	if err == nil {
		if err := tx.Commit(); err != nil {
			return word, err
		}
		return word, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return word, err
	}

	// otherwise pick an unused word and mark it as today's word
	err = tx.QueryRow("SELECT guid, word, used_at FROM words WHERE used_at IS NULL ORDER BY RAND() LIMIT 1").Scan(&word.Guid, &word.Word, &word.UsedAt)
	if err != nil {
		return word, err
	}

	result, err := tx.Exec("UPDATE words SET used_at = CURRENT_DATE WHERE guid = ? AND used_at IS NULL", word.Guid)
	if err != nil {
		return word, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return word, err
	}

	// another request may have won the race; read today's word in that case
	if rowsAffected == 0 {
		err = tx.QueryRow("SELECT guid, word, used_at FROM words WHERE used_at = CURRENT_DATE LIMIT 1").Scan(&word.Guid, &word.Word, &word.UsedAt)
		if err != nil {
			return word, err
		}
	} else {
		err = tx.QueryRow("SELECT used_at FROM words WHERE guid = ?", word.Guid).Scan(&word.UsedAt)
		if err != nil {
			return word, err
		}
	}

	if err := tx.Commit(); err != nil {
		return word, err
	}

	return word, nil
}
