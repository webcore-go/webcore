# Module Development Guide

This guide explains how to develop modules for the WebCoreGo framework.

## Understanding the Module Architecture

In WebCoreGo, modules are self-contained units of functionality that can be developed, tested, and deployed independently. Each module follows a clean architecture pattern:

```
Module
├── Config (Module Configuration)
├── Handler (HTTP Layer, Messaging Consumer, gRPC, Data Stream Consumer)
├── Service (Business Logic Layer)
├── Repository (Data Access Layer)
├── Entity (Data Entity)
├── module.go (Module entry point)
└── go.mod (Module dependencies)
```

## Creating a New Module

### 1. Module Structure

Create a new module with the following structure:

```
my-module/
├── module.go              # Module implementation
├── handler/
│   └── handler.go          # HTTP, Messaging Consumer, gRPC, Data Stream Consumer handlers
├── service/
│   └── service.go          # Business logic
├── repository/
│   └── repository.go      # Data access layer
├── entity/
│   └── entity_a.go          # Data Entity
└── go.mod                 # Module dependencies
```

### 2. Module Interface

Every module must implement the `core.Module` interface:

```go
package mymodule

import (
    "github.com/gofiber/fiber/v2"
    "github.com/webcore-go/webcore/app/core"
    "github.com/webcore-go/webcore/app/loader"
    appConfig "github.com/webcore-go/webcore/infra/config"
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

    // Load singleton library via core.LibraryManager.GetSingleton
    // The parameter is taken from the key in APP_LIBRARIES variable in webcore/deps/libraries.go
    if lib, ok := core.Instance().Context.GetDefaultSingletonInstance("database"); ok {
        db := lib.(port.IDatabase)
        
        // Initialize your module components with the new repository pattern
        m.repository = repository.NewRepository(ctx, m.config, nil, db)
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

#### 3.1 HTTP Handler

The handler layer manages HTTP requests and responses:

```go
package handler

import (
    "strconv"
    
    "github.com/gofiber/fiber/v2"
    "github.com/webcore-go/webcore/app/core"
    "github.com/webcore-go/webcore/app/helper"
    "github.com/webcore-go/webcore/app/shared"
    "github.com/webcore-go/webcore/infra/logger"
    "github.com/webcore-go/webcore/modules/mymodule/service"
)

type Handler struct {
    itemService service.ItemService
}

func NewHandler(ctx *core.AppContext, itemService service.ItemService) *Handler {
    return &Handler{
        itemService: itemService,
    }
}

// GetItems returns all items
func (h *Handler) GetItems(c *fiber.Ctx) error {
    page, _ := strconv.Atoi(c.Query("page", "1"))
    pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))
    
    items, total, err := h.itemService.GetItems(c.Context(), page, pageSize)
    if err != nil {
        return h.handleError(c, err)
    }
    
    paginatedItems, pagination := shared.Paginate(items, page, pageSize)
    return c.JSON(shared.NewPaginatedResponse(paginatedItems, pagination))
}

// GetItem returns a single item by ID
func (h *Handler) GetItem(c *fiber.Ctx) error {
    id := c.Params("id")
    
    item, err := h.itemService.GetItem(c.Context(), id)
    if err != nil {
        return h.handleError(c, err)
    }
    
    return c.JSON(shared.NewSuccessResponse(item))
}

// CreateItem creates a new item
func (h *Handler) CreateItem(c *fiber.Ctx) error {
    var item map[string]any
    if err := c.BodyParser(&item); err != nil {
        return h.handleError(c, err)
    }
    
    newItem, err := h.itemService.CreateItem(c.Context(), item)
    if err != nil {
        return h.handleError(c, err)
    }
    
    return c.JSON(shared.NewSuccessResponse(newItem))
}

