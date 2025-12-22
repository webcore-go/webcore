# Module Development Guide

This guide explains how to develop modules for the WebCoreGo framework.

## Understanding the Module Architecture

In WebCoreGo, modules are self-contained units of functionality that can be developed, tested, and deployed independently. Each module follows a clean architecture pattern:

```
Module
├── Config (Module Configuration)
├── Handler (HTTP Layer)
├── Service (Business Logic Layer)
├── Repository (Data Access Layer)
└── Models (Data Structures)
└── module.go (Module entry point)
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

Every module must implement the `core.Module` interface:

```go
package mymodule

import (
    "github.com/gofiber/fiber/v2"
    "github.com/semanggilab/webcore-go/app/core"
    "github.com/semanggilab/webcore-go/app/loader"
    appConfig "github.com/semanggilab/webcore-go/app/config"
)

const (
	ModuleName    = "modulea"
	ModuleVersion = "1.0.0"
)

type Module struct {
    config     *config.ModuleConfig
    // Your module fields
    repository repository.Repository
    service    service.Service
    handler    *handler.Handler
    routes     []*core.ModuleRoute
}

// NewModule creates a new Module instance
func NewModule() *Module {
	return &Module{}
}

// Name returns the unique name of the module
func (m *Module) Name() string {
	return ModuleName
}

// Version returns the version of the module
func (m *Module) Version() string {
	return ModuleVersion
}

// Dependencies returns the dependencies of the module to other modules
func (m *Module) Dependencies() []string {
	return []string{}
}

// Init initializes the module with the given app and dependencies
func (m *Module) Init(ctx *core.AppContext) error {
    // Load configuration into ModuleConfig (bind to key)
    m.config = &config.ModuleConfig{}
    if err := core.LoadDefaultConfig(m.Name(), m.config); err != nil {
        return err
    }

    libmanager := core.Instance().LibraryManager
    lName := "db:" + ctx.Config.Database.Driver

    // Load singleton library via core.LibraryManager.GetSingleton
    if lib, ok := libmanager.GetSingleton(lName); ok {
        db := lib.(loader.IDatabase)
        // Initialize your module components
        m.repository = repository.NewRepository(db)
        m.service = service.NewService(ctx, m.repository)
        m.handler = handler.NewHandler(ctx, m.service)
    }

    // Register routes
    m.registerRoutes(ctx.Root)

    return nil
}

func (m *Module) Destroy() error {
	return nil
}

func (m *Module) Config() appConfig.Configurable {
    // Return your configuration that inherits from config.ModuleConfig
	return m.config
}

// Routes returns the routes provided by this module
func (m *Module) Routes() []*core.ModuleRoute {
    return m.routes
}

// Middleware returns the middleware provided by this module
func (m *Module) Middleware() []fiber.Handler {
    // Return your middleware
    return []fiber.Handler{}
}

// Services returns the services provided by this module
func (m *Module) Services() map[string]any {
    // Return services that can be used by other modules
    return map[string]any{
        "service": m.service,
    }
}

// Repositories returns the repositories provided by this module
func (m *Module) Repositories() map[string]any {
    // Return repositories that can be used by other modules
    return map[string]any{
        "repository": m.repository,
    }
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
    "github.com/semanggilab/webcore-go/modules/mymodule/service"
)

type Handler struct {
    userService service.UserService
}

