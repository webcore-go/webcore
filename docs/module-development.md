# Module Development Guide

This guide explains how to develop modules for the WebCoreGo framework.

## Understanding the Module Architecture

In WebCoreGo, modules are self-contained units of functionality that can be developed, tested, and deployed independently. Each module follows a clean architecture pattern:

```
Module
├── Handler (HTTP Layer)
├── Service (Business Logic Layer)
├── Repository (Data Access Layer)
└── Models (Data Structures)
```

## Creating a New Module

### 1. Module Structure

Create a new module with the following structure:

```
my-module/
├── module.go              # Module implementation
├── handler/
│   └── handler.go          # HTTP handlers
├── service/
│   └── service.go          # Business logic
├── repository/
│   └── repository.go      # Data access layer
├── models/
│   └── models.go          # Data models
└── go.mod                 # Module dependencies
```

### 2. Module Interface

Every module must implement the `module.Module` interface:

```go
package mymodule

import (
    "github.com/gofiber/fiber/v2"
    "github.com/semanggilab/webcore-go/app/registry"
)

type Module struct {
    // Your module fields
}

// Name returns the unique name of the module
func (m *Module) Name() string {
    return "my-module"
}

// Version returns the version of the module
func (m *Module) Version() string {
    return "1.0.0"
}

// Init initializes the module with the given app and dependencies
func (m *Module) Init(app *fiber.App, deps *module.Context) error {
    // Initialize your module components
    return nil
}

// Routes returns the routes provided by this module
func (m *Module) Routes() []*fiber.Route {
    // Return your routes
    return []*fiber.Route{}
}

// Middleware returns the middleware provided by this module
func (m *Module) Middleware() []fiber.Handler {
    // Return your middleware
    return []fiber.Handler{}
}

// Services returns the services provided by this module
func (m *Module) Services() map[string]any {
    // Return your services
    return map[string]any{}
}

// Repositories returns the repositories provided by this module
func (m *Module) Repositories() map[string]any {
    // Return your repositories
    return map[string]any{}
}
```

### 3. Handler Layer

The handler layer manages HTTP requests and responses:

```go
package handler

import (
    "strconv"
    
    "github.com/gofiber/fiber/v2"
    "github.com/semanggilab/webcore-go/app/registry"
    "github.com/semanggilab/webcore-go/app/shared"
    "github.com/semanggilab/webcore-go/packages/mymodule/service"
)

type Handler struct {
    logger      *shared.Logger
    userService service.UserService
}

func NewHandler(deps *module.Context, userService service.UserService) *Handler {
    return &Handler{
        logger:      deps.Logger,
        userService: userService,
    }
}

// GetItems returns all items
func (h *Handler) GetItems(c *fiber.Ctx) error {
    page, _ := strconv.Atoi(c.Query("page", "1"))
    pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))
    
    items, total, err := h.userService.GetItems(c.Context(), page, pageSize)
    if err != nil {
        return h.handleError(c, err)
    }
    
    paginatedItems, pagination := shared.Paginate(items, page, pageSize)
    return c.JSON(shared.NewPaginatedResponse(paginatedItems, pagination))
}

// handleError handles errors and returns appropriate HTTP responses
func (h *Handler) handleError(c *fiber.Ctx, err error) error {
    h.logger.Error("API error", "error", err.Error())
    
    if apiErr, ok := err.(*shared.APIError); ok {
        return c.Status(apiErr.Code).JSON(shared.NewErrorResponse(apiErr.Message))
    }
    
    return c.Status(fiber.StatusInternalServerError).JSON(shared.NewErrorResponse("Internal server error"))
}
```

### 4. Service Layer

The service layer contains business logic:

