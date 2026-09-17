# Go Base Project

A production-ready Go HTTP server with PostgreSQL database connection, structured for scalability and maintainability.

## Features

- ✅ HTTP server running on port 8080
- ✅ PostgreSQL database connection with connection pooling
- ✅ Health check endpoint
- ✅ Modular project structure
- ✅ Request logging middleware
- ✅ Graceful shutdown
- ✅ Environment-based configuration

## Project Structure

```
go-base-project/
├── main.go              # Application entry point
├── config/              # Configuration management
│   └── config.go
├── controllers/         # HTTP request handlers
│   └── health_controller.go
├── models/             # Data models and schemas
├── routes/             # Route definitions
│   └── routes.go
├── services/           # Business logic and database
│   └── database.go
└── utils/              # Helper functions and middleware
    └── middleware.go
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher

## Setup

### 1. Clone the repository

```bash
git clone <repository-url>
cd go-base-project
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Setup PostgreSQL

Create a PostgreSQL database:

```sql
CREATE DATABASE go_base_db;
```

### 4. Configure environment variables

Copy the example environment file and update with your settings:

```bash
cp .env.example .env
```

Edit `.env` with your database credentials.

### 5. Run the application

```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Health Check

```bash
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "database": "connected",
  "message": "Server is running"
}
```

### Root Endpoint

```bash
GET /
```

**Response:**
```json
{
  "message": "Welcome to Go Base Project API"
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| SERVER_PORT | Server port | 8080 |
| DB_HOST | PostgreSQL host | localhost |
| DB_PORT | PostgreSQL port | 5432 |
| DB_USER | Database user | postgres |
| DB_PASSWORD | Database password | postgres |
| DB_NAME | Database name | go_base_db |
| DB_SSL_MODE | SSL mode | disable |

## Development

### Build the application

```bash
go build -o app main.go
```

### Run tests

```bash
go test ./...
```

### Format code

```bash
go fmt ./...
```

## License

MIT
