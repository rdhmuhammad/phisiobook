# AGENTS.md

This file provides guidance to Codex (Codex.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based backend API project (phisiobook-api) built with the Gin web framework. The codebase follows a clean architecture pattern with generic repository implementations and comprehensive middleware support.

## Development Commands

### Running the Application

```bash
# Run with specific environment file
go run -env .env.stag cmd/api/api.go

# Or use the Makefile
make dev
```

The application runs on the port specified in `APP_PORT` environment variable (default: 8999).

### Environment Configuration

- Copy `.env.example` to create environment files (`.env.stag`, `.env.dev`, etc.)
- The application defaults to `.env.stag` if no `-env` flag is provided
- Environment file path can be specified using the `-env` flag

### Testing and Building

```bash
# Build the application
go build -o bin/api cmd/api/api.go

# Run tests
go test ./...

# Run tests for a specific package
go test ./pkg/db

# Run a single test
go test -run TestName ./path/to/package
```

### Dependencies

```bash
# Install dependencies
go mod download

# Update dependencies
go mod tidy

# Add a new dependency
go get github.com/package/name
```

## Architecture

### Directory Structure

- `cmd/api/` - Application entry point. Loads environment variables and starts the API server
- `internal/` - Private application code not meant to be imported by other projects
  - `constant/` - Application-wide constants (datetime formats, error messages, roles)
  - `core/domain/` - Domain models/entities (User, UserAdmin, etc.)
  - `dto/` - Data transfer objects for internal use
  - `localerror/` - Custom error types
- `pkg/` - Public packages that can be imported by other projects
  - `api/` - API server setup and routing infrastructure
  - `cache/` - Redis cache client wrapper
  - `db/` - Database connection and generic repository pattern
  - `middleware/` - HTTP middleware (auth, CORS, validation, Sentry)
  - `dto/` - Shared DTOs for requests/responses
  - `localize/` - i18n support using go-i18n
  - `davinci/` - Hashing and security utilities
  - `environment/` - Environment variable helper
  - `mapper/` - Data mapping utilities
  - `miniostorage/` - MinIO object storage client
  - `inetproto/` - HTTP client utilities
- `resource/message/` - i18n message files (en.json, id.json)

### API Routing Pattern

The API uses a router interface pattern defined in `pkg/api/api.go`:

```go
type Router interface {
    Route(handler *gin.RouterGroup)
}
```

All routes are mounted under `/api/v1` prefix. To add new routes:
1. Create a router struct implementing the `Router` interface
2. Add it to the `routers` slice in `pkg/api/default.go`

### Generic Repository Pattern

The codebase uses a powerful generic repository (`pkg/db/generic_repository.go`) for database operations:

- `GenericRepository[T]` - Type-safe repository for any GORM model implementing `schema.Tabler`
- Query building helpers: `Search()`, `Equal()`, `InArray()`, `NotInArray()`, `ExpressionDateRange()`
- Expression types: `ExpressionOr` and `ExpressionAnd` for combining clauses
- Pagination support via `PaginationQuery` struct
- Preload with conditions via `PreloadWithCondition`
- Selection queries for fetching specific columns

Example usage:
```go
repo := db.NewGenericeRepo(dbConn, domain.User{})
users, err := repo.FindAllByExpression(ctx, []clause.Expression{
    db.Equal(userId, "id"),
})
```

### Database Configuration

MySQL connection is managed in `pkg/db/default.go`:
- Environment-specific host selection (development, staging, docker)
- Configurable log levels via `DB_LOG_MODE`:
  - 1 = Silent
  - 2 = Error (default)
  - 3 = Warn
  - 4 = Info
- Batch insert size: 500 records

### Authentication & Middleware

Authentication is JWT-based (`pkg/middleware/authenticate.go`):
- `Auth.Validate()` - Middleware to validate JWT tokens
- `Auth.Authorize(roles...)` - Role-based authorization
- `Auth.GetAuthDataFromContext(c)` - Extract user data from context
- Session data stored in Redis cache
- Automatic user activity tracking (updates `last_active` field)

Other middleware:
- `AllowCORS()` - CORS configuration
- `SentryMiddleware()` - Error tracking and enrichment
- Custom validators with i18n support
- Idempotency support

### External Services

- **Database**: MySQL via GORM
- **Cache**: Redis for session management and caching
- **Storage**: MinIO for object storage
- **Monitoring**: Sentry for error tracking
- **Email**: SendInBlue (Brevo) for transactional emails

### Localization

The app supports multiple languages using go-i18n:
- Message files in `resource/message/` (en.json, id.json)
- Access via `localize.Language` interface
- Language detection from request context

### Error Handling

- Custom error types in `pkg/localerror/`
- Sentry integration for error capture
- `CaptureErrorUsecase()` helper for logging errors to Sentry
- Standardized error responses via `pkg/dto/response.go`

## Code Patterns

### Creating a New Feature Module

1. Define domain models in `internal/core/domain/` if needed
2. Create repository using `GenericRepository[YourModel]`
3. Implement business logic in a service/usecase layer
4. Create DTOs for request/response in appropriate dto package
5. Create router implementing `Router` interface
6. Register router in `pkg/api/default.go`

### Database Queries

Use the generic repository's expression builders:
```go
// Search across multiple columns
db.Query(
    db.Search(searchTerm, "name", "email"),
    db.Equal(status, "status"),
)

// Date range queries
db.ExpressionDateRange(startDate, endDate, "created_at", "table_name")

// Array operations
db.InArray([]string{"active", "pending"}, "status")
```

### Working with Transactions

Use `pkg/db/dbTransaction.go` for transaction management:
```go
err := WithTransaction(ctx, func(tx *gorm.DB) error {
    // Your transactional operations
    return nil
})
```

### Adding Middleware

Middleware is registered in `pkg/api/default.go`:
- CORS is applied globally
- Sentry middleware for error tracking
- Custom middleware can be added via `server.Use()`

### Pattern of Codex command
If i want to create new feature, i had to start with sentence NEW FEATURE => {info}, {info} is describing following point:
- feature name => with format => POST FeatureName /api/v1/endpoint
- module name => if module already exist, append this new feature to its controller, usecase
- request body => json formatted
- validation => list of validation for request
- response body => json formatted
- entity => if entity is not exist, than create one with field info at entity ddl
- entity ddl => if i want to create new table


### Workflow of Codex agent based on the command
- check if controller for (module name).go is exist at /internal/adapter/controller. following example code of /internal/adapter/controller/health-check.go
  - if exist, then append new endpoint, controller method, usecase interface method for (feature name) to controller file
  - if not exist, then create new controller, append the controller to []Router at /pkg/api/default.go
- check if usecase folder for (module name) is exist at /internal/core/usecase. following example code of /internal/core/usecase/health
  - if exist, then append new usecase method, dto response and request, for (feature name) to controller file
  - if not exist, then create new usecase folder
  - validation placed on usecase
- if entity is not exist
  - append ddl for creating new database to /resource/db-changelog/versioning-0.0.1.sql
  - create new entity at /internal/core/domain. entity should extend BaseEntity struct which contains required column for all table format

## Important Notes

- All API endpoints are prefixed with `/api/v1`
- JWT tokens expire based on `EXPIRED_TOKEN_JWT` environment variable
- Redis session keys are prefixed with `LOGIN_KEY_`
- The app uses UTC timezone for database operations (`parseTime=True&loc=UTC`)
- Column selection queries use GORM tags to map struct fields
- User activity tracking happens automatically on authenticated requests