```go
package service

import (
    "context"
    
    "github.com/semanggilab/webcore-go/app/registry"
    "github.com/semanggilab/webcore-go/app/shared"
    "github.com/semanggilab/webcore-go/packages/mymodule/repository"
)

// ItemService defines the interface for item operations
type ItemService interface {
    GetItems(ctx context.Context, page, pageSize int) ([]map[string]any, int, error)
    GetItem(ctx context.Context, id int) (map[string]any, error)
    CreateItem(ctx context.Context, item map[string]any) (map[string]any, error)
    UpdateItem(ctx context.Context, id int, item map[string]any) (map[string]any, error)
    DeleteItem(ctx context.Context, id int) error
}

// Service represents the service layer
type Service struct {
    logger *shared.Logger
    itemRepository repository.ItemRepository
}

// NewService creates a new Service instance
func NewService(deps *module.Context, itemRepository repository.ItemRepository) *Service {
    return &Service{
        logger:         deps.Logger,
        itemRepository: itemRepository,
    }
}

// GetItems retrieves items with pagination
func (s *Service) GetItems(ctx context.Context, page, pageSize int) ([]map[string]any, int, error) {
    items, total, err := s.itemRepository.GetItems(ctx, page, pageSize)
    if err != nil {
        return nil, 0, err
    }
    
    return items, total, nil
}

// GetItem retrieves an item by ID
func (s *Service) GetItem(ctx context.Context, id int) (map[string]any, error) {
    item, err := s.itemRepository.GetItem(ctx, id)
    if err != nil {
        return nil, err
    }
    
    return item, nil
}

// CreateItem creates a new item
func (s *Service) CreateItem(ctx context.Context, item map[string]any) (map[string]any, error) {
    // Validate input
    if name, ok := item["name"].(string); !ok || name == "" {
        return nil, &shared.APIError{
            Code:    400,
            Message: "Name is required",
        }
    }
    
    // Call repository layer
    newItem, err := s.itemRepository.CreateItem(ctx, item)
    if err != nil {
        return nil, err
    }
    
    s.logger.Info("Item created", "item_id", newItem["id"])
    return newItem, nil
}

// UpdateItem updates an existing item
func (s *Service) UpdateItem(ctx context.Context, id int, item map[string]any) (map[string]any, error) {
    // Validate input
    if name, ok := item["name"].(string); ok && name == "" {
        return nil, &shared.APIError{
            Code:    400,
            Message: "Name cannot be empty",
        }
    }
    
    // Call repository layer
    updatedItem, err := s.itemRepository.UpdateItem(ctx, id, item)
    if err != nil {
        return nil, err
    }
    
    s.logger.Info("Item updated", "item_id", id)
    return updatedItem, nil
}

// DeleteItem deletes an item
func (s *Service) DeleteItem(ctx context.Context, id int) error {
    // Call repository layer
    err := s.itemRepository.DeleteItem(ctx, id)
    if err != nil {
        return err
    }
    
    s.logger.Info("Item deleted", "item_id", id)
    return nil
}
```

### 5. Repository Layer

The repository layer handles data access:

```go
package repository

import (
    "context"
    "time"
    
    "github.com/semanggilab/webcore-go/app/shared"
    "gorm.io/gorm"
)

// ItemRepository defines the interface for item operations
type ItemRepository interface {
    GetItems(ctx context.Context, page, pageSize int) ([]map[string]any, int, error)
    GetItem(ctx context.Context, id int) (map[string]any, error)
    CreateItem(ctx context.Context, item map[string]any) (map[string]any, error)
    UpdateItem(ctx context.Context, id int, item map[string]any) (map[string]any, error)
    DeleteItem(ctx context.Context, id int) error
}

// Repository represents the repository layer
type Repository struct {
    db *gorm.DB
}

// NewRepository creates a new Repository instance
func NewRepository(db *gorm.DB) *Repository {
    return &Repository{
        db: db,
    }
}

// Item represents the item model
type Item struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" gorm:"size:100;not null"`
    Status    string    `json:"status" gorm:"size:20;default:'active'"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for the Item model
func (Item) TableName() string {
    return "items"
}

// GetItems retrieves items with pagination
func (r *Repository) GetItems(ctx context.Context, page, pageSize int) ([]map[string]any, int, error) {
    var items []Item
    var total int64
    
    // Get total count
    if err := r.db.Model(&Item{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // Get paginated items
    offset := (page - 1) * pageSize
    if err := r.db.Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
        return nil, 0, err
    }
    
    // Convert to []map[string]any
    result := make([]map[string]any, len(items))
    for i, item := range items {
        result[i] = map[string]any{
            "id":         item.ID,
            "name":       item.Name,
            "status":     item.Status,
            "created_at": item.CreatedAt,
            "updated_at": item.UpdatedAt,
        }
    }
    
    return result, int(total), nil
}

// GetItem retrieves an item by ID
func (r *Repository) GetItem(ctx context.Context, id int) (map[string]any, error) {
    var item Item
    if err := r.db.First(&item, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, &shared.APIError{
                Code:    404,
                Message: "Item not found",
            }
        }
        return nil, err
    }
    
    return map[string]any{
        "id":         item.ID,
        "name":       item.Name,
        "status":     item.Status,
        "created_at": item.CreatedAt,
        "updated_at": item.UpdatedAt,
    }, nil
}

// CreateItem creates a new item
func (r *Repository) CreateItem(ctx context.Context, item map[string]any) (map[string]any, error) {
    newItem := Item{
        Name:   item["name"].(string),
        Status: "active",
    }
    
    if err := r.db.Create(&newItem).Error; err != nil {
        return nil, &shared.APIError{
            Code:    400,
            Message: "Failed to create item",
            Details: err.Error(),
        }
    }
    
    return map[string]any{
        "id":         newItem.ID,
        "name":       newItem.Name,
        "status":     newItem.Status,
        "created_at": newItem.CreatedAt,
        "updated_at": newItem.UpdatedAt,
    }, nil
}

// UpdateItem updates an existing item
func (r *Repository) UpdateItem(ctx context.Context, id int, item map[string]any) (map[string]any, error) {
    var existingItem Item
    if err := r.db.First(&existingItem, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, &shared.APIError{
                Code:    404,
                Message: "Item not found",
            }
        }
        return nil, err
    }
    
    // Update fields
    if name, ok := item["name"].(string); ok {
        existingItem.Name = name
    }
    if status, ok := item["status"].(string); ok {
        existingItem.Status = status
    }
    
    if err := r.db.Save(&existingItem).Error; err != nil {
        return nil, &shared.APIError{
            Code:    400,
            Message: "Failed to update item",
            Details: err.Error(),
        }
    }
    
    return map[string]any{
        "id":         existingItem.ID,
        "name":       existingItem.Name,
        "status":     existingItem.Status,
        "created_at": existingItem.CreatedAt,
        "updated_at": existingItem.UpdatedAt,
    }, nil
}

// DeleteItem deletes an item
func (r *Repository) DeleteItem(ctx context.Context, id int) error {
    var item Item
    if err := r.db.First(&item, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return &shared.APIError{
                Code:    404,
                Message: "Item not found",
            }
        }
        return err
    }
    
    if err := r.db.Delete(&item).Error; err != nil {
        return &shared.APIError{
            Code:    400,
            Message: "Failed to delete item",
            Details: err.Error(),
        }
    }
    
    return nil
}
```

