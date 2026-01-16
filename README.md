# WebCoreGo - Pluggable API Framework

A RESTful API framework built with Go and Fiber featuring a pluggable architecture that separates infrastructure from business logic.

## 🏗️ Architecture Overview

The WebCoreGo framework follows a **Pluggable Architecture** pattern that allows teams to work independently on isolated modules:

```
┌─────────────────────────────────────────────────────────────┐
│              Main Repository (repo-utama-api)               │
├─────────────────────────────────────────────────────────────┤
│  WebCore Engine  │  Global Middleware  │  Shared Libraries  │
│  - config        │  - auth.go          │  - database:       │
|  - logger        |  - logging.go       |     - postgres     |
|  - DI:           |  - rate_limit.go    |     - mysql        |
|     - libraries  |                     |     - sqlite       |
|     - modules    |                     |     - mongo        |
│  - middleware    │                     │  - redis           │
│  - helper        │                     │  - kafka           │
│                  │                     │  - pubsub          │
├─────────────────────────────────────────────────────────────┤
│                    Central Registry                         │
│                  (Module Management & DI)                   │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    Module Repositories                      │
│                 (Separate Git Repositories)                 │
├─────────────────────────────────────────────────────────────┤
│  Module A        │  Module B          │  Module C           │
│  - config        │  - config          │  - config           │
│  - handler       │  - handler         │  - handler          │
│  - service       │  - service         │  - service          │
│  - repository    │  - repository      │  - repository       │
│  - module.go     │  - module.go       │  - module.go        │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 Features

### Core Features
- **Pluggable Architecture**: Modules can be developed, tested, and deployed independently
- **Central Registry**: Automatic module registration and dependency injection
- **Go-Fiber Framework**: Fast, expressive, minimalist web framework for Go
- **Shared Libraries**: Common utilities, database connections, and Redis client
- **Global Middleware**: Authentication, logging, rate limiting, CORS
- **CI/CD Ready**: GitHub Actions workflows for testing and deployment

### Module Features
- **Isolated Development**: Each module has its own repository
- **Standard Interface**: All modules implement the same interface
- **Dependency Injection**: Shared dependencies injected into modules
- **Auto-Loading**: Modules can be loaded from various sources
- **Version Management**: Module version tracking and compatibility

## 📁 Project Structure

```
webcore-go/                          
├── webcore/
│   ├── go.mod                       # Go module definition
│   ├── go.sum                       # Go module checksums
│   ├── main.go                      # Application entry point
│   ├── deps/                        # Application modules and dependencies Management
│   │   ├── libraries.go             # List of library dependencies
│   │   └── packages.go              # List of module dependencies
│   └── app/                         # Core application logic
│       ├── config/                      # Configuration management
│       │   └── config.go                # Configuration loading
│       ├── core/                        # Core logic
│       │   └── app.go                   # Application main logic
│       │   ├── module.go                # Central registry implementation
│       │   └── loader.go                # Module loader implementation
│       ├── helper/                      # Some helper functions
│       │   ├── api.go                   # API
│       │   ├── json.go                  # Override default JSON Encoding/Decoding using goccy-json
│       │   ├── log.go                   # Log
│       │   ├── string.go                # String
│       │   ├── task.go                  # Task
│       │   └── utils.go                 # Some utility functions
│       ├── loader/                      # Dependency Injection interface
│       │   └── conn.go                  # Dependency Injection interface for Database, Kafka, Redis, PubSub etc
│       ├── logger/                      # Logger definition
│       │   └── logger.go                # Override default logger implementation
│       └── middleware/                  # Global middleware
│           ├── auth.go                  # Authentication middleware
│           ├── logging.go               # Logging middleware
│           ├── rate_limit.go            # Rate limiting middleware
│           └── middleware.go            # Middleware registration
├── libraries/                       # Global Shared Libraries and implement DI interface for Database, Kafka, Redis, PubSub etc  
│   ├── db/                          # Database
│   │   ├── mongo                    # MongoDB database implementation
│   │   ├── sql                      # Relational database abstraction
│   │   ├── mysql                    # MySQL database implementation
│   │   ├── sqlite                   # SQLite database implementation
│   │   └── postgres                 # PostgreSQL database implementation
│   ├── kafka/                       # Kafka
│   ├── pubsub/                      # PubSub
│   └── redis/                       # Redis
├── modules/
│   └── module-a/                    # Example module
│       ├── go.mod                   # Go module definition
│       ├── module.go                # Module implementation
│       ├── config/
│       │   └── config.go            # Module configuration
│       ├── handler/
│       │   └── handler.go           # HTTP handlers
│       ├── model/
│       │   ├── model1.go            # Model
│       │   └── model2.go            # Model
│       ├── service/
│       │   └── service.go           # Business logic
│       └── repository/
│           └── repository.go        # Data access layer
├── .github/
│   └── workflows/
│       ├── ci.yml                   # CI pipeline
│       └── cd.yml                   # CD pipeline
├── go.work                          # Go workspace configuration
├── go.work.sum                      # Go workspace configuration checksum
├── config.yaml                      # Configuration file
├── Dockerfile                       # Docker configuration
├── docker-compose.yml               # Docker Compose configuration
└── run.sh                           # script to run go webcore/main.go
```

## 🛠️ Installation

### Prerequisites
- Go 1.19 or higher
- PostgreSQL, MySQL, SQLite or MongoDB (for database)
- Redis (optional, for caching)
- Docker (optional, for containerization)

### Local Development

1. **Clone the repository**:
```bash
git clone <repository-url>
cd webcore-go
```

2. **Install dependencies**:
```bash
go mod tidy
```

3. **Set up configuration**:
```bash
cp config.yaml.example config.yaml
# Edit config.yaml with your settings
```

4. **Set up database** (optional):
```bash
# Create PostgreSQL database
createdb konsolidator

