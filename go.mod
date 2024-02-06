module word-it-out

require word-it-out/app v0.0.0

require word-it-out/game v0.0.0 // indirect

require (
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/gorilla/context v1.1.2 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/gorilla/sessions v1.2.2 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	word-it-out/game/service v0.0.0-00010101000000-000000000000 // indirect
	word-it-out/game/types v0.0.0-00010101000000-000000000000 // indirect
	word-it-out/repository v0.0.0-00010101000000-000000000000 // indirect
)

replace word-it-out/app => ./app

replace word-it-out/game => ./game

replace word-it-out/game/types => ./game/types

replace word-it-out/game/service => ./game/service

replace word-it-out/repository => ./repository

go 1.21.6