// UpdateItem updates an existing item
func (h *Handler) UpdateItem(c *fiber.Ctx) error {
    id := c.Params("id")
    
    var item map[string]any
    if err := c.BodyParser(&item); err != nil {
        return h.handleError(c, err)
    }
    
    updatedItem, err := h.itemService.UpdateItem(c.Context(), id, item)
    if err != nil {
        return h.handleError(c, err)
    }
    
    return c.JSON(shared.NewSuccessResponse(updatedItem))
}

// DeleteItem deletes an item
func (h *Handler) DeleteItem(c *fiber.Ctx) error {
    id := c.Params("id")
    
    if err := h.itemService.DeleteItem(c.Context(), id); err != nil {
        return h.handleError(c, err)
    }
    
    return c.JSON(shared.NewSuccessResponse(map[string]any{
        "message": "Item deleted successfully",
    }))
}

// handleError handles errors and returns appropriate HTTP responses
func (h *Handler) handleError(c *fiber.Ctx, err error) error {
    logger.Error("API error", "error", err.Error())
    
    if apiErr, ok := err.(*helper.APIError); ok {
        return c.Status(apiErr.Code).JSON(shared.NewErrorResponse(apiErr.Message))
    }
    
    return c.Status(fiber.StatusInternalServerError).JSON(shared.NewErrorResponse("Internal server error"))
}
```

#### 3.2 Kafka Handler
The handler layer manages Kafka Consumer incomming message
```go
package handler

import (
	"context"
	"sync"

	"github.com/goccy/go-json"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/helper"
	"github.com/webcore-go/webcore/infra/logger"

	"github.com/semanggilab/lib-go-fhir/helper/types"
	"github.com/semanggilab/lib-go-fhir/service"
)

// Config menyimpan semua konfigurasi yang dibutuhkan untuk Kafka Consumer.
type KafkaHandler struct {
	context *core.AppContext
	service *service.FhirTransactionService
}

// NewKafkaHandler membuat dan mengembalikan instance kafka.Reader (consumer) baru.
func NewKafkaHandler(wctx *core.AppContext, service *service.FhirTransactionService) *KafkaHandler {
	return &KafkaHandler{
		context: wctx,
		service: service,
	}
}

func (kc *KafkaHandler) Run(ctx context.Context, data []byte) {
	transaction := types.Transaction{}
	err := json.Unmarshal(data, &transaction)
	if err != nil {
		logger.Info("JSON format invalid:", "error", err)
	} else {
		logger.Debug("Data: " + helper.ToLogJSON(transaction))

		// perbaiki reference dan simpan ke database (synchronous)
		newBundle, register := kc.preprocess(transaction)

		// kemudian teruskan ke HAPI FHIR di thread terpisah (asynchronous)
		if newBundle != nil {
			var wg sync.WaitGroup
			wg.Go(func() {
				kc.handleTransaction(transaction.Method, transaction.Env, transaction.Authorization, newBundle, register)
			})
			wg.Wait()
		}
	}
}

func (kc *KafkaHandler) preprocess(transaction types.Transaction) (any, any) {
	newBundle, set, err := kc.service.Preprocess(transaction)
	if err != nil {
		logger.Info("Gagal saat menrjemahkan data transaksi", "error", err)
		return nil, nil
	}

	return newBundle, set
}
```

#### 3.3 Message Broker (Google Pub/Sub) Consumer
The handler layer manages incomming message from Message Broker (Google Pub/Sub) Consumer

```go
package handler

import (
	"context"
	"fmt"
	"slices"

	"github.com/semanggilab/webcore-go/app/loader"
	"github.com/semanggilab/webcore-go/app/logger"
	ppubsub "github.com/semanggilab/webcore-go/lib/pubsub"
	tbentity "github.com/semanggilab/webcore-go/modules/tb/entity"
	"github.com/semanggilab/webcore-go/modules/tbpubsub/config"
	"github.com/semanggilab/webcore-go/modules/tbpubsub/entity"
	"github.com/semanggilab/webcore-go/modules/tbpubsub/repository"
)