### 6. Module Registration

Finally, register your module in the main application:

```go
// In module.go
func (m *Module) Init(app *fiber.App, deps *module.Context) error {
    // Initialize module components
    itemRepository := repository.NewRepository(deps.Database.DB)
    m.itemService = service.NewService(deps, itemRepository)
    itemHandler := handler.NewHandler(deps, m.itemService)
    
    // Register routes
    m.registerRoutes(app, itemHandler)
    
    return nil
}

func (m *Module) registerRoutes(app *fiber.App, handler *handler.Handler) {
    // Item routes
    itemGroup := app.Group("/api/v1/mymodule/items")
    
    itemGroup.Get("/", handler.GetItems)
    itemGroup.Post("/", handler.CreateItem)
    itemGroup.Get("/:id", handler.GetItem)
    itemGroup.Put("/:id", handler.UpdateItem)
    itemGroup.Delete("/:id", handler.DeleteItem)
    
    // Module-specific routes
    app.Get("/api/v1/mymodule/health", handler.ModuleHealth)
    app.Get("/api/v1/mymodule/info", handler.ModuleInfo)
}
```

## Module Dependencies

### Using Shared Dependencies

Your module can use shared dependencies:

```go
type Module struct {
    db     *shared.Database
    redis  *shared.Redis
    logger *shared.Logger
    eventBus *shared.EventBus
}

func (m *Module) Init(app *fiber.App, deps *module.Context) error {
    m.db = deps.Database
    m.redis = deps.Redis
    m.logger = deps.Logger
    m.eventBus = deps.EventBus
    
    // Use shared dependencies
    if m.db != nil {
        // Use database
    }
    
    if m.redis != nil {
        // Use Redis
    }
    
    return nil
}
```

### Inter-Module Communication

Modules can communicate through the central registry:

```go
// In module A
func (m *Module) Init(app *fiber.App, deps *module.Context) error {
    // Get service from another module
    if service, exists := deps.Services["moduleB.userService"]; exists {
        // Use the service from module B
    }
    
    return nil
}
```

## Testing Your Module

### Unit Tests

```go
// repository_test.go
package repository

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestItemRepository_GetItems(t *testing.T) {
    // Setup test database
    db := setupTestDatabase(t)
    repo := NewRepository(db)
    
    // Test data
    items := []Item{
        {Name: "Item 1", Status: "active"},
        {Name: "Item 2", Status: "active"},
    }
    
    // Insert test data
    for _, item := range items {
        db.Create(&item)
    }
    
    // Test
    result, total, err := repo.GetItems(context.Background(), 1, 10)
    
    // Assertions
    require.NoError(t, err)
    assert.Equal(t, len(items), total)
    assert.Len(t, result, len(items))
}

// service_test.go
package service

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestItemService_CreateItem(t *testing.T) {
    // Setup
    deps := setupTestContext()
    repo := &mockRepository{}
    service := NewService(deps, repo)
    
    // Test data
    item := map[string]any{
        "name": "Test Item",
    }
    
    // Test
    result, err := service.CreateItem(context.Background(), item)
    
    // Assertions
    require.NoError(t, err)
    assert.NotEmpty(t, result["id"])
    assert.Equal(t, item["name"], result["name"])
}
```