func NewHandler(deps *module.Context, userService service.UserService) *Handler {
    return &Handler{
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
    "github.com/semanggilab/webcore-go/modules/mymodule/repository"
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
    itemRepository repository.ItemRepository
}

// NewService creates a new Service instance
func NewService(deps *module.Context, itemRepository repository.ItemRepository) *Service {
    return &Service{
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

#### Registering Module in webcore/deps/packages.go

To ensure your module is properly loaded and its dependencies are triggered safely, register your module in `webcore/deps/packages.go`:

```go
// webcore/deps/packages.go
package deps

import (
    "github.com/semanggilab/webcore-go/app/core"
    mymodule "github.com/semanggilab/webcore-go/modules/mymodule"
)

var APP_PACKAGES = []core.Module{
    mymodule.NewModule(),

    // Add your packages here
}
```

#### Module Initialization with Configuration Inheritance

Your module should inherit configuration from `config.ModuleConfig` and load from `Init()` function:

```go
// In module.go
func (m *Module) Init(ctx *core.AppContext) error {
    // Load configuration into ModuleConfig (bind to key)
    m.config = &config.ModuleConfig{}
    if err := core.LoadDefaultConfig(m.Name(), m.config); err != nil {
        return err
    }

    return nil
}
```

#### Route Registration

Register your module's routes using the custom `registerRoutes` function and call it from `Init()`.

```go
// registerRoutes registers the module's routes
func (m *Module) registerRoutes(root fiber.Router) {
    // Module routes
    moduleRoot := root.Group("/" + m.Name())

    // Business logic routes
    m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
        Method:  "GET",
        Path:    "/items",
        Handler: m.handler.GetItems,
        Root:    moduleRoot,
    })
    m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
        Method:  "POST",
        Path:    "/items",
        Handler: m.handler.CreateItem,
        Root:    moduleRoot,
    })
    m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
        Method:  "GET",
        Path:    "/items/:id",
        Handler: m.handler.GetItem,
        Root:    moduleRoot,
    })
    m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
        Method:  "PUT",
        Path:    "/items/:id",
        Handler: m.handler.UpdateItem,
        Root:    moduleRoot,
    })
    m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
        Method:  "DELETE",
        Path:    "/items/:id",
        Handler: m.handler.DeleteItem,
        Root:    moduleRoot,
    })

    // Optional: Health and Info endpoints
    m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
        Method:  "GET",
        Path:    "/health",
        Handler: m.Health,
        Root:    moduleRoot,
    })

    m.routes = core.AppendRouteToArray(m.routes, &core.ModuleRoute{
        Method:  "GET",
        Path:    "/info",
        Handler: m.Info,
        Root:    moduleRoot,
    })
}
```

You are recommend to add the following route to `root.Group("/" + module.Name())` to make clean module-based path.
```go
// In module.go
func (m *Module) Init(ctx *core.AppContext) error {
    
    // Register routes
    m.registerRoutes(ctx.Root)

    return nil
}
```

#### Optional: Health and Info Endpoints

You can add health and info endpoints to provide module status information:

```go
// ModuleHealth returns the health status of the module
func (m *Module) Health(c *fiber.Ctx) error {
    health := map[string]any{
        "status":    "healthy",
        "module":    ModuleName,
        "version":   ModuleVersion,
        "timestamp": time.Now().Format(time.RFC3339),
    }
    return c.JSON(health)
}

// ModuleInfo returns information about the module
func (m *Module) Info(c *fiber.Ctx) error {
    endpoints := []string{}
    for _, endpoint := range m.routes {
        endpointStr := endpoint.Method + " " + endpoint.Path
        endpoints = append(endpoints, endpointStr)
    }

    path := "/" + ModuleName

    info := map[string]any{
        "name":        ModuleName,
        "version":     ModuleVersion,
        "description": "Your module description",
        "path":        path,
        "endpoints":   endpoints,
        "config":      m.config,
    }
    return c.JSON(info)
}
```

## Module Dependencies

### Load WebCore Libraries

```go
// In module.go
func (m *Module) Init(ctx *core.AppContext) error {
    // Get the library manager from the core instance
    libmanager := core.Instance().LibraryManager
    lName := "db:" + ctx.Config.Database.Driver

    // Load singleton library via core.LibraryManager.GetSingleton
    // The parameter is taken from the key in APP_LIBRARIES variable in webcore/deps/libraries.go
    if lib, ok := libmanager.GetSingleton(lName); ok {
        // shared library successfully loaded
        db := lib.(loader.IDatabase)

        // Initialize your module components
        m.repository = repository.NewRepository(db)
        m.service = service.NewService(ctx, m.repository)
        m.handler = handler.NewHandler(ctx, m.service)
    }

    return nil
}
```

WebCore Library must be import using standard golang dependency or put library repo into `/libraries/` folder or put `mylibrary.so` file into `./packages/` folder. To activate library you must register it in `APP_LIBRARIES` in `webcore/deps/libraries.go`.

```go
// webcore/deps/libraries.go
package deps

import (
	"github.com/semanggilab/webcore-go/app/core"
	"github.com/semanggilab/webcore-go/lib/mongo"
)

var APP_LIBRARIES = map[string]core.LibraryLoader{
	"db:mongodb": &mongo.MongoLoader{},

	// Add your library here
}

```

### Using Shared Modules

Your module can use shared dependencies (config, handler, repository, service, etc.) directly from other modules using standar import. You must ensure modules is registered in `webcore/deps/packages.go` or import as package from golang repository. Here is an example:

```go
// in modules/moduleb/config.go
package config

import (
	configa "github.com/semanggilab/webcore-go/modulesa/tb/config"
)

type ModuleConfig struct {
	TB  *configa.ModuleConfig // refer to config in modulea.ModuleConfig
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

### Configuration Inheritance from config.ModuleConfig

Your module should inherit configuration from `config.ModuleConfig` to ensure proper configuration management. Here's how to implement it:

```go
// config/config.go
package config

import (
    "github.com/semanggilab/webcore-go/app/config"
)

type ModuleConfig struct {
    // Your module-specific configuration. `mapstructure` needed for yaml binding
    DatabaseTable string `mapstructure:"database_table"`
    CacheEnabled  bool   `mapstructure:"cache_enabled"`
    CacheTTL      int    `mapstructure:"cache_ttl"`
    
    // Inherited from base config
    config.BaseConfig
}

// SetEnvBindings help to map environment variables to struct fields
// Map key must be start with prefix `module.<module_name>.<field_name>`
// <field_name> must be match with mapstructure tag in struct field
func (c *ModuleConfig) SetEnvBindings() map[string]string {
    return map[string]string{
        "module.mymodule.database_table": "MODULE_MYMODULE_DATABASE_TABLE",
        "module.mymodule.cache_enabled": "MODULE_MYMODULE_CACHE_ENABLED",
        "module.mymodule.cache_ttl":     "MODULE_MYMODULE_CACHE_TTL",
    }
}

// SetDefaults sets default values for configuration fields
// Map key use same as SetEnvBindings
func (c *ModuleConfig) SetDefaults() map[string]any {
    return map[string]any{
        "module.mymodule.database_table": "my_items",
        "module.mymodule.cache_enabled": true,
        "module.mymodule.cache_ttl":     300,
    }
}
```

In your module initialization:

```go
func (m *Module) Init(ctx *core.AppContext) error {
    // Load configuration into ModuleConfig (bind to key)
    m.config = &config.ModuleConfig{}
    if err := core.LoadDefaultConfig(m.Name(), m.config); err != nil {
        return err
    }
    
    // Now you can access configuration
    if m.config.CacheEnabled {
        // Initialize caching
    }
    
    return nil
}
```

### Environment Variables

Your module can read configuration from environment variables. Map between environment variables and struct fields defined in `SetEnvBindings()` function. Name of environment variable must be set as value of map `SetEnvBindings`. Format for environment variable must be start with prefix `MODULE_<MODULE_NAME>_<FIELD_NAME>` in capital letters.


### Module-Specific Configuration

Add module configuration to the main config or other yaml file. Configuration must be put inside field `module`.

```yaml
# modulea.yaml
module:
    # Your module-specific configuration
    modulea:
        database_table: "my_items"
        cache_enabled: true
        cache_ttl: 300
```

```yaml
# moduleb.yaml
module:
    # Your module-specific configuration
    moduleb:
        secret_key: "no-secret"
        secondary_db:
            driver: "postgres"
            host: "localhost"
            port: 5432
            user: "postgres"
```

Or put it all in main config:

```yaml
# config.yaml
module:
    modulea:
        database_table: "my_items"
        cache_enabled: true
        cache_ttl: 300 
    moduleb:
        secret_key: "no-secret"
        secondary_db:
            driver: "postgres"
            host: "localhost"
            port: 5432
            user: "postgres"
```
Load module from other yaml file must be placed in `Init()` funtion in `mymodule/module.go`.

```go
// In module.go
func (m *Module) Init(ctx *core.AppContext) error {
    m.config = &config.ModuleConfig{}

    // Load configuration from main file config.yaml
    if err := core.LoadDefaultConfigModule(m.Name(), m.config); err != nil {
        return err
    }

    // Load configuration from file `config-a.yaml`, empty search paths `[]string{}` will be assume search from working directory `.`
    if err := core.LoadConfigModule(m.Name(), m.config, "config-a", []string{}); err != nil {
        return err
    }

    return nil
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
// Option 1: Load plugin module manually
err := centralRegistry.LoadModuleFromPath("./mymodule.so")

// Or
// Option 2: register directly runtime
module := mymodule.NewModule()
err := centralRegistry.Register(module)
```

*Option 3* you can put compiled module `mymodule.so` file in directory `./modules/`

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

Enable debug mode for detailed logging edit `config/config.yaml`:

```yaml
app:
  logging:
    level: debug
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