# Run migrations (if any)
go run main.go migrate
```

5. **Run the application**:
```bash
go run webcore/main.go
```
Or use run.sh from root directory
```bash
./run.sh
```

The application will start on `http://localhost:7272`

### Using Docker

1. **Build and run with Docker Compose**:
```bash
docker-compose up -d
```

2. **Build manually**:
```bash
docker build -t konsolidator .
docker run -p 7272:7272 konsolidator
```

## 📖 Configuration

The application configuration is managed through `config/config.yaml`:

```yaml
app:
  name: "konsolidator-api"
  version: "1.0.0"
  environment: "development"

server:
  host: "0.0.0.0"
  port: 7272
  read_timeout: 30
  write_timeout: 30

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  database: "konsolidator"
  ssl_mode: "disable"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret_key: "your-secret-key"
  expires_in: 86400  # 24 hours in seconds

modules:
  base_path: "./modules"
  disabled:
    - "modulea"
```

## 📚 Development

### Library Development

Learn how to develop shared libraries for the WebCoreGo framework. Libraries are reusable components that provide shared functionality across modules and the core application.

For detailed guidance on creating libraries, see [Library Development Documentation](docs/library-development.md).

### Module Development

Learn how to develop pluggable modules for the WebCoreGo framework. Modules are self-contained units of functionality that can be developed, tested, and deployed independently.

For detailed guidance on creating modules, see [Module Development Documentation](docs/module-development.md).

1. **Create a new repository** for your module:
```bash
cd modules
git clone <module-template> module-b
cd module-b
```

2. **Implement the Module interface**:
```go
package moduleb

import (
    "github.com/gofiber/fiber/v2"
    "github.com/webcore-go/webcore/app/registry"
)

type Module struct {
    // Your module fields
}

func (m *Module) Name() string {
    return "module-b"
}

func (m *Module) Version() string {
    return "1.0.0"
}

// Dependencies returns the dependencies of the module to other modules
func (m *Module) Dependencies() []string {
	return []string{}
}

func (m *Module) Init(app *fiber.App, deps *module.Context) error {
    // Initialize your module
    return nil
}

func (m *Module) Destroy() error {
	return nil
}

func (m *Module) Config() appConfig.Configurable {
	return m.config
}

func (m *Module) Routes() []*fiber.Route {
    // Return your routes
    return []*fiber.Route{}
}

func (m *Module) Middleware() []fiber.Handler {
    // Return your middleware
    return []fiber.Handler{}
}

func (m *Module) Services() map[string]any {
    // Return your services
    return map[string]any{}
}

func (m *Module) Repositories() map[string]any {
    // Return your repositories
    return map[string]any{}
}
```

