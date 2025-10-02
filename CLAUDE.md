# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go backend application for "kabancount" - a web API service built with:
- **Framework**: Go with Chi router for HTTP routing
- **Database**: PostgreSQL with pgx driver
- **Migrations**: Goose for database migrations with embedded filesystem
- **Configuration**: Viper with environment variable support
- **Authentication**: JWT-based authentication with role-based access control
- **Architecture**: Clean architecture with separated concerns (handlers, stores, middleware)

## Development Commands

### Running the Application

**Using Makefile (Recommended)**:
```bash
# Set up development environment (installs Air, starts databases, tidies modules)
make setup-dev

# Start development server with hot reload
make dev

# Run application directly
make run

# Build the application
make build

# Start PostgreSQL databases
make docker-up

# Stop PostgreSQL databases
make docker-down
```

**Manual Commands**:
```bash
# Start PostgreSQL database (includes test database)
docker-compose up -d

# Run the application (defaults to port 8080)
go run main.go

# Run with custom port
go run main.go -port=3000
```

### Database Operations
```bash
# Start database only
docker-compose up -d db

# Start test database only
docker-compose up -d test_db

# Connect to main database
docker exec -it kabancount_db psql -U postgres -d kabancount

# Connect to test database
docker exec -it kabancount_test_db psql -U postgres -d kabancount_test
```

### Development Commands

**Makefile Commands (Full List)**:
```bash
# Development
make help           # Show all available commands
make setup-dev      # Set up development environment
make dev            # Start with hot reload (Air)
make start          # Alias for 'make dev'
make stop           # Stop development server
make restart        # Restart development server

# Building & Testing
make build          # Build the application binary
make run            # Run the application directly
make test           # Run all tests
make test-coverage  # Run tests with coverage report
make clean          # Clean build artifacts

# Database Operations
make docker-up      # Start PostgreSQL databases
make docker-down    # Stop PostgreSQL databases
make docker-reset   # Reset databases (fresh start)
make docker-logs    # View database logs
make db-connect     # Connect to main database
make db-connect-test # Connect to test database

# Code Quality
make tidy           # Clean up Go modules
make lint           # Run linter (requires golangci-lint)
make all            # Build and test (CI/CD target)
```

**Manual Go Commands**:
```bash
# Build the application
go build

# Run tests
go test ./...

# Run tests for specific package
go test ./internal/store

# Install dependencies
go mod tidy

# View dependencies
go mod graph
```

## Architecture

### Project Structure
- `main.go` - Application entry point with HTTP server setup and command-line flag parsing
- `internal/app/` - Application initialization, dependency injection, and Application struct
- `internal/routes/` - HTTP route definitions using Chi router with middleware groups
- `internal/api/` - HTTP handlers for API endpoints (users, organizations, items, categories, auth, tokens)
- `internal/store/` - Database connection, migration execution, and data access layer
- `internal/middleware/` - Authentication and authorization middleware
- `internal/config/` - Configuration management using Viper with environment variable binding
- `internal/tokens/` - JWT token utilities
- `internal/cookie/` - Cookie management utilities
- `internal/pagination/` - Pagination utilities
- `internal/utils/` - General utility functions
- `migrations/` - SQL migration files with embedded filesystem

### Key Components

**Application Bootstrap** (`internal/app/app.go`):
- Loads configuration from environment variables and .env file
- Initializes database connection using configuration
- Runs migrations automatically on startup using embedded filesystem
- Creates all store instances (user, organization, token, item, category)
- Initializes handlers with dependency injection
- Sets up middleware with user store dependency

**Configuration Management** (`internal/config/config.go`):
- Uses Viper for configuration with environment variable binding
- Supports .env file loading with godotenv
- Validates required configuration (JWT secret, database credentials)
- Provides database DSN construction
- Includes configurations for server, database, JWT, AWS, and app environment

**Database Layer** (`internal/store/database.go`):
- PostgreSQL connection using pgx driver via database/sql
- Automatic migration execution using Goose with embedded filesystem
- Configuration-driven connection (supports environment variables)
- Separate test database configuration on port 5433

**HTTP Layer** (`internal/routes/routes.go`):
- Chi router with middleware groups for authentication
- Public routes: `/healthcheck`, `/auth/signin`, `/auth/signup`
- Protected routes requiring authentication and user context
- Admin-only routes for organization management
- RESTful endpoints for items and categories with CRUD operations

**Authentication & Authorization**:
- JWT-based authentication with configurable secret
- Role-based access control (admin vs regular users)
- Middleware for authentication, user context, and admin requirements
- Token management with database persistence

### Database Schema
The application includes migrations for:
- `users` table with organization relationships
- `organizations` table
- `tokens` table for JWT token management
- `items` table with category relationships
- `categories` table with organization scoping
- `stocks` table

### Development Notes
- Configuration is environment-driven with sensible defaults
- Migrations run automatically on application start using embedded filesystem
- Clean architecture with clear separation between handlers, business logic, and data access
- Uses Go 1.25.1 with standard HTTP server configuration (timeouts configured)
- Docker Compose provides both main and test databases
- Authentication required for most endpoints with role-based access control
- **Air Integration**: Live reload for development with `.air.toml` configuration
- **Makefile**: Comprehensive set of commands for development workflow
- Hot reload excludes test files, database directories, and temporary files