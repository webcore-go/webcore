package core

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/port"
)

// Context represents shared dependencies that can be injected into modules
type AppContext struct {
	Context  context.Context
	Config   *config.Config
	Web      *fiber.App
	Root     fiber.Router
	EventBus *EventBus
}

func (a *AppContext) Start() error {
	libmanager := Instance().LibraryManager

	// Initialize database if configured
	if a.Config.Database.Host != "" {
		// lName := "database:" + a.Config.Database.Driver
		// loader, ok := libmanager.GetLoader(lName)
		loader, e := a.GetDefaultLibraryLoader("database")
		if e != nil {
			return e
		}

		_, err := libmanager.LoadSingletonFromLoader(loader, a.Context, a.Config.Database)
		if err != nil {
			return err
		}
	}

	// Initialize database if configured
	if a.Config.Memory.Enabled {
		loader, ok := libmanager.GetLoader("memory")
		if !ok {
			loader, ok = libmanager.GetLoader("cache:memory")
			if !ok {
				return fmt.Errorf("LibraryLoader 'cache:memory' tidak ditemukan")
			}
		}
		_, err := libmanager.LoadSingletonFromLoader(loader, a.Config.Memory)
		if err != nil {
			return err
		}
	}

	// Initialize Redis if configured
	if a.Config.Redis.Host != "" {
		// a.SetupRedis(a.Config.Redis)
		loader, ok := libmanager.GetLoader("redis")
		if !ok {
			loader, ok = libmanager.GetLoader("cache:redis")
			if !ok {
				return fmt.Errorf("LibraryLoader 'redis' tidak ditemukan")
			}
		}
		_, err := libmanager.LoadSingletonFromLoader(loader, a.Config.Redis)
		if err != nil {
			return err
		}
	}

	// Initialize Kafka if configured
	if len(a.Config.Kafka.Brokers) > 0 {
		// a.SetupKafka("default", a.Config.Kafka)
		loaderProducer, okProducer := libmanager.GetLoader("kafka:producer")
		if okProducer {
			_, err := libmanager.LoadSingletonFromLoader(loaderProducer, a.Config.Kafka)
			if err != nil {
				return err
			}
		}

		// Kafka Consumer tidak bisa otomatis di-load di sini karena membutuhkan
		// handler khusus yang didefinisikan di modul masing-masing.
		_, okConsumer := libmanager.GetLoader("kafka:consumer")

		if !okProducer && !okConsumer {
			return fmt.Errorf("LibraryLoader 'kafka' tidak ditemukan")
		}
	}

	// Initialize PubSub if configured
	if a.Config.PubSub.ProjectID != "" && a.Config.PubSub.Topic != "" {
		// a.SetupPubSub("default", a.Config.PubSub)
		loader, ok := libmanager.GetLoader("pubsub")
		if !ok {
			return fmt.Errorf("LibraryLoader 'pubsub' tidak ditemukan")
		}
		_, err := libmanager.LoadSingletonFromLoader(loader, a.Context, a.Config.PubSub)
		if err != nil {
			return err
		}
	}

	return nil
}

// Destroy release all resources
func (a *AppContext) Destroy() error {
	// Shutdown Fiber app
	if a.Web != nil {
		return a.Web.Shutdown()
	}

	return nil
}

func (a *AppContext) GetLibraryLoader(name string) (LibraryLoader, error) {
	loader, ok := Instance().LibraryManager.GetLoader(name)
	if !ok {
		return nil, fmt.Errorf("LibraryLoader '%s' tidak ditemukan", name)
	}

	return loader, nil
}

func (a *AppContext) GetDefaultLibraryLoader(name string) (LibraryLoader, error) {
	return a.GetLibraryLoader(a.getDefaultName(name))
}

func (a *AppContext) LoadSingletonInstance(loader LibraryLoader, args ...any) (port.Library, error) {
	return Instance().LibraryManager.LoadSingletonFromLoader(loader, args...)
}

func (a *AppContext) LoadInstance(loader LibraryLoader, key string, args ...any) (port.Library, error) {
	return Instance().LibraryManager.LoadInstanceFromLoader(loader, key, args...)
}

func (a *AppContext) GetSingletonInstance(name string) (port.Library, bool) {
	return Instance().LibraryManager.GetSingletonInstance(name)
}

func (a *AppContext) GetDefaultSingletonInstance(name string) (port.Library, bool) {
	return a.GetSingletonInstance(a.getDefaultName(name))
}

func (a *AppContext) GetInstance(name string, key string) (port.Library, bool) {
	return Instance().LibraryManager.GetInstance(name, key)
}

func (a *AppContext) GetDefaultInstance(name string, key string) (port.Library, bool) {
	return a.GetInstance(a.getDefaultName(name), key)
}

func (a *AppContext) getDefaultName(name string) string {
	switch name {
	case "database":
		name = name + ":" + a.Config.Database.Driver
	case "authstorage":
		name = name + ":" + a.Config.Auth.Store
	case "authentication":
		name = name + ":" + a.Config.Auth.Type
	}
	return name
}

func AppendRouteToArray(routes []*ModuleRoute, route *ModuleRoute) []*ModuleRoute {
	route.Root.Add(route.Method, route.Path, route.Handler)

	routes = append(routes, route)
	return routes
}