3. **Register your module** in the APP_PACKAGES located in deps/packages.go:
```go
var APP_PACKAGES = []core.Module{
	modulea.NewModule(),

// Add your packages here
	moduleb.NewModule(), // your module
}
```

### Module Structure

Each module should follow this structure:

```
module-b/
├── go.mod                   # Go module definition
├── module.go                # Module implementation
├── config/
│   └── config.go            # Module configuration
├── handler/
│   └── handler.go           # HTTP handlers
├── model/
│   ├── model1.go            # Model
│   └── model2.go            # Model
├── service/
│   └── service.go           # Business logic
└── repository/
    └── repository.go        # Data access layer
```

### Module Dependencies

Modules can depend on shared libraries:

```go
import (
    "github.com/webcore-go/webcore/modules/module-a/repository"
)

type Module struct {
    db     *loader.IDatabase
    redis  *loader.IRedis
}

func (m *Module) Init(app *fiber.App, deps *module.Context) error {
    m.db = deps.Database
    m.redis = deps.Redis
    m.logger = deps.Logger
    return nil
}
```

## 🔄 Module Loading

### Auto-Loading

Modules can be automatically loaded from configured directories:

```yaml
modules:
  base_path: "./modules"
  disabled:
    - "modulea"
```

### Manual Loading

Modules can also be loaded programmatically:

```go
// Load from file path
err := manager.LoadModuleFromPath("/path/to/module.so")

// Load from git repository
err := manager.LoadModuleFromGit(
    "https://github.com/user/module-b.git",
    "main",
    "./module-b",
)

// Load from go module path
err := manager.LoadModuleFromPackage("github.com/user/module-b")
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/registry
```

### Testing Modules

Each module should have its own test suite:

```bash
# Test specific module
cd packages/module-a
go test ./...
```

## 🚀 Deployment

### Build for Production

```bash
# Build the application
go build -o konsolidator main.go

# Build with optimizations
go build -ldflags="-s -w" -o konsolidator main.go
```

### Docker Deployment

```bash
# Build image
docker build -t konsolidator:latest .

# Run container
docker run -d -p 7272:7272 konsolidator:latest
```

### Kubernetes Deployment

Example Kubernetes configuration:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: konsolidator
spec:
  replicas: 3
  selector:
    matchLabels:
      app: konsolidator
  template:
    metadata:
      labels:
        app: konsolidator
    spec:
      containers:
      - name: konsolidator
        image: konsolidator:latest
        ports:
        - containerPort: 7272
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: REDIS_HOST
          value: "redis-service"
```

## 📊 API Documentation

### Health Check

```bash
GET /health
```

Response:
```json
{
  "status": "ok",
  "service": "konsolidator-api"
}
```

### API Version

```bash
GET /api/v1
```

Response:
```json
{
  "version": "1.0.0",
  "modules": ["module-a", "module-b"],
  "environment": "development"
}
```

### Module Endpoints

Each module can expose its own endpoints. For example, the `module-a` user management module:

```bash
# Get users with pagination
GET /api/v1/module-a/users?page=1&page_size=10

# Get user by ID
GET /api/v1/module-a/users/1

# Create user
POST /api/v1/module-a/users
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com"
}

# Update user
PUT /api/v1/module-a/users/1
Content-Type: application/json

{
  "name": "John Updated"
}

# Delete user
DELETE /api/v1/module-a/users/1
```

## 📖 Development Guidelines

### Code Style

- Follow Go standard code style
- Use meaningful variable and function names
- Add comments for complex logic
- Use interfaces for dependency injection

### Error Handling

- Use structured error responses
- Log errors appropriately
- Handle errors gracefully
- Provide meaningful error messages

### Security

- Validate all input data
- Use parameterized queries for database operations
- Implement proper authentication and authorization
- Sanitize user input

### Performance

- Use connection pooling for database and Redis
- Implement proper caching strategies
- Monitor and optimize database queries
- Use appropriate HTTP status codes

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions:
- Create an issue on GitHub
- Check the documentation
- Review the example modules
- Join the community discussions

## 🗺️ Roadmap

- [ ] Module hot-reloading
- [ ] Advanced dependency injection
- [ ] Configuration management for modules
- [ ] Monitoring and metrics
- [ ] Plugin marketplace
- [ ] API documentation generation
- [ ] GraphQL support
- [ ] WebSocket support