type CkgReceiver struct {
	Configurations *config.ModuleConfig
	CkgRepo        *repository.CKGTBRepository
	PubSubRepo     *repository.PubSubRepository
}

func NewCkgReceiver(ctx context.Context, config *config.ModuleConfig, ckgRepo *repository.CKGTBRepository, pubsubRepo *repository.PubSubRepository) *CkgReceiver {
	return &CkgReceiver{
		Configurations: config,
		PubSubRepo:     pubsubRepo,
		CkgRepo:        ckgRepo,
	}
}

func (r *CkgReceiver) Prepare(ctx context.Context, messages []loader.IPubSubMessage) map[string][]any {
	validMessages := make(map[string][]any)

	// Extract message IDs
	messageIDs := make([]string, 0, len(messages))
	for _, msg := range messages {
		messageIDs = append(messageIDs, msg.GetID())
	}

	// Periksa semua message ID lalu hanya ambil yang belum pernah diproses saja
	existingIDs, err := r.PubSubRepo.GetIncomingIDs(messageIDs)
	if err != nil {
		logger.Debug("Gagal mengambil daftar message ID existing", "error", err)
		existingIDs = []string{}
	}

	// Process semua message satu-satu
	for _, msg := range messages {
		// Skip jika message ID sudah diproses sebelumnya
		if slices.Contains(existingIDs, msg.GetID()) {
			logger.Debug("Skip message", "id", msg.GetID())
			continue
		}

		// Parse message data
		dataStr := string(msg.GetData())

		pubsubObjectWrapper := entity.NewPubSubConsumerWrapper[*tbentity.StatusPasienTBInput](r.Configurations)
		err := pubsubObjectWrapper.FromJSON(dataStr)
		if err != nil {
			logger.Debug("Gagal parsing", "id", msg.GetID(), "error", err)
			continue
		}
		logger.DebugJson("PubSub Receive Object:", pubsubObjectWrapper)

		// Hanya pedulikan Object CKG yang valid
		if !pubsubObjectWrapper.IsCKGObject() {
			logger.Debug("Abaikan message non-CKG", "id", msg.GetID())
			continue
		}

		// Simpan incomming message agar tidak diproses berulang kali
		incoming := entity.IncomingMessageStatusTB{
			ID:   msg.GetID(),
			Data: &dataStr,
			// ReceivedAt:  msg.PublishTime.String(),
			ProcessedAt: nil,
		}
		if err := r.PubSubRepo.SaveNewIncoming(incoming); err != nil {
			logger.Info("Gagal menyimpan incoming message", "id", msg.GetID(), "error", err)
		}

		// register ke validMessages
		validMessages[msg.GetID()] = []any{incoming, msg, pubsubObjectWrapper.Data}
	}

	return validMessages
}

func (r *CkgReceiver) Consume(ctx context.Context, messages []loader.IPubSubMessage) (map[string]bool, error) {
	results := make(map[string]bool)

	// Filter message hanya yang belum diproses saja
	validMessages := r.Prepare(ctx, messages)

	// Process each valid message
	for msgID, data := range validMessages {
		logger.DebugJson("DATA0", data)

		// incoming := data[0].(*entity.IncomingMessageStatusTB)
		msg := data[1].(*ppubsub.PubSubMessage)
		rawStatusPasien := data[2].([]*tbentity.StatusPasienTBInput)
		statusPasien := make([]tbentity.StatusPasienTBInput, 0)
		for _, status := range rawStatusPasien {
			statusPasien = append(statusPasien, *status)
		}

		// Process the message
		err := r.Process(ctx, statusPasien, msg)
		if err != nil {
			logger.Info("Saat memproses message", "id", msgID, "error", err)
			results[msgID] = false
			continue
		}

		results[msgID] = true
	}

	return results, nil
}

