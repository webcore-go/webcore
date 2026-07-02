package core

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

// Context represents shared dependencies that can be injected into modules
type AppContext struct {
	Context     context.Context
	Config      *config.Config
	Web         *fiber.App
	Root        fiber.Router
	AuthHandler fiber.Handler
	EventBus    *EventBus
	Hook        *Hook
}

func (a *AppContext) Start() error {
	libmanager := Instance().LibraryManager

	if a.Config.App.Logging.Remote.Uri != "" {
		loader, e := a.GetDefaultLibraryLoader("remotelog")
		if e != nil {
			return e
		}

		if loader != nil {
			libLog, err := libmanager.LoadSingletonFromLoader(loader, a.Context, a.Config.App.Logging.Remote, a.Config.App.Environment)
			if err != nil {
				return err
			}

			// segera regiseter remote log handler
			remoteLog := libLog.(port.IRemoteLog)
			remoteLog.SetMinimumLevelCapture(slog.LevelError)
			if len(a.Config.App.Logging.Remote.DefaultTags) > 0 {
				remoteLog.SetDefaultTags(a.Config.App.Logging.Remote.DefaultTags)
			}
			if len(a.Config.App.Logging.Remote.DefaultContexts) > 0 {
				remoteLog.SetDefaultContexts(a.Config.App.Logging.Remote.DefaultContexts)
			}
			logger.SetRemote(remoteLog)
			a.Web.Use(remoteLog.NewHandler())
		}
	}

	// Initialize database if configured
	if a.Config.Database.Host != "" || a.Config.Database.Uri != "" {
		/*loader, e := a.GetDefaultLibraryLoader("database")
		if e != nil {
			return e
		}

		if loader != nil {
			_, err := libmanager.LoadSingletonFromLoader(loader, a.Context, a.Config.Database)
			if err != nil {
				return err
			}
		}*/
		_, err := a.StartDefaultSingletonInstance("database", a.Context, a.Config.Database)
		if err != nil {
			return err
		}

		logger.Info("Library Database loaded", "driver", a.Config.Database.Driver)
	}

	// Initialize database if configured
	if a.Config.Memory.Enabled {
		name := "memory"
		loader, ok := libmanager.GetLoader(name)
		if !ok {
			name = "cache:memory"
			loader, _ = libmanager.GetLoader(name) // tidak perlu error kalau library tidak ditemukan
		}

		if loader != nil {
			_, err := libmanager.LoadSingletonFromLoader(loader, a.Config.Memory)
			if err != nil {
				return err
			}

			logger.Info("Library Cache", "loaded", name)
		}
	}

	// Initialize Redis if configured
	if a.Config.Redis.Host != "" {
		name := "redis"
		// a.SetupRedis(a.Config.Redis)
		loader, ok := libmanager.GetLoader(name)
		if !ok {
			name = "cache:redis"
			loader, _ = libmanager.GetLoader(name) // tidak perlu error kalau library tidak ditemukan
		}

		if loader != nil {
			_, err := libmanager.LoadSingletonFromLoader(loader, a.Config.Redis)
			if err != nil {
				return err
			}

			logger.Info("Library Cache", "loaded", name, "host", a.Config.Redis.Host)
		}
	}

	// Initialize Kafka if configured
	if a.Config.Kafka.Enabled && len(a.Config.Kafka.Brokers) > 0 {
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
		libmanager.GetLoader("kafka:consumer") // tidak perlu error kalau library tidak ditemukan

		// _, okConsumer := libmanager.GetLoader("kafka:consumer")

		// if !okProducer && !okConsumer {
		// 	return fmt.Errorf("LibraryLoader 'kafka' tidak ditemukan")
		// }

		logger.Info("Library Kafka loaded", "brokers", a.Config.Kafka.Brokers)
	}

	// Initialize PubSub if configured
	if a.Config.PubSub.Driver != "" {
		if a.Config.PubSub.Driver != "gpubsub" {
			_, ok := a.Config.GetOtherItem(a.Config.PubSub.Driver)
			if ok {
				loader, _ := libmanager.GetLoader("pubsub") // tidak perlu error kalau library tidak ditemukan
				if loader != nil {
					// _, err := libmanager.LoadSingletonFromLoader(loader, a.Context, a.Config.PubSub)
					// if err != nil {
					// 	return err
					// }

					logger.Info("Library PubSub loaded", "Driver", a.Config.PubSub.Driver)
				}
			}
		} else if a.Config.PubSub.ProjectID != "" && a.Config.PubSub.Topic != "" {
			// a.SetupPubSub("default", a.Config.PubSub)
			loader, _ := libmanager.GetLoader("pubsub") // tidak perlu error kalau library tidak ditemukan
			if loader != nil {
				// _, err := libmanager.LoadSingletonFromLoader(loader, a.Context, a.Config.PubSub)
				// if err != nil {
				// 	return err
				// }

				logger.Info("Library PubSub loaded", "Driver", a.Config.PubSub.Driver, "Project", a.Config.PubSub.ProjectID, "Topic", a.Config.PubSub.Topic)
			}
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

func (a *AppContext) AddHook(pos string, fn HookFunc) {
	a.Hook.AddFunc(pos, fn)
}

func (a *AppContext) RunHook(pos string) {
	a.Hook.RunFunc(pos)
}

func (a *AppContext) StartSingletonInstance(name string, args ...any) (port.Library, error) {
	loader, e := a.GetLibraryLoader(name)
	if e != nil {
		return nil, e
	}

	return a.LoadSingletonInstance(loader, args...)
}

func (a *AppContext) StartDefaultSingletonInstance(name string, args ...any) (port.Library, error) {
	loader, e := a.GetDefaultLibraryLoader(name)
	if e != nil {
		return nil, e
	}

	return a.LoadSingletonInstance(loader, args...)
}

func (a *AppContext) StartInstance(name string, key string, args ...any) (port.Library, error) {
	loader, e := a.GetLibraryLoader(name)
	if e != nil {
		return nil, e
	}

	return a.LoadInstance(loader, key, args...)
}

func (a *AppContext) StartDefaultInstance(name string, key string, args ...any) (port.Library, error) {
	loader, e := a.GetDefaultLibraryLoader(name)
	if e != nil {
		return nil, e
	}

	return a.LoadInstance(loader, key, args...)
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
	case "authsession":
		name = name + ":" + a.Config.Auth.Session.Backend
	case "authentication":
		name = name + ":" + a.Config.Auth.Type
	}
	return name
}

// AppendRouteToArray collects a route without registering it immediately.
// Registration is deferred so that, once every route is known, they can be
// registered with the router in priority order via RegisterRoutes. This keeps a
// parameterised route such as "/posyandu/:id" from shadowing a static route such
// as "/posyandu/batch", because Fiber matches routes strictly in registration
// order and a route cannot be reordered once it has been added.
func AppendRouteToArray(routes []*ModuleRoute, route *ModuleRoute) []*ModuleRoute {
	return append(routes, route)
}

// RegisterRoutes registers all collected routes onto their routers. Routes are
// grouped per router and sorted by path priority (static segments before
// parameterised ones) so that the most specific route is always matched first.
func RegisterRoutes(routes []*ModuleRoute) {
	groups := make(map[fiber.Router][]*ModuleRoute)
	order := make([]fiber.Router, 0)
	for _, route := range routes {
		if route == nil || route.Root == nil {
			continue
		}
		if _, exists := groups[route.Root]; !exists {
			order = append(order, route.Root)
		}
		groups[route.Root] = append(groups[route.Root], route)
	}

	for _, root := range order {
		group := groups[root]
		sort.SliceStable(group, func(i, j int) bool {
			return routeHasHigherPriority(group[i], group[j])
		})
		for _, route := range group {
			registerRoute(route)
		}
	}
}

func registerRoute(route *ModuleRoute) {
	handlers := route.Handlers
	if len(handlers) == 0 && route.Handler != nil {
		handlers = []fiber.Handler{route.Handler}
	}
	route.Root.Add(route.Method, route.Path, handlers...)
}

// routeHasHigherPriority reports whether route i must be registered before
// route j so that it is not shadowed by a less specific route.
func routeHasHigherPriority(i, j *ModuleRoute) bool {
	// Different methods never conflict (the router keeps per-method stacks), but
	// grouping them keeps the resulting order deterministic.
	if i.Method != j.Method {
		return i.Method < j.Method
	}
	// Same method: the more specific path wins and must be registered first.
	if cmp := comparePathSpecificity(i.Path, j.Path); cmp != 0 {
		return cmp > 0
	}
	// Equal specificity: fall back to a deterministic lexicographic order.
	return i.Path < j.Path
}

// comparePathSpecificity returns a value > 0 when a is more specific than b,
// < 0 when b is more specific than a, and 0 when both are equally specific.
func comparePathSpecificity(a, b string) int {
	sa := splitPathSegments(a)
	sb := splitPathSegments(b)

	n := len(sa)
	if len(sb) < n {
		n = len(sb)
	}
	for k := 0; k < n; k++ {
		ra, rb := segmentRank(sa[k]), segmentRank(sb[k])
		if ra != rb {
			return ra - rb
		}
	}
	return len(sa) - len(sb)
}

func splitPathSegments(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

// segmentRank scores a single path segment: static literals rank highest, named
// parameters lower, and greedy wildcard/plus parameters lowest.
func segmentRank(seg string) int {
	if seg == "" {
		return 0
	}
	switch seg[0] {
	case ':':
		return 1
	case '*', '+':
		return 0
	default:
		return 2
	}
}
