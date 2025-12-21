# Quick Start Guide

This guide will help you get started with the WebCoreGo framework in minutes.

## Prerequisites

- Go 1.19 or higher
- Docker and Docker Compose (optional)
- Basic knowledge of Go and RESTful APIs

## Installation

### 1. Clone the Repository

```bash
git clone <repository-url>
cd webcore-go
```

### 2. Install Dependencies

```bash
go mod download
go mod tidy
```

### 3. Configure Environment

Create a `.env` file from the template:

```bash
cp .env.example .env
```

Edit the `.env` file with your configuration:

```bash
# Application settings
APP_ENV=development
APP_DEBUG=true
APP_PORT=3000

# Database settings
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=konsolidator

# Redis settings
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT settings
JWT_SECRET=your_jwt_secret_here
JWT_EXPIRES_IN=3600
```

## Quick Start with Docker

### 1. Start Services

```bash
docker-compose up -d
```

This will start:
- PostgreSQL database
- Redis cache
- The WebCoreGo API application

### 2. Run Migrations

```bash
go run cmd/migrate/main.go
```

### 3. Start the Application

```bash
go run main.go
```

The API will be available at `http://localhost:3000`

## Quick Start without Docker

### 1. Start PostgreSQL

```bash
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=konsolidator \
  -p 5432:5432 \
  postgres:13
```

### 2. Start Redis

```bash
docker run -d --name redis \
  -p 6379:6379 \
  redis:6-alpine
```

### 3. Run Migrations

```bash
go run cmd/migrate/main.go
```

### 4. Start the Application

```bash
go run main.go
```

## Test the API

### 1. Health Check

```bash
curl http://localhost:3000/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

### 2. Get Module Info

```bash
curl http://localhost:3000/info
```

Response:
```json
{
  "name": "WebCoreGo API",
  "version": "1.0.0",
  "modules": [
    {
      "name": "module-a",
      "version": "1.0.0",
      "status": "loaded"
    }
  ]
}
```

### 3. Test Module A Endpoints

#### Create an Item

```bash
curl -X POST http://localhost:3000/api/v1/module-a/items \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Item", "status": "active"}'
```

Response:
```json
{
  "id": 1,
  "name": "Test Item",
  "status": "active",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### Get Items

```bash
curl http://localhost:3000/api/v1/module-a/items
```

Response:
```json
{
  "data": [
    {
      "id": 1,
      "name": "Test Item",
      "status": "active",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

## Next Steps

### 1. Create Your First Module

Follow the [Module Development Guide](./module-development.md) to create your own module.

### 2. Explore the API

Check the [API Documentation](./api-documentation.md) for detailed endpoint information.

### 3. Deploy to Production

Follow the [Deployment Guide](./deployment.md) for production deployment instructions.

### 4. Customize Configuration

Edit `config/config.yaml` to customize the application configuration.

### 5. Add Authentication

Implement JWT authentication by adding the auth module and following the authentication guide.

## Troubleshooting

### Common Issues

1. **Port Already in Use**
   ```bash
   # Check what's using the port
   lsof -i :3000
   
   # Kill the process
   kill -9 <PID>
   ```

2. **Database Connection Failed**
   - Ensure PostgreSQL is running
   - Check database credentials in `.env`
   - Verify the database name exists

3. **Module Not Loading**
   - Check module implementation
   - Verify all required interface methods are implemented
   - Check for compilation errors

### Get Help

- Check the [README.md](../README.md) for detailed setup instructions
- Review the [troubleshooting section](./deployment.md#troubleshooting) in the deployment guide
- Contact the development team with specific error messages

## Example Usage

Here's a simple example of how to use the API:

```bash
#!/bin/bash

# Get all items
curl -s http://localhost:3000/api/v1/module-a/items | jq .

# Create a new item
curl -s -X POST http://localhost:3000/api/v1/module-a/items \
  -H "Content-Type: application/json" \
  -d '{"name": "New Item", "status": "active"}' | jq .

# Get a specific item
curl -s http://localhost:3000/api/v1/module-a/items/1 | jq .

# Update an item
curl -s -X PUT http://localhost:3000/api/v1/module-a/items/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Item", "status": "inactive"}' | jq .

# Delete an item
curl -s -X DELETE http://localhost:3000/api/v1/module-a/items/1 | jq .
```

## Contributing

We welcome contributions! Please follow these steps:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

For more information, see the [contribution guidelines](../CONTRIBUTING.md).

## License

This project is licensed under the MIT License. See the [LICENSE](../LICENSE) file for details.

---

Happy coding with WebCoreGo! 🚀