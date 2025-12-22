package core

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/semanggilab/webcore-go/app/config"
	"github.com/semanggilab/webcore-go/app/logger"
)

// Context represents shared dependencies that can be injected into modules
type AppContext struct {
	Context  context.Context
	Config   *config.Config
	Web      *fiber.App
	Root     fiber.Router
	EventBus *EventBus
	// Database map[string]db.Database
	// Redis    *redis.Redis
	// PubSub   map[string]*pubsub.PubSub
}

func (a *AppContext) Start() error {
	libmanager := Instance().LibraryManager

	// Initialize shared dependencies
	// a.Context.Database["default"] = nil
	// a.Context.PubSub["default"] = nil

	// Initialize database if configured
	if a.Config.Database.Host != "" {
		// a.SetupDatabase("default", a.Config.Database)
		lName := "db:" + a.Config.Database.Driver
		loader, ok := libmanager.GetLoader(lName)
		if !ok {
			return fmt.Errorf("LibraryLoader '%s' tidak ditemukan", lName)
		}

		// Nama dari struct yang implement interface Library
		// lTypeName := "SQLDatabase"
		// if a.Config.Database.Driver == "mongodb" {
		// 	lTypeName = "MongoDatabase"
		// }

		_, err := libmanager.LoadSingletonFromLoader(loader, a.Context, a.Config.Database)
		if err != nil {
			return err
		}
	}

	// Initialize Redis if configured
	if a.Config.Redis.Host != "" {
		// a.SetupRedis(a.Config.Redis)
		loader, ok := libmanager.GetLoader("redis")
		if !ok {
			return fmt.Errorf("LibraryLoader 'redis' tidak ditemukan")
		}
		_, err := libmanager.LoadSingletonFromLoader(loader, a.Context, a.Config.Database)
		if err != nil {
			return err
		}
	}

	// Initialize PubSub if configured
	if a.Config.PubSub.ProjectID != "" && a.Config.PubSub.Topic != "" {
		// a.SetupPubSub("default", a.Config.PubSub)
		loader, ok := libmanager.GetLoader("pubsub")
		if !ok {
			return fmt.Errorf("LibraryLoader 'pubsub' tidak ditemukan")
		}
		_, err := libmanager.LoadSingletonFromLoader(loader, a.Context, a.Config.Database)
		if err != nil {
			return err
		}
	}

	return nil
}

// func (a *AppContext) SetupDatabase(name string, config config.DatabaseConfig) (db.Database, error) {
// 	db, err := db.GetConnection(&config)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to initialize database: %v", err)
// 	}

// 	d.Database[name] = db
// 	return db, nil
// }

// func (a *AppContext) SetupRedis(config config.RedisConfig) (*redis.Redis, error) {
// 	redisClient, err := redis.NewRedis(config)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to initialize Redis: %v", err)
// 	}
// 	d.Redis = redisClient
// 	return redisClient, nil
// }

// // SetupPubSub initializes PubSub connection
// func (a *AppContext) SetupPubSub(name string, config config.PubSubConfig) error {
// 	if config.ProjectID == "" || config.Topic == "" || config.Subscription == "" {
// 		d.PubSub = nil
// 		return nil
// 	}

// 	pubSub, err := pubsub.NewPubSub(
// 		d.Context,
// 		config,
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to initialize PubSub: %v", err)
// 	}

// 	d.PubSub[name] = pubSub
// 	return nil
// }

// Destroy release all resources
func (a *AppContext) Destroy() error {
	// Shutdown Fiber app
	if a.Web != nil {
		return a.Web.Shutdown()
	}

	return nil

	// var errors []error
	// // Close database connections
	// for name, db := range d.Database {
	// 	if db != nil {
	// 		if err := db.Close(d.Context); err != nil {
	// 			errors = append(errors, fmt.Errorf("failed to close database %s: %v", name, err))
	// 		}
	// 	}
	// }

	// // Close Redis connection
	// if d.Redis != nil {
	// 	if err := d.Redis.Close(); err != nil {
	// 		errors = append(errors, fmt.Errorf("failed to close Redis: %v", err))
	// 	}
	// }

	// // Close PubSub connection
	// for name, pb := range d.PubSub {
	// 	if pb != nil {
	// 		if err := pb.Close(); err != nil {
	// 			errors = append(errors, fmt.Errorf("failed to close PubSub %s: %v", name, err))
	// 		}
	// 	}
	// }

	// if len(errors) > 0 {
	// 	return fmt.Errorf("encountered %d errors while closing dependencies: %v", len(errors), errors)
	// }

	// // Close database connections
	// if a.Database != nil {
	// 	for name, db := range a.Database {
	// 		if db == nil {
	// 			continue
	// 		}

	// 		if err := db.Close(a.Context); err != nil {
	// 			log.Printf("Error closing database connection %s: %v", name, err)
	// 		}
	// 	}
	// }

	// if a.Redis != nil {
	// 	if err := a.Redis.Close(); err != nil {
	// 		log.Printf("Error closing Redis connection: %v", err)
	// 	}
	// }

	// // Close PubSub connection
	// if a.PubSub != nil {
	// 	for name, pb := range a.PubSub {
	// 		if pb == nil {
	// 			continue
	// 		}

	// 		if err := pb.Close(); err != nil {
	// 			log.Printf("Error closing PubSub connection %s: %v", name, err)
	// 		}
	// 	}
	// }
}

func CheckSingleLoader[L any](name string, loaders []L) []L {
	newLoaders := []L{}
	list := []string{}
	for _, loader := range loaders {
		lType := reflect.TypeOf(loader)
		if lType.Kind() == reflect.Ptr {
			lType = lType.Elem()
		}

		lName := lType.Name()
		if slices.Contains(list, lName) {
			logger.Fatal(name+" is registered multiple times", "name", lName)
		}

		newLoaders = append(newLoaders, loader)
	}

	return newLoaders
}

func AppendRouteToArray(routes []*ModuleRoute, route *ModuleRoute) []*ModuleRoute {
	route.Root.Add(route.Method, route.Path, route.Handler)

	routes = append(routes, route)
	return routes
}
