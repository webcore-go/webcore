# Development Library Documentation

This document provides comprehensive guidance on developing libraries for the WebCore framework.

## Overview

Libraries in WebCore are shared components that provide functionality across the application. They follow a standardized interface pattern and are managed through a central library manager.

## Library Structure

Each library consists of two main components:
1. **Library Implementation** - The actual functionality (implements `core.Library`)
2. **Library Loader** - A proxy for instantiating the library (implements `core.LibraryLoader`)

## Step 1: Create Library Directory Structure

Create a new directory under `libraries/` for your library:

```
libraries/your-library/
├── go.mod
├── go.sum
├── loader.go         # Library loader implementation
└── your_library.go   # Library implementation
```

## Step 2: Update go.work

Add your library to the `go.work` file:

```go
use (
    // ... existing libraries
    ./libraries/your-library
    // ... other libraries
)
```

## Step 3: Implement the Library Interface

Create your library implementation that implements the `core.Library` interface:

```go
package yourlibrary

import (
    "github.com/semanggilab/webcore-go/app/loader"
)

// YourLibrary represents your shared library
type YourLibrary struct {
    // Your fields here
}

// NewYourLibrary creates a new YourLibrary instance
func NewYourLibrary(config YourConfig) *YourLibrary {
    return &YourLibrary{
        // Initialize your library
    }
}

// Install is called when the library is loaded
func (l *YourLibrary) Install(args ...any) error {
    // Installation logic
    return nil
}

// Connect establishes the connection/initialization
func (l *YourLibrary) Connect() error {
    // Connection logic
    return nil
}

// Close cleans up resources
func (l *YourLibrary) Close() error {
    // Cleanup logic
    return nil
}

// Uninstall is called when the library is unloaded
func (l *YourLibrary) Uninstall() error {
    // Uninstallation logic
    return nil
}
```

## Step 4: Implement the Library Loader

Create a loader that acts as a proxy for instantiating your library:

```go
package yourlibrary

import (
    "github.com/semanggilab/webcore-go/app/config"
    "github.com/semanggilab/webcore-go/app/loader"
)

type YourLibraryLoader struct {
    YourLibrary *YourLibrary
}

func (l *YourLibraryLoader) ClassName() string {
    return "YourLibrary"
}

func (l *YourLibraryLoader) Init(args ...any) (loader.Library, error) {
    config := args[0].(YourConfig)
    library := NewYourLibrary(config)
    err := library.Install(args...)
    if err != nil {
        return nil, err
    }

    library.Connect()

    l.YourLibrary = library
    return library, nil
}
```

## Step 5: Register Library in webcore/deps/libraries.go

Add your library loader to the `ALL_LIBRARIES` map:

```go
package deps

import (
    "github.com/semanggilab/webcore-go/app/core"
    "github.com/semanggilab/webcore-go/lib/yourlibrary"
)

var APP_LIBRARIES = map[string]core.LibraryLoader{
    // ... existing libraries
    "yourlibrary": &yourlibrary.YourLibraryLoader{},
    
    // Add your library here
}
```

## Step 6: Initialize Singleton in your module (modules/mymodule/module.go)

Add initialization logic in the `Start()` method:

```go
// In module.go
func (m *Module) Init(ctx *core.AppContext) error {
    // Get the library manager from the core instance
    libmanager := core.Instance().LibraryManager

    // Initialize your library if configured
    if m.Config.YourLibrary.Enabled {
        loader, ok := libmanager.GetLoader("yourlibrary")
        if !ok {
            return fmt.Errorf("LibraryLoader 'yourlibrary' tidak ditemukan")
        }

        // arg[0] context.Context, arg[1] config.YourLibrary
        _, err := libmanager.LoadSingletonFromLoader(loader, ctx.Context, m.Config.YourLibrary)
        if err != nil {
            return err
        }
    }

    // ... other initialization
    return nil
}
```

When calling `libmanager.LoadSingletonFromLoader(loader, args ...anyl)` after loader argument you can define your own arguments as needed by library. This argument arrangement will passed into `YourLibraryLoader.Init()` and `YourLibrary.Install()`.

Here example in `yourlibrary/loader.go`
```go
// In yourlibrary/loader.go
func (l *YourLibraryLoader) Init(args ...any) (loader.Library, error) {
    // arg[0] context.Context, arg[1] config.YourLibrary
    context := args[0].(context.Context)
    config := args[1].(config.YourLibrary)

    library := NewYourLibrary(config)
    err := library.Install(args...)
    if err != nil {
        return nil, err
    }

    // ... other initialization

    return library, nil
}
```

