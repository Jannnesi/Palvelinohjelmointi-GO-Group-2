# Palvelinohjelmointi GO Group 2

Time-tracking example for SeAMK’s service programming course. The app shows how to build a Go HTTP API backed by SQLite via GORM, serve a simple frontend, and package everything with Docker.

## Features
- RESTful JSON endpoints for `/timeentries`, plus `/health` and a mock `/login`.
- SQLite relational database (`worklogger.db`) auto-migrated and seeded at startup.
- GORM domain models (`internal/domain/*.go`) for ORM access without raw SQL.
- Static worker dashboard (`frontend/workerdashboard.html`) that consumes the API.
- Dockerfile for reproducible builds and easy sharing.

## Tech Stack
- Go 1.22+ (module `github.com/Jannnesi/Palvelinohjelmointi-GO-Group-2`)
- GORM ORM with `github.com/glebarez/sqlite`
- SQLite (file-based) database
- Vanilla HTML/CSS/JS frontend served by `net/http`
- Docker multi-stage build

## Repo Layout
```
cmd/server/         # main entrypoint (starts HTTP server)
internal/config     # env config loader
internal/database   # DB connection, migrations, seeding
internal/domain     # GORM models for users and time entries
internal/logger     # lightweight logger wrapper
internal/router     # HTTP handlers and routing
frontend/           # static UI assets
scripts/            # helper scripts (e.g., backlog importer)
Dockerfile          # container build definition
worklogger.db       # SQLite database file
```

## Configuration
| Variable      | Default | Description                  |
|---------------|---------|------------------------------|
| `SERVER_PORT` | `8080`  | Port for the HTTP server     |
| `LOG_LEVEL`   | `info`  | Logger level (`debug`, etc.) |

`internal/config/config.go` pulls these from the environment.

## Running Locally (Go)
```sh
go run ./cmd/server
# open http://localhost:8080/frontend/workerdashboard.html
```
`internal/database.Connect()` auto-migrates schemas and seeds a sample entry, so no manual SQL prep is required.

## API Summary
| Method | Path                | Purpose                          |
|--------|---------------------|----------------------------------|
| GET    | `/`                 | List endpoints                   |
| GET    | `/health`           | Health probe (`{"status":"ok"}`) |
| GET    | `/timeentries`      | List all time entries            |
| POST   | `/timeentries`      | Create a new entry               |
| GET    | `/timeentries/{id}` | Fetch entry by ID                |
| PUT    | `/timeentries/{id}` | Update entry                     |
| DELETE | `/timeentries/{id}` | Delete entry                     |
| POST   | `/login`            | Mock login (worker/manager role) |

Requests/responses are JSON and align with `internal/domain/timeentry.go`.

## Frontend
`frontend/workerdashboard.html` is served at `/frontend/…`. It:
- pulls `/timeentries` to render a table and totals,
- posts new entries via `fetch("/timeentries")`,
- defaults the date field to today and calculates `EndTime` from the entered hours.

## Docker Workflow
Build:
```sh
docker build -t worklogger .
```
Run:
```sh
docker run --rm -p 8080:8080 \
  -v "$(pwd)/worklogger.db:/app/worklogger.db" \
  worklogger
```
- Port 8080 is exposed; browse `http://localhost:8080/frontend/workerdashboard.html`.
- The volume mount keeps the SQLite file on the host. Omit it if you want a throwaway DB.
- Docker Desktop users can use the UI: select the `worklogger` image → Run → map port 8080 and optional volume.

## Notes
- Delete `worklogger.db` to reset seeded demo data.
- `.scripts/backlog/README.md` documents the backlog-import helper script.
- `/login` is presently a mock endpoint; add real authentication + middleware before production use.
- Tests aren’t included yet—add handler/service tests as the next step.
