# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`word-it-out` (Finnish name "Sanaseppo") is the Go backend for a Wordle-style daily word
guessing game. Guesses and daily-word selection are in Finnish (see the user-facing
notification strings in `game/service/game.go`).

## Commands

```bash
go build ./...       # build everything (module + submodules)
go vet ./...          # static checks
go run .              # run the server locally (reads config from .env)
go test ./...          # run tests (none exist yet)
```

There is no Makefile, Dockerfile, or CI config in this repo.

The server needs a running MySQL instance and a `.env` file (gitignored) with at least:
`PORT`, `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `SESSION_NAME`,
`SESSION_SECRET`, `CLIENT_URL`, `ADMIN_SECRET`, and optionally `CERT_DIR`/`CERT_FILE`/`KEY_FILE`
to serve TLS instead of plain HTTP.

## Architecture

This is a **multi-module Go workspace** (no `go.work` file — wiring is done entirely via
`replace` directives in each `go.mod`). Each directory below is its own Go module:

- `word-it-out` (root, `main.go`) — entrypoint, just calls `app.App{}.Run()`.
- `app` — HTTP wiring: router (gorilla/mux), CORS middleware, session middleware, route table,
  and TLS vs. plain HTTP server startup. `app.go` is the place to look for the full route list.
- `game` — HTTP handlers (`Controller` in `controller.go`) and DB access
  (`GameRepository` in `repository.go`). Depends on the two submodules below.
  - `game/service` — pure game logic: comparing a guess against the daily word
    (`CompareWord`), guess validation against Wordle-style constraints
    (`CheckWordBoundaries` — repeats, and letters that were CORRECT/FOUND in the immediately
    previous guess), win/completion detection (`GameIsComplete`), and session
    serialization of `types.Game` (`SetGameToSession`/`GetGameFromSession`).
  - `game/types` — shared structs (`Game`, `Word`, `Notification`, `Debug`) with JSON tags.
- `repository` — the MySQL connection pool (`database/sql` + `go-sql-driver/mysql`),
  configured from env vars.

Each submodule is pulled in via `replace word-it-out/x => ./x` in the consuming module's
`go.mod`, so after editing a submodule's exported API, its consumers' `go.mod`/`go.sum`
may need `go mod tidy` run from that consumer's directory.

### Request flow

1. `app.Run()` registers routes on a gorilla/mux router: `POST /word` (admin-only, bulk
   insert candidate words), `POST /guess`, `GET /game`, `GET /debug`.
2. `sessionMiddleware` loads a gorilla/sessions cookie session and stashes it on the request
   context (via `gorilla/context`, not `r.Context()`) under the key `"session"`; handlers
   pull it back out with `context.Get(r, "session").(*sessions.Session)`.
3. Per-request game state (`types.Game`) lives **in the session cookie** as a JSON string
   (`SetGameToSession`/`GetGameFromSession`), not in the database — the database only stores
   the word pool and which word/date is "today's" word.
4. `GameRepository.GetDailyWord()` is the source of truth for "today's word": it first looks
   for a word already marked `used_at = CURRENT_DATE`; if none exists it picks a random unused
   word and marks it, inside a transaction, re-reading if another concurrent request won the
   race (`rowsAffected == 0` branch) — this is the mechanism that keeps the daily word
   consistent across concurrent requests without an explicit lock.
5. `GET /game` compares the session's game to the current daily word
   (`service.GameIsActive`); if they don't match (new day), it resets guesses and carries the
   streak forward unless `service.GameIsTooOld` says the player missed a day.
6. `POST /guess` validates the guess is a real word (`GameRepository.WordExists`), checks
   boundary rules (`CheckWordBoundaries`), scores it (`CompareWord` — first pass marks exact
   position matches as `correct`, second pass marks remaining present letters as `found`,
   respecting per-letter remaining counts so duplicate letters are scored correctly), then
   updates completion/win state and streak.

### Notes

- Indentation in this codebase is inconsistent (mix of 2-space in older files and tabs/gofmt
  style in newer ones) — match the surrounding file rather than reformatting wholesale.
- `game/repository.go`'s `InsertWords` and `NewGameRepository` call `log.Fatal` on error,
  which kills the whole process — be aware of this blast radius if touching startup or the
  `POST /word` path.
