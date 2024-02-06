package repository

import (
  "database/sql"
  "time"
  "github.com/joho/godotenv"
  "os"
  "log"
  _ "github.com/go-sql-driver/mysql"
)

func NewDatabase() (*sql.DB, error) {
  // Load environment variables
  err := godotenv.Load()
  // Check if environment variables are loaded
  if err != nil {
    return nil, err
  }

  db, err := sql.Open("mysql", os.Getenv("DB_USER")+":"+os.Getenv("DB_PASSWORD")+"@tcp("+os.Getenv("DB_HOST")+":"+os.Getenv("DB_PORT")+")/"+os.Getenv("DB_NAME"))
	if err != nil {
    return nil, err
	}

	err = db.Ping()
	if err != nil {
    return nil, err
	}

	log.Println("Connected to the database!")
  // Set maximum number of connections
  db.SetMaxOpenConns(25)
  db.SetConnMaxLifetime(time.Minute)
  db.SetMaxIdleConns(10)

  return db, nil
}

