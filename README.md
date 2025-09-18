# Kabancount Backend

A multi-tenant inventory management API built with Go, designed to evolve into a comprehensive e-commerce platform. Currently focused on robust inventory functionality with plans for work orders, automatic material deduction, product costing, and full e-commerce capabilities.

## 🚀 Project Overview

Kabancount is a learning project transitioning toward production-ready status. It provides a foundation for inventory management with multi-organizational support, role-based access control, and RESTful API design.

### Current Features

- **Multi-tenant Architecture**: Organizations can independently manage their inventory
- **Inventory Management**: Full CRUD operations for items and categories
- **User Management**: Role-based access control (admin/user roles)
- **Authentication**: JWT-based authentication system
- **Data Integrity**: PostgreSQL with automated migrations
- **Pagination**: Built-in pagination support for list endpoints

### Planned Features

- **E-commerce Platform**: Full online store functionality
- **Work Orders**: Automated material deduction and product assembly
- **Cost Management**: Comprehensive costing and pricing systems
- **Advanced Inventory**: Stock tracking, alerts, and analytics

## 🛠 Tech Stack

- **Language**: Go 1.24.4
- **Framework**: Chi router for HTTP routing
- **Database**: PostgreSQL with pgx driver
- **Migrations**: Goose with embedded filesystem
- **Configuration**: Viper with environment variable support
- **Authentication**: JWT tokens with configurable expiration
- **Architecture**: Clean architecture with separated concerns

## 📋 Prerequisites

- Go 1.24.4 or later
- Docker and Docker Compose
- PostgreSQL (via Docker)

## 🚀 Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd kabancount-be
go mod tidy
```

### 2. Environment Configuration

Create a `.env` file in the root directory:

```env
# Database Configuration
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=kabancount
DATABASE_USER=postgres
DATABASE_PASSWORD=password
DATABASE_SCHEMA=public

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here

# Server Configuration
PORT=8080

# Application Environment
APP_ENV=local

# AWS Configuration (for future features)
AWS_REGION=us-east-1
S3_BUCKET_MASTER=your-bucket-name
```

### 3. Start Database

```bash
# Start PostgreSQL (includes test database)
docker-compose up -d

# Verify database is running
docker ps
```

### 4. Run Application

```bash
# Run with default port (8080)
go run main.go

# Run with custom port
go run main.go -port=3000
```

The server will automatically run database migrations on startup.

## 🧪 Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./internal/store

# Run tests with verbose output
go test -v ./...
```

### Database Operations

```bash
# Connect to main database
docker exec -it kabancount_db psql -U postgres -d kabancount

# Connect to test database
docker exec -it kabancount_test_db psql -U postgres -d kabancount_test

# Start only database services
docker-compose up -d db
docker-compose up -d test_db
```

### Build Commands

```bash
# Build application
go build

# Install dependencies
go mod tidy

# View dependency graph
go mod graph
```

## 📚 API Documentation

### Authentication

All protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

### Endpoints

#### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/healthcheck` | Service health status |
| POST | `/auth/signup` | Register new organization and admin user |
| POST | `/auth/signin` | Authenticate and get JWT token |

#### Protected Endpoints

**Users**
| Method | Endpoint | Description | Admin Only |
|--------|----------|-------------|------------|
| POST | `/users` | Create new user in organization | No |

**Organizations**
| Method | Endpoint | Description | Admin Only |
|--------|----------|-------------|------------|
| POST | `/organizations` | Create new organization | Yes |
| GET | `/organizations/{id}` | Get organization by ID | Yes |
| PUT | `/organizations/{id}` | Update organization | Yes |
| DELETE | `/organizations/{id}` | Delete organization | Yes |

**Items**
| Method | Endpoint | Description | Admin Only |
|--------|----------|-------------|------------|
| POST | `/items` | Create new item | No |
| GET | `/items` | List items with pagination | No |
| GET | `/items/{id}` | Get item by ID | No |
| PUT | `/items/{id}` | Update item | No |
| DELETE | `/items/{id}` | Delete item | No |