Here example in `yourlibrary/your_library.go`
```go
// In yourlibrary/your_library.go

// Install is called when the library is loaded
func (l *YourLibrary) Install(args ...any) error {
    // arg[0] context.Context, arg[1] config.YourLibrary
    context := args[0].(context.Context)
    config := args[1].(config.YourLibrary)

    // ... Installation logic
    
    return nil
}

```

## Step 7: Use Library in Modules

Access your library from within modules:

```go
// In module.go
func (m *Module) Init(ctx *core.AppContext) error {
    libmanager := core.Instance().LibraryManager

    // Initialize your library if configured
    if a.Config.YourLibrary.Enabled {
        // ... other initialization

        // Get your library instance
        if lib, ok := libmanager.GetSingleton("YourLibrary"); ok {
            yourLib := lib.(loader.YourLibraryInterface) // Cast to your interface
            
            // Use your library
            m.service = service.NewService(ctx, yourLib)
        }
    }

    // ... rest of initialization
    return nil
}
```

## Example: Complete Redis Library Implementation

### libraries/redis/redis.go
```go
package redis

import (
    "fmt"

    "github.com/go-redis/redis/v8"
    "github.com/semanggilab/webcore-go/app/config"
)

// Redis represents shared Redis connection
type Redis struct {
    Client *redis.Client
}

// NewRedis creates a new Redis connection
func NewRedis(config config.RedisConfig) *Redis {
    client := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
        Password: config.Password,
        DB:       config.DB,
    })

    return &Redis{Client: client}
}

func (r *Redis) Install(args ...any) error {
    return nil
}

func (r *Redis) Connect() error {
    // Test connection
    _, err := r.Client.Ping(r.Client.Context()).Result()
    if err != nil {
        return fmt.Errorf("failed to connect to Redis: %v", err)
    }
    return nil
}

func (r *Redis) Close() error {
    return r.Client.Close()
}

func (r *Redis) Uninstall() error {
    return nil
}
```

### libraries/redis/loader.go
```go
package redis

import (
    "github.com/semanggilab/webcore-go/app/config"
    "github.com/semanggilab/webcore-go/app/loader"
)

type RedisLoader struct {
    Redis *Redis
}

func (l *RedisLoader) ClassName() string {
    return "Redis"
}

func (l *RedisLoader) Init(args ...any) (loader.Library, error) {
    config := args[0].(config.RedisConfig)
    redis := NewRedis(config)
    err := redis.Install(args...)
    if err != nil {
        return nil, err
    }

    redis.Connect()

    l.Redis = redis
    return redis, nil
}
```

## Configuration

Add your library configuration to the main config:

```yaml
your-library:
  enabled: true
  host: localhost
  port: 5432
  username: user
  password: pass
  database: yourdb
```

## Best Practices

1. **Interface Segregation**: Define clear interfaces for your library
2. **Error Handling**: Return meaningful error messages
3. **Connection Management**: Implement proper connection lifecycle
4. **Configuration**: Use the framework's configuration system
5. **Logging**: Use the framework's logger for consistent logging
6. **Testing**: Write unit and integration tests for your library

## Common Patterns

### Singleton Pattern
Most libraries should be singletons to ensure shared state and efficient resource usage.

### Configuration Pattern
Use the framework's configuration system to make libraries configurable.

### Interface Pattern
Define clear interfaces for your library to improve testability and decoupling.

### Lifecycle Management
Implement proper `Install`, `Connect`, `Close`, and `Uninstall` methods.

## Troubleshooting

### Library Not Found
- Check if your library is registered in `webcore/deps/libraries.go`
- Verify the loader name matches what you're using in `core.go`

### Configuration Issues
- Ensure your configuration structure matches the config file
- Check environment variable bindings

### Connection Issues
- Verify connection parameters in configuration
- Check network connectivity
- Implement proper error handling in `Connect()` method

## Integration with Existing Libraries

Your library can integrate with other libraries by:

1. **Dependency Injection**: Request other libraries during initialization
2. **Event System**: Use the event bus for communication
3. **Service Registry**: Register services that other modules can use

## Performance Considerations

1. **Lazy Loading**: Only initialize when needed
2. **Connection Pooling**: Use connection pools for database-like libraries
3. **Caching**: Implement caching where appropriate
4. **Resource Management**: Properly close connections and free resources