# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A full-stack debt collection application (Age Analysis Messaging Application) with Go backend, React frontend, and Keycloak authentication.

## Architecture

### Backend (Go)
- **Framework**: Echo v4 web framework with middleware
- **Database**: PostgreSQL 15 with pgx/v5 driver
- **Authentication**: Keycloak integration with JWT validation
- **Project Structure**:
  - `cmd/api/main.go` - Entry point with graceful shutdown
  - `internal/server/` - HTTP server and route definitions
  - `internal/handlers/` - Request handlers (upload, messaging, admin, reports, system)
  - `internal/services/` - Business logic (excel, account, messaging, progress)
  - `internal/database/` - Database connection and repository pattern
  - `internal/middleware/` - Auth and user context middleware
  - `internal/models/` - Data models (account, upload, message_log, user)
  - `internal/config/` - Configuration loaders (keycloak, messaging)
  - `internal/utils/` - Utility functions and helpers

### Frontend (React + TypeScript)
- **Build Tool**: Vite 7
- **Styling**: Tailwind CSS v4
- **UI Components**: Radix UI primitives with shadcn/ui patterns
- **State Management**: React Query v5 (TanStack Query)
- **Authentication**: Keycloak JS adapter with React Keycloak
- **Forms**: React Hook Form with Zod validation
- **Routing**: React Router v7
- **Development Port**: 5173
- **Path Alias**: `@/` maps to `src/` directory

### Services Architecture
- **PostgreSQL**: Shared database for app and Keycloak (port 5432)
- **Keycloak**: Authentication service (port 8081)
- **Docker Compose**: Multi-service orchestration with health checks

## Development Commands

### Quick Start
```bash
make docker-run  # Start PostgreSQL and Keycloak
make run        # Start backend (8080) and frontend (5173) together
```

### Backend Development
- `make build` - Build Go binary
- `make watch` - Live reload with Air (backend only)
- `make test` - Run all Go tests
- `make itest` - Run integration tests (uses testcontainers)
- `go test ./internal/handlers -v` - Test specific package
- `go test -run TestName ./...` - Run specific test
- `go test -cover ./...` - Run tests with coverage report

### Frontend Development (from `/frontend`)
- `npm run dev` - Start Vite dev server
- `npm run build` - Production build
- `npm run lint` - Run ESLint
- `npm run preview` - Preview production build
- `npm install` - Install dependencies

### Docker & Services
- `make docker-run` - Start PostgreSQL and Keycloak
- `make docker-down` - Stop all Docker services
- `docker-compose logs -f keycloak` - View Keycloak logs
- `docker-compose logs -f psql_bp` - View PostgreSQL logs
- `docker-compose ps` - Check service health
- `docker-compose restart keycloak` - Restart Keycloak service

### Cleanup
- `make clean` - Remove Go binaries
- `make all` - Build and test the application

## Environment Configuration

### Required Environment Variables
```bash
# Backend Server
PORT=8080

# Database (BLUEPRINT_DB_ prefix)
BLUEPRINT_DB_HOST=localhost
BLUEPRINT_DB_PORT=5432
BLUEPRINT_DB_USERNAME=postgres
BLUEPRINT_DB_PASSWORD=postgres
BLUEPRINT_DB_DATABASE=debt_collection
BLUEPRINT_DB_SCHEMA=public

# Keycloak Authentication
KEYCLOAK_SERVER_URL=http://localhost:8081
KEYCLOAK_REALM=debt-collection
KEYCLOAK_CLIENT_ID=debt-collection-backend
KEYCLOAK_CLIENT_SECRET=your-client-secret

# Messaging Service (optional, defaults to simulation mode)
MESSAGING_API_URL=https://api.messaging.example.com
MESSAGING_API_KEY=your-api-key
MESSAGING_SIMULATION=true  # Set false for production
MESSAGING_RATE_LIMIT=5      # Messages per second

# User Provisioning API (for Django Admin integration)
PROVISIONING_API_KEY=your-secure-provisioning-api-key-here
```

## API Structure