**Categories**
| Method | Endpoint | Description | Admin Only |
|--------|----------|-------------|------------|
| POST | `/categories` | Create new category | No |
| GET | `/categories` | List categories with pagination | No |
| GET | `/categories/{id}` | Get category by ID | No |
| PUT | `/categories/{id}` | Update category | No |
| DELETE | `/categories/{id}` | Delete category | No |

### Request/Response Examples

#### Register Organization and Admin User

```bash
POST /auth/signup
Content-Type: application/json

{
  "company_name": "Acme Corp",
  "username": "admin",
  "email": "admin@acme.com",
  "password": "SecurePass123!",
  "bio": "System Administrator"
}
```

#### Sign In

```bash
POST /auth/signin
Content-Type: application/json

{
  "username": "admin",
  "password": "SecurePass123!"
}

Response:
{
  "token": {
    "plaintext": "jwt-token-here",
    "expiry": "2024-01-02T15:04:05Z"
  }
}
```

#### Create Category

```bash
POST /categories
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "name": "Electronics",
  "description": "Electronic items and components"
}
```

#### Create Item

```bash
POST /items
Authorization: Bearer <jwt-token>
Content-Type: application/json

{
  "category_id": "uuid-here",
  "name": "Laptop Computer",
  "description": "High-performance laptop",
  "sku": "LAP-001",
  "unit_price": 99999,
  "reorder_level": 10
}
```

#### List Items with Pagination

```bash
GET /items?page=1&page_size=20
Authorization: Bearer <jwt-token>

Response:
{
  "data": [...],
  "count": 15,
  "total": 150,
  "page": 1,
  "page_size": 20
}
```

## 🗃 Database Schema

### Core Tables

- **organizations**: Company/tenant isolation
- **users**: User accounts with role-based access
- **categories**: Item categorization
- **items**: Inventory items with pricing and stock info
- **tokens**: JWT token management
- **stocks**: Stock tracking (planned)

### Key Relationships

- Users belong to organizations (multi-tenant)
- Items belong to categories and organizations
- Categories are scoped to organizations
- Admin users can manage organization settings

## 🔧 Configuration

The application uses environment-driven configuration with sensible defaults:

- **Server**: Configurable port (default: 8080)
- **Database**: Full PostgreSQL connection configuration
- **JWT**: Configurable secret and token expiration
- **Environment**: Development/production mode switching

Required environment variables:
- `JWT_SECRET`: JWT signing secret
- `DATABASE_NAME`: Database name
- `DATABASE_USER`: Database username
- `DATABASE_PASSWORD`: Database password

## 🏗 Architecture

### Project Structure

```
├── main.go                    # Application entry point
├── internal/
│   ├── app/                   # Application bootstrap and DI
│   ├── api/                   # HTTP handlers
│   ├── store/                 # Data access layer
│   ├── routes/                # Route definitions
│   ├── middleware/            # Authentication middleware
│   ├── config/                # Configuration management
│   ├── tokens/                # JWT utilities
│   ├── cookie/                # Cookie management
│   ├── pagination/            # Pagination utilities
│   └── utils/                 # General utilities
├── migrations/                # Database migrations
└── docker-compose.yml         # Database containers
```

### Design Principles

- **Clean Architecture**: Clear separation between handlers, business logic, and data access
- **Dependency Injection**: Handlers receive their dependencies through constructors
- **Multi-tenancy**: Organization-based data isolation
- **Security**: Role-based access control and JWT authentication
- **Testability**: Structured for easy unit and integration testing

## 🚦 Development Status

This project is in active development with a focus on:

1. **Current Phase**: Inventory management foundation
2. **Next Phase**: E-commerce platform features
3. **Future Phases**: Work orders, costing, and advanced analytics

## 🤝 Contributing

This is primarily a learning project, but contributions and suggestions are welcome. Please ensure any contributions maintain the clean architecture principles and include appropriate tests.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For questions or issues, please [create an issue](link-to-issues) in the repository.