func (r *CkgReceiver) Process(ctx context.Context, statusPasien []tbentity.StatusPasienTBInput, msg *ppubsub.PubSubMessage) error {
	logger.Debug(fmt.Sprintf("Received valid CKG SkriningCKG object [%s].\n Data: %s\n Attributes: %v", msg.ID, string(msg.Data), msg.Attributes))
	logger.DebugJson("DATA", statusPasien)
	// Save to database
	_, err := r.CkgRepo.UpdateTbPatientStatus(statusPasien)
	r.PubSubRepo.UpdateIncoming(msg.ID, nil)

	return err
}
```

### 4. Service Layer

The service layer contains business logic:

```go
package service

import (
    "context"
    
    "github.com/webcore-go/webcore/app/core"
    "github.com/webcore-go/webcore/app/helper"
    "github.com/webcore-go/webcore/app/shared"
    "github.com/webcore-go/webcore/infra/logger"
    "github.com/webcore-go/webcore/modules/mymodule/repository"
)

// ItemService defines the interface for item operations
type ItemService interface {
    GetItems(ctx context.Context, page, pageSize int) ([]map[string]any, int, error)
    GetItem(ctx context.Context, id string) (map[string]any, error)
    CreateItem(ctx context.Context, item map[string]any) (map[string]any, error)
    UpdateItem(ctx context.Context, id string, item map[string]any) (map[string]any, error)
    DeleteItem(ctx context.Context, id string) error
}

// Service represents the service layer
type Service struct {
    itemRepository repository.ItemRepository
}

// NewService creates a new Service instance
func NewService(ctx *core.AppContext, itemRepository repository.ItemRepository) *Service {
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
func (s *Service) GetItem(ctx context.Context, id string) (map[string]any, error) {
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
        return nil, helper.WebResponse(&helper.Response{
            Code:    400,
            Message: "Name is required",
        })
    }
    
    // Call repository layer
    newItem, err := s.itemRepository.CreateItem(ctx, item)
    if err != nil {
        return nil, err
    }
    
    logger.Info("Item created", "item_id", newItem["id"])
    return newItem, nil
}

// UpdateItem updates an existing item
func (s *Service) UpdateItem(ctx context.Context, id string, item map[string]any) (map[string]any, error) {
    // Validate input
    if name, ok := item["name"].(string); ok && name == "" {
        return nil, helper.WebResponse(&helper.Response{
            Code:    400,
            Message: "Name cannot be empty",
        })
    }
    
    // Call repository layer
    updatedItem, err := s.itemRepository.UpdateItem(ctx, id, item)
    if err != nil {
        return nil, err
    }
    
    logger.Info("Item updated", "item_id", id)
    return updatedItem, nil
}

// DeleteItem deletes an item
func (s *Service) DeleteItem(ctx context.Context, id string) error {
    // Call repository layer
    err := s.itemRepository.DeleteItem(ctx, id)
    if err != nil {
        return err
    }
    
    logger.Info("Item deleted", "item_id", id)
    return nil
}
```

### 5. Repository Layer

The repository layer handles data access using the WebCoreGo database port interface:

```go
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/helper"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

// ItemRepository defines the interface for item operations
type ItemRepository interface {
	GetItems(ctx context.Context, page, pageSize int) ([]map[string]any, int, error)
	GetItem(ctx context.Context, id string) (map[string]any, error)
	CreateItem(ctx context.Context, item map[string]any) (map[string]any, error)
	UpdateItem(ctx context.Context, id string, item map[string]any) (map[string]any, error)
	DeleteItem(ctx context.Context, id string) error
	CheckDataExists(ctx context.Context, id string) (bool, error)
}

// Repository represents the repository layer
type Repository struct {
	Connection port.IDatabase
	Context    *core.AppContext
	Config     *config.ModuleConfig
	Memory     port.ICacheMemory
}

// NewRepository creates a new Repository instance
func NewRepository(wctx *core.AppContext, config *config.ModuleConfig, mem port.ICacheMemory, conn port.IDatabase) *Repository {
	return &Repository{
		Config:     config,
		Connection: conn,
		Context:    wctx,
		Memory:     mem,
	}
}