### Public Endpoints (No Auth)
- `GET /health` - Health check
- `GET /auth/config` - Keycloak configuration for frontend

### API Key Protected Endpoints
- **Provisioning**: `/api/provisioning/*` - User provisioning from Django Admin
- **Webhooks**: `/api/webhooks/*` - Real-time sync events from external systems

### Protected Endpoints (JWT Required)
- **Uploads**: `/api/uploads/*` - File upload and account management
- **Messaging**: `/api/messaging/*` - Send messages and view logs
- **Reports**: `/api/reports/*` - Various reporting endpoints
- **System**: `/api/system/*` - System information and metrics

### Admin Endpoints (Admin Role Required)
- `/api/admin/*` - User management and system administration

## Key Dependencies

### Backend
- `github.com/labstack/echo/v4` - Web framework
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/xuri/excelize/v2` - Excel file processing
- `github.com/golang-jwt/jwt/v5` - JWT token handling
- `github.com/testcontainers/testcontainers-go` - Integration testing
- `github.com/air-verse/air` - Live reload (dev only)

### Frontend
- React 19 with TypeScript 5.8
- Vite 7 - Build and dev server
- Tailwind CSS v4 - Styling
- Radix UI - Headless components
- React Query v5 - Server state management
- Keycloak JS 26 - Authentication
- React Hook Form 7 + Zod 4 - Form handling

## Testing Strategy

### Backend Testing
- Unit tests: Standard Go testing with mocks
- Integration tests: testcontainers for database tests
- Run specific: `go test -run TestName ./package`
- Coverage: `go test -cover ./...`

### Frontend Testing
- ESLint for code quality
- TypeScript for type safety

## Authentication Flow

1. Frontend requests auth config from `/auth/config`
2. User authenticates via Keycloak
3. JWT tokens validated by backend middleware
4. User context injected into protected routes
5. Role-based access control for admin endpoints

## Messaging Service

The application includes a messaging service abstraction that:
- Supports rate limiting and retry logic
- Operates in simulation mode by default (development)
- Tracks all messages in the database
- Provides progress tracking for batch operations

## Database Schema

Key tables:
- `users` - Application users synced from Keycloak
- `accounts` - Debt collection accounts
- `uploads` - File upload tracking
- `message_logs` - Messaging history and status

Migration files:
- `/migrations/001_initial_schema.sql` - Initial database schema

## Development Workflow

1. **Initial Setup**:
   - Copy `.env.example` to `.env`
   - Configure Keycloak credentials (see KEYCLOAK_SETUP.md for details)
   - Start services: `make docker-run`
   - Wait for health checks to pass

2. **Development**:
   - Run app: `make run` (starts both backend and frontend)
   - Backend changes: Auto-reload if using `make watch`
   - Frontend changes: Vite hot module replacement
   - Air config: `.air.toml` for live reload settings

3. **Testing**:
   - Run tests: `make test` (unit) and `make itest` (integration)
   - Check health: `curl localhost:8080/health`
   - Frontend linting: `cd frontend && npm run lint`

4. **Keycloak Admin**:
   - Access: http://localhost:8081
   - Default: admin/admin123
   - Realm: debt-collection
   - Backend client: debt-collection-backend

## CORS Configuration

Backend configured to accept requests from:
- http://localhost:5173 (Vite dev)
- http://localhost:3000 (Alternative React port)
- HTTPS variants of above

## File Upload Processing

1. Excel files uploaded via `/api/uploads`
2. Parsed using excelize library
3. Accounts extracted and stored in database
4. Progress tracked via progress service
5. Selection management for messaging

## Common Issues & Solutions

- **Port conflicts**: Check ports 8080, 8081, 5173, 5432 are free
- **Keycloak startup**: Wait for PostgreSQL health check before accessing
- **JWT validation**: Ensure Keycloak realm and client configured correctly
- **Database connection**: Verify PostgreSQL container is running and healthy
- **Frontend proxy**: Vite proxies `/api` and `/auth` to backend automatically
- **Air not installed**: Run `go install github.com/air-verse/air@latest` for live reload
- **Testcontainers**: Docker must be running for integration tests