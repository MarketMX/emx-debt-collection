# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Architecture

This is a full-stack debt collection application built with Go backend and React frontend:

### Backend (Go)
- **Framework**: Echo v4 web framework
- **Database**: PostgreSQL with pgx/v5 driver
- **Structure**: Standard Go project layout
  - `cmd/api/main.go` - Application entry point with graceful shutdown
  - `internal/server/` - HTTP server setup and route handlers
  - `internal/database/` - Database connection and queries
- **Testing**: Uses testcontainers for integration tests

### Frontend (React + TypeScript)
- **Build Tool**: Vite
- **Styling**: Tailwind CSS v4
- **Location**: `frontend/` directory
- **Port**: Development server runs on port 5173

## Development Commands

### Core Development
- `make run` - Starts both Go backend and React frontend (recommended for development)
- `make watch` - Live reload using Air (Go backend only)
- `make build` - Build the Go application binary
- `make test` - Run Go unit tests
- `make itest` - Run database integration tests

### Database
- `make docker-run` - Start PostgreSQL container
- `make docker-down` - Stop PostgreSQL container

### Frontend (from `/frontend` directory)
- `npm run dev` - Start Vite dev server
- `npm run build` - Build for production
- `npm run lint` - Run ESLint

### Cleanup
- `make clean` - Remove built binaries

## Environment Configuration

The application uses `.env` file for configuration:
- `PORT=8080` - Backend server port
- Database connection settings prefixed with `BLUEPRINT_DB_`
- PostgreSQL container accessible at localhost:5432

## Key Dependencies

### Backend
- Echo v4 - Web framework
- pgx/v5 - PostgreSQL driver
- testcontainers - Integration testing
- Air - Live reload (development)

### Frontend
- React 19 with TypeScript
- Vite 7 - Build tool and dev server
- Tailwind CSS v4 - Styling
- ESLint - Code linting

## Development Workflow

1. Start database: `make docker-run`
2. Run application: `make run` (starts both backend and frontend)
3. Backend runs on port 8080, frontend on port 5173
4. Use `make watch` for backend-only development with live reload
5. Run tests with `make test` and `make itest`