// Item represents the item entity
type Item struct {
	ID        string `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	Status    string `json:"status" db:"status"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

// TableName returns the table name for the Item entity
func (Item) TableName() string {
	return "items"
}

// GetPkName returns the primary key name
func (i *Item) GetPkName() string {
	return "id"
}

// GetID returns the ID value
func (i *Item) GetID() string {
	return i.ID
}

// GetItems retrieves items with pagination
func (r *Repository) GetItems(ctx context.Context, page, pageSize int) ([]map[string]any, int, error) {
	var items []Item

	// Build filter for active items
	filter := []port.DbExpression{
		{
			Expr: "deleted_at",
			Args: []any{nil},
		},
	}

	// Get total count
	total, err := r.Connection.Count(r.Context.Context, Item{}.TableName(), filter)
	if err != nil {
		logger.Error("Failed to count items", "error", err)
		return nil, 0, err
	}

	// Get paginated items
	offset := (page - 1) * pageSize
	sort := map[string]int{"created_at": -1}
	err = r.Connection.Find(r.Context.Context, &items, Item{}.TableName(), []string{}, filter, sort, offset, pageSize)
	if err != nil {
		logger.Error("Failed to find items", "error", err)
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
func (r *Repository) GetItem(ctx context.Context, id string) (map[string]any, error) {
	if id == "" {
		return nil, fmt.Errorf("cannot get item with empty ID")
	}

	item := Item{}
	filter := []port.DbExpression{
		{
			Expr: "id",
			Args: []any{id},
		},
		{
			Expr: "deleted_at",
			Args: []any{nil},
		},
	}

	err := r.Connection.FindOne(r.Context.Context, &item, item.TableName(), []string{}, filter, map[string]int{})
	if err != nil {
		logger.Error("Failed to find item", "id", id, "error", err)
		return nil, helper.WebResponse(&helper.Response{
			Code:    404,
			Message: "Item not found",
		})
	}

	return map[string]any{
		"id":         item.ID,
		"name":       item.Name,
		"status":     item.Status,
		"created_at": item.CreatedAt,
		"updated_at": item.UpdatedAt,
	}, nil
}

// CheckDataExists checks if an item exists by ID
func (r *Repository) CheckDataExists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, fmt.Errorf("cannot check data with empty ID for table %s", Item{}.TableName())
	}

	filter := []port.DbExpression{
		{
			Expr: "id",
			Args: []any{id},
		},
		{
			Expr: "deleted_at",
			Args: []any{nil},
		},
	}

	count, err := r.Connection.Count(r.Context.Context, Item{}.TableName(), filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateItem creates a new item
func (r *Repository) CreateItem(ctx context.Context, item map[string]any) (map[string]any, error) {
	// Convert map to database map
	dbmap, err := helper.MarshalDbMap(item)
	if err != nil {
		return nil, err
	}

	// Set default status if not provided
	if _, ok := dbmap["status"]; !ok {
		dbmap["status"] = "active"
	}

	// Insert into database
	_, err = r.Connection.InsertOne(r.Context.Context, Item{}.TableName(), dbmap)
	if err != nil {
		logger.Error("Failed to create item", "error", err)
		return nil, helper.WebResponse(&helper.Response{
			Code:    400,
			Message: "Failed to create item",
			Details: err.Error(),
		})
	}

	// Return the created item
	return item, nil
}

// UpdateItem updates an existing item
func (r *Repository) UpdateItem(ctx context.Context, id string, item map[string]any) (map[string]any, error) {
	if id == "" {
		return nil, fmt.Errorf("cannot update item with empty ID")
	}

	// Check if item exists
	exists, err := r.CheckDataExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, helper.WebResponse(&helper.Response{
			Code:    404,
			Message: "Item not found",
		})
	}

	// Convert map to database map
	dbmap, err := helper.MarshalDbMap(item)
	if err != nil {
		return nil, err
	}

	// Build filter for update
	filter := []port.DbExpression{
		{
			Expr: "id",
			Args: []any{id},
		},
	}

	// Update in database
	_, err = r.Connection.UpdateOne(r.Context.Context, Item{}.TableName(), filter, dbmap)
	if err != nil {
		logger.Error("Failed to update item", "id", id, "error", err)
		return nil, helper.WebResponse(&helper.Response{
			Code:    400,
			Message: "Failed to update item",
			Details: err.Error(),
		})
	}

	// Return the updated item
	return item, nil
}

// DeleteItem deletes an item (soft delete)
func (r *Repository) DeleteItem(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("cannot delete item with empty ID")
	}

	// Check if item exists
	exists, err := r.CheckDataExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return helper.WebResponse(&helper.Response{
			Code:    404,
			Message: "Item not found",
		})
	}

	// Build filter for delete
	filter := []port.DbExpression{
		{
			Expr: "id",
			Args: []any{id},
		},
	}

	// Soft delete by setting deleted_at
	deletedAt := map[string]any{
		"deleted_at": time.Now(),
	}

	_, err = r.Connection.UpdateOne(r.Context.Context, Item{}.TableName(), filter, deletedAt)
	if err != nil {
		logger.Error("Failed to delete item", "id", id, "error", err)
		return helper.WebResponse(&helper.Response{
			Code:    400,
			Message: "Failed to delete item",
			Details: err.Error(),
		})
	}

	return nil
}
```

#### Key Repository Patterns

The repository layer follows these patterns from the FHIR repository:

1. **Database Port Interface**: Uses `port.IDatabase` for database operations instead of direct ORM dependencies
2. **DbExpression for Filtering**: Uses `port.DbExpression` to build query filters
3. **Context Management**: Uses `core.AppContext` for application context
4. **Cache Memory**: Supports optional caching via `port.ICacheMemory`
5. **Helper Functions**: Uses `helper.MarshalDbMap` to convert entities to database maps
6. **Soft Delete**: Implements soft delete pattern using `deleted_at` field
7. **Error Handling**: Uses structured error responses with `helper.WebResponse`

### 6. Module Registration

#### Registering Module in webcore/deps/packages.go

To ensure your module is properly loaded and its dependencies are triggered safely, register your module in `webcore/deps/packages.go`:

```go
// webcore/deps/packages.go
package deps