### Integration Tests

```go
// integration_test.go
package handler

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    
    "github.com/gofiber/fiber/v2"
    "github.com/stretchr/testify/assert"
)

func TestHandler_GetItems(t *testing.T) {
    // Setup
    app := fiber.New()
    handler := setupTestHandler()
    
    // Register routes
    app.Get("/api/v1/mymodule/items", handler.GetItems)
    
    // Test
    req := httptest.NewRequest(http.MethodGet, "/api/v1/mymodule/items", nil)
    resp, err := app.Test(req)
    
    // Assertions
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## Module Configuration

### Environment Variables

Your module can read configuration from environment variables:

```go
type Config struct {
    DatabaseURL string `env:"DATABASE_URL"`
    RedisURL    string `env:"REDIS_URL"`
    LogLevel    string `env:"LOG_LEVEL" default:"info"`
}

func LoadConfig() (*Config, error) {
    cfg := &Config{}
    
    if err := env.Parse(cfg); err != nil {
        return nil, err
    }
    
    return cfg, nil
}
```

### Module-Specific Configuration

Add module configuration to the main config:

```yaml
# config/config.yaml
mymodule:
  database_table: "my_items"
  cache_enabled: true
  cache_ttl: 300
```

```go
type ModuleConfig struct {
    DatabaseTable string `mapstructure:"database_table"`
    CacheEnabled  bool   `mapstructure:"cache_enabled"`
    CacheTTL      int    `mapstructure:"cache_ttl"`
}
```

## Deployment

### Building Your Module

```bash
# Build as a plugin
go build -buildmode=plugin -o mymodule.so ./mymodule

# Build as a standalone module
go build -o mymodule ./mymodule
```

### Loading Your Module

```go
// Load plugin module
err := centralRegistry.LoadModuleFromPath("./mymodule.so")

// Or register directly
module := mymodule.NewModule()
err := centralRegistry.Register(module)
```

## Best Practices

### 1. Keep Modules Focused

- Each module should have a single responsibility
- Avoid creating "god modules" that do everything
- Keep module boundaries clear and well-defined

### 2. Use Interfaces

- Define interfaces for your services
- Use dependency injection
- Make your modules testable

### 3. Handle Errors Gracefully

- Use structured error responses
- Log errors appropriately
- Provide meaningful error messages

### 4. Validate Input

- Validate all input data
- Use validation libraries
- Sanitize user input

### 5. Use Shared Context

- Leverage the shared database and Redis connections
- Use the shared logger and event bus
- Follow the established patterns

### 6. Write Tests

- Write unit tests for your services and repositories
- Write integration tests for your handlers
- Aim for high test coverage

### 7. Document Your Module

- Provide clear documentation
- Document your API endpoints
- Include examples

### 8. Version Your Module

- Use semantic versioning
- Maintain backward compatibility when possible
- Document breaking changes

## Troubleshooting

### Common Issues

1. **Module Not Loading**
   - Check that the module implements all required interface methods
   - Verify the module name is unique
   - Check for compilation errors

2. **Dependency Issues**
   - Verify all dependencies are available
   - Check that shared dependencies are properly initialized
   - Ensure database and Redis connections are working

3. **Route Conflicts**
   - Use unique route prefixes for each module
   - Check for overlapping routes
   - Use route groups to organize endpoints

4. **Performance Issues**
   - Use database connection pooling
   - Implement proper caching
   - Monitor query performance

### Debug Mode

Enable debug mode for detailed logging:

```go
// In your module
func (m *Module) Init(app *fiber.App, deps *module.Context) error {
    if deps.Config.App.Environment == "development" {
        deps.Logger.SetLevel("debug")
    }
    return nil
}
```

## Conclusion

Developing modules for WebCoreGo allows you to build scalable, maintainable applications with clear separation of concerns. Follow these guidelines to create high-quality modules that integrate seamlessly with the framework.

Remember to:
- Keep modules focused and independent
- Use interfaces and dependency injection
- Write comprehensive tests
- Document your code
- Follow the established patterns

Happy coding!