import (
    "github.com/webcore-go/webcore/app/core"
    mymodule "github.com/webcore-go/webcore/modules/mymodule"
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
    // Load singleton library via core.LibraryManager.GetSingleton
    // The parameter is taken from the key in APP_LIBRARIES variable in webcore/deps/libraries.go
    if lib, ok := core.Instance().Context.GetDefaultSingletonInstance("database"); ok {
        // shared library successfully loaded
        db := lib.(port.IDatabase)

        // Initialize your module components with the new repository pattern
        m.repository = repository.NewRepository(ctx, m.config, nil, db)
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
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/lib/mongo"
)

var APP_LIBRARIES = map[string]core.LibraryLoader{
	"database:mongodb": &mongo.MongoLoader{},

	// Add your library here
}

```

### Using Shared Modules

Your module can use shared dependencies (config, handler, repository, service, etc.) directly from other modules using standar import. You must ensure modules is registered in `webcore/deps/packages.go` or import as package from golang repository. Here is an example:

```go
// in modules/moduleb/config.go
package config

import (
	configa "github.com/webcore-go/webcore/modulesa/tb/config"
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
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestItemRepository_GetItems(t *testing.T) {
    // Setup test database
    conn := setupTestDatabase(t)
    wctx := setupTestContext()
    repo := NewRepository(wctx, &config.ModuleConfig{}, nil, conn)
    
    // Test data
    items := []Item{
        {ID: "1", Name: "Item 1", Status: "active"},
        {ID: "2", Name: "Item 2", Status: "active"},
    }
    
    // Insert test data
    for _, item := range items {
        dbmap, _ := helper.MarshalDbMap(item)
        conn.InsertOne(context.Background(), Item{}.TableName(), dbmap)
    }
    
    // Test
    result, total, err := repo.GetItems(context.Background(), 1, 10)
    
    // Assertions
    require.NoError(t, err)
    assert.Equal(t, len(items), total)
    assert.Len(t, result, len(items))
}

func TestItemRepository_CheckDataExists(t *testing.T) {
    // Setup test database
    conn := setupTestDatabase(t)
    wctx := setupTestContext()
    repo := NewRepository(wctx, &config.ModuleConfig{}, nil, conn)
    
    // Test data
    item := Item{ID: "1", Name: "Item 1", Status: "active"}
    dbmap, _ := helper.MarshalDbMap(item)
    conn.InsertOne(context.Background(), Item{}.TableName(), dbmap)
    
    // Test existing item
    exists, err := repo.CheckDataExists(context.Background(), "1")
    require.NoError(t, err)
    assert.True(t, exists)
    
    // Test non-existing item
    exists, err = repo.CheckDataExists(context.Background(), "999")
    require.NoError(t, err)
    assert.False(t, exists)
}

// service_test.go
package service

import (
    "context"
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

func TestItemService_GetItem(t *testing.T) {
    // Setup
    deps := setupTestContext()
    repo := &mockRepository{
        data: map[string]map[string]any{
            "1": {"id": "1", "name": "Test Item", "status": "active"},
        },
    }
    service := NewService(deps, repo)
    
    // Test
    result, err := service.GetItem(context.Background(), "1")
    
    // Assertions
    require.NoError(t, err)
    assert.Equal(t, "1", result["id"])
    assert.Equal(t, "Test Item", result["name"])
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
    "github.com/stretchr/testify/require"
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

func TestHandler_GetItem(t *testing.T) {
    // Setup
    app := fiber.New()
    handler := setupTestHandler()
    
    // Register routes
    app.Get("/api/v1/mymodule/items/:id", handler.GetItem)
    
    // Test
    req := httptest.NewRequest(http.MethodGet, "/api/v1/mymodule/items/1", nil)
    resp, err := app.Test(req)
    
    // Assertions
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_CreateItem(t *testing.T) {
    // Setup
    app := fiber.New()
    handler := setupTestHandler()
    
    // Register routes
    app.Post("/api/v1/mymodule/items", handler.CreateItem)
    
    // Test data
    body := strings.NewReader(`{"name": "Test Item"}`)
    req := httptest.NewRequest(http.MethodPost, "/api/v1/mymodule/items", body)
    req.Header.Set("Content-Type", "application/json")
    resp, err := app.Test(req)
    
    // Assertions
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_UpdateItem(t *testing.T) {
    // Setup
    app := fiber.New()
    handler := setupTestHandler()
    
    // Register routes
    app.Put("/api/v1/mymodule/items/:id", handler.UpdateItem)
    
    // Test data
    body := strings.NewReader(`{"name": "Updated Item"}`)
    req := httptest.NewRequest(http.MethodPut, "/api/v1/mymodule/items/1", body)
    req.Header.Set("Content-Type", "application/json")
    resp, err := app.Test(req)
    
    // Assertions
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_DeleteItem(t *testing.T) {
    // Setup
    app := fiber.New()
    handler := setupTestHandler()
    
    // Register routes
    app.Delete("/api/v1/mymodule/items/:id", handler.DeleteItem)
    
    // Test
    req := httptest.NewRequest(http.MethodDelete, "/api/v1/mymodule/items/1", nil)
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
// config.go
package config

import (
    "github.com/webcore-go/webcore/infra/config"
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

Enable debug mode for detailed logging edit `config.yaml`:

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
