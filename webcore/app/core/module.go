package core

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/semanggilab/webcore-go/app/config"
	"github.com/semanggilab/webcore-go/app/logger"
)

// Module represents a pluggable module interface
type Module interface {
	// Name returns the unique name of the module
	Name() string

	// Version returns the version of the module
	Version() string

	// Dependencies returns the dependencies of the module
	Dependencies() []string

	// Config returns the module-specific configuration
	Config() config.Configurable

	// Routes returns the routes provided by this module
	Routes() []*ModuleRoute

	// Services returns the services provided by this module
	Services() map[string]any

	// Repositories returns the repositories provided by this module
	Repositories() map[string]any

	// Init initializes the module with the given app and dependencies
	Init(ctx *AppContext) error

	Destroy() error
}

type ModuleRoute struct {
	Method  string
	Path    string
	Handler fiber.Handler
	Root    fiber.Router
}

// ModuleManager manages module registration and loading
type ModuleManager struct {
	mu            sync.RWMutex
	modules       map[string]Module // loaded modules
	loadedModules map[string]LoadedModule
	loaded        bool
	context       *AppContext
	config        *config.ModuleConfig
}

// LoadedModule represents a loaded module and its metadata
type LoadedModule struct {
	Name      string
	Path      string
	Module    Module
	Plugin    *plugin.Plugin
	LoadedAt  string
	DependsOn []string
}

// CreateModuleManager creates a new central registry instance
func CreateModuleManager(config *config.ModuleConfig, modules []Module) *ModuleManager {
	manager := &ModuleManager{
		modules:       make(map[string]Module),
		loadedModules: make(map[string]LoadedModule),
		config:        config,
	}

	// Avoid redundant Modules
	newModules := CheckSingleLoader("Module", modules)

	// Register all modules
	for _, module := range newModules {
		err := manager.Register(module)
		if err != nil {
			logger.Fatal(err.Error())
		}
	}

	return manager
}

func (lm *ModuleManager) Destroy() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	for _, module := range lm.modules {
		if err := module.Destroy(); err != nil {
			logger.Warn(err.Error())
		}
	}

	lm.modules = make(map[string]Module)
	lm.loaded = false
	lm.context = nil
	return nil
}

// IsLoaded checks if all modules have been initialized
func (r *ModuleManager) IsLoaded() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.loaded
}

// Register registers a new module with the central registry
func (r *ModuleManager) Register(module Module) error {
	return r.registerModuleInstance(module, "", nil)
}

func LoadConfig[T config.Configurable](name string, c T, file string, ext string, path []string) error {
	return config.LoadConfig(name, c, file, ext, path, true)
}

func LoadDefaultConfig[T config.Configurable](name string, c T) error {
	return config.LoadConfig(name, c, "config", "yaml", []string{}, true)
}

// GetModule retrieves a registered module by name
func (r *ModuleManager) GetModule(name string) (Module, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	module, exists := r.modules[name]
	if !exists {
		return nil, fmt.Errorf("module '%s' not found", name)
	}

	return module, nil
}

// ListModules returns all registered module names
func (r *ModuleManager) ListModules() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.modules))
	for name := range r.modules {
		names = append(names, name)
	}

	return names
}

// InitializeModules initializes all registered modules with the app and dependencies
func (r *ModuleManager) InitializeModules() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, module := range r.modules {
		if err := module.Init(r.context); err != nil {
			return fmt.Errorf("failed to initialize module '%s': %v", name, err)
		}
	}

	r.loaded = true
	return nil
}

// InitializeModulesWithDependencies initializes modules in dependency order
func (r *ModuleManager) InitializeModulesWithDependencies() error {
	// Build dependency graph
	dependencyGraph := r.buildDependencyGraph()

	// Topological sort to get initialization order
	initializationOrder, err := r.topologicalSort(dependencyGraph)
	if err != nil {
		return fmt.Errorf("failed to resolve module dependencies: %v", err)
	}

	// Initialize modules in order
	for _, moduleName := range initializationOrder {
		loadedModule, exists := r.loadedModules[moduleName]
		if !exists {
			return fmt.Errorf("module '%s' not found in loaded modules", moduleName)
		}

		if err := loadedModule.Module.Init(r.context); err != nil {
			return fmt.Errorf("failed to initialize module '%s': %v", moduleName, err)
		}
	}

	return nil
}

// GetRoutes returns all routes from all registered modules
func (r *ModuleManager) GetRoutes() []*ModuleRoute {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var routes []*ModuleRoute
	for _, module := range r.modules {
		routes = append(routes, module.Routes()...)
	}

	return routes
}

// GetServices returns all services from all registered modules
func (r *ModuleManager) GetServices() map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make(map[string]any)
	for _, module := range r.modules {
		moduleServices := module.Services()
		for key, service := range moduleServices {
			// Create unique key to avoid conflicts
			uniqueKey := fmt.Sprintf("%s.%s", module.Name(), key)
			services[uniqueKey] = service
		}
	}

	return services
}

// GetRepositories returns all repositories from all registered modules
func (r *ModuleManager) GetRepositories() map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	repositories := make(map[string]any)
	for _, module := range r.modules {
		moduleRepos := module.Repositories()
		for key, repo := range moduleRepos {
			// Create unique key to avoid conflicts
			uniqueKey := fmt.Sprintf("%s.%s", module.Name(), key)
			repositories[uniqueKey] = repo
		}
	}

	return repositories
}

// AutoLoadModulesFromPath automatically loads modules from specified paths
func (r *ModuleManager) AutoLoadModulesFromPath(modulePaths []string) error {
	// Load modules from paths
	for _, path := range modulePaths {
		if err := r.LoadModuleFromPath(path); err != nil {
			return fmt.Errorf("failed to load module from path %s: %v", path, err)
		}
	}

	return nil
}

// AutoLoadModulesFromConfig automatically loads modules from configured paths
func (r *ModuleManager) AutoLoadModulesFromConfig() error {
	// Load modules from modules directory
	modulesPath := filepath.Join(r.config.BasePath, "modules")
	if _, err := os.Stat(modulesPath); !os.IsNotExist(err) {
		if err := r.loadModulesFromDirectoryWithDisabledCheck(modulesPath, "modules"); err != nil {
			return fmt.Errorf("failed to load modules from modules: %v", err)
		}
	}

	return nil
}

// LoadModuleFromPath loads a module from a file path
func (r *ModuleManager) LoadModuleFromPath(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("module file not found: %s", path)
	}

	// Load plugin
	plug, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %v", err)
	}

	// Look for module symbol
	symModule, err := plug.Lookup("Module")
	if err != nil {
		return fmt.Errorf("module symbol not found: %v", err)
	}

	// Convert to Module interface
	module, ok := symModule.(Module)
	if !ok {
		return fmt.Errorf("module does not implement Module interface")
	}

	return r.registerModuleInstance(module, path, plug)
}

// LoadModuleFromGit loads a module from a git repository
func (r *ModuleManager) LoadModuleFromGit(repoURL, branch, path string) error {
	// Check if the module is disabled
	moduleName := getRepoName(repoURL)
	if r.isModuleDisabled(moduleName) {
		logger.Warn("Module is disabled in configuration", "name", moduleName)
		return nil
	}

	// In a real implementation, this would:
	// 1. Clone the repository
	// 2. Checkout the specified branch
	// 3. Build the module
	// 4. Load the compiled binary/plugin

	// For now, we'll simulate the process
	tempPath := filepath.Join(r.config.BasePath, "modules", moduleName)

	// Create temporary directory if it doesn't exist
	if err := os.MkdirAll(tempPath, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}

	// In a real implementation, you would:
	// - git clone repoURL tempPath
	// - git checkout branch
	// - go build -buildmode=plugin -o module.so path/to/module

	// For demonstration, we'll assume the module is already built
	modulePath := filepath.Join(tempPath, "module.so")

	return r.LoadModuleFromPath(modulePath)
}

// LoadModuleFromPackage loads a module using go module path
func (r *ModuleManager) LoadModuleFromPackage(modulePath string) error {
	// Check if the module is disabled
	if r.isModuleDisabled(modulePath) {
		logger.Warn("Module is disabled in configuration", "name", modulePath)
		return nil
	}

	// In a real implementation, this would:
	// 1. Download the module using go get
	// 2. Build the module as a plugin
	// 3. Load the compiled plugin

	// For now, we'll simulate the process
	tempPath := filepath.Join(r.config.BasePath, "packages", modulePath)

	// Create temporary directory if it doesn't exist
	if err := os.MkdirAll(tempPath, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}

	// In a real implementation, you would:
	// - go get modulePath
	// - go build -buildmode=plugin -o module.so ./path/to/module

	// For demonstration, we'll assume the module is already built
	modulePathFile := filepath.Join(tempPath, "module.so")

	return r.LoadModuleFromPath(modulePathFile)
}

// GetLoadedModules returns all loaded modules with their metadata
func (r *ModuleManager) GetLoadedModules() []LoadedModule {
	modules := make([]LoadedModule, 0, len(r.loadedModules))
	for _, module := range r.loadedModules {
		modules = append(modules, module)
	}
	return modules
}

// GetModuleMetadata returns metadata for a specific loaded module
func (r *ModuleManager) GetModuleMetadata(name string) (LoadedModule, error) {
	loadedModule, exists := r.loadedModules[name]
	if !exists {
		return LoadedModule{}, fmt.Errorf("module '%s' is not loaded", name)
	}
	return loadedModule, nil
}

// UnloadModule unloads a module by name
func (r *ModuleManager) UnloadModule(name string) error {
	loadedModule, exists := r.loadedModules[name]
	if !exists {
		return fmt.Errorf("module '%s' is not loaded", name)
	}

	// Unregister from central registry
	// Note: Manager doesn't have an unregister method yet
	// We'll need to add that functionality

	// Close plugin if possible
	if loadedModule.Plugin != nil {
		// Plugins can't be explicitly closed in Go
		// They will be garbage collected when no longer referenced
	}

	// Remove from loaded modules
	delete(r.loadedModules, name)

	return nil
}

// isModuleDisabled checks if a module is in the disabled list
func (r *ModuleManager) isModuleDisabled(moduleName string) bool {
	for _, disabledModule := range r.config.Disabled {
		if disabledModule == moduleName {
			return true
		}
	}
	return false
}

// validateModule validates that a module implements the required interface
func (r *ModuleManager) validateModule(module Module) error {
	// Check required methods
	methods := []string{"Name", "Version", "Init", "Destroy", "Config", "Routes", "Services", "Repositories"}

	moduleType := reflect.TypeOf(module)
	for _, methodName := range methods {
		if _, exists := moduleType.MethodByName(methodName); !exists {
			return fmt.Errorf("module missing required method: %s", methodName)
		}
	}

	// Check if name is valid
	if module.Name() == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	// Check if version is valid
	if module.Version() == "" {
		return fmt.Errorf("module version cannot be empty")
	}

	return nil
}

func (r *ModuleManager) registerModuleInstance(module Module, path string, plugin *plugin.Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if module is disabled
	if r.isModuleDisabled(module.Name()) {
		logger.Warn("Module is disabled in configuration", "name", module.Name())
		return nil
	}

	// Validate module
	if err := r.validateModule(module); err != nil {
		return fmt.Errorf("module validation failed: %v", err)
	}

	// Store loaded module
	loadedModule := LoadedModule{
		Name:      module.Name(),
		Path:      path,
		Module:    module,
		Plugin:    plugin,
		LoadedAt:  getCurrentTimestamp(),
		DependsOn: r.extractDependencies(module),
	}

	r.loadedModules[module.Name()] = loadedModule
	r.modules[module.Name()] = module

	return nil
}

// loadModulesFromDirectoryWithDisabledCheck loads modules from a directory, skipping disabled modules
func (ml *ModuleManager) loadModulesFromDirectoryWithDisabledCheck(dirPath, basePath string) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".so") {
			// Extract module name from file path
			moduleName := strings.TrimSuffix(file.Name(), ".so")

			// Check if module is disabled
			if ml.isModuleDisabled(moduleName) {
				slog.Info("skipping disabled module '%s' from %s\n", moduleName, basePath)
				continue
			}

			modulePath := filepath.Join(dirPath, file.Name())
			if err := ml.LoadModuleFromPath(modulePath); err != nil {
				// Log error but continue loading other modules
				slog.Warn("Failed to load module %s: %v\n", modulePath, err)
			}
		}
	}

	return nil
}

// extractDependencies extracts module dependencies from the module
func (r *ModuleManager) extractDependencies(module Module) []string {
	// Option #1: [HIGH-RISK] Strict implementation, all module must implement function Dependencies()
	dependencies := module.Dependencies()

	// Option #2: [LOW-RISK] Safer implementation, function Dependencies() not required
	// var dependencies []string
	//
	// // Check if the module implements a method that can provide dependencies
	// if depProvider, ok := module.(interface {
	// 	GetDependencies() []string
	// }); ok {
	// 	// Module provides its own dependencies
	// 	dependencies = depProvider.GetDependencies()
	// }

	// Remove duplicates
	uniqueDeps := make(map[string]bool)
	for _, dep := range dependencies {
		uniqueDeps[dep] = true
	}

	result := make([]string, 0, len(uniqueDeps))
	for dep := range uniqueDeps {
		result = append(result, dep)
	}

	return result
}

// Helper functions
func getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

func getRepoName(url string) string {
	// Extract repository name from URL
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}

// buildDependencyGraph builds a dependency graph from loaded modules
func (r *ModuleManager) buildDependencyGraph() map[string][]string {
	graph := make(map[string][]string)

	for name, loadedModule := range r.loadedModules {
		graph[name] = loadedModule.DependsOn
	}

	return graph
}

// topologicalSort performs topological sort on a dependency graph
func (r *ModuleManager) topologicalSort(graph map[string][]string) ([]string, error) {
	var result []string
	var inDegree = make(map[string]int)
	var queue []string

	// Calculate in-degree for each node
	for node := range graph {
		inDegree[node] = 0
	}

	for _, dependencies := range graph {
		for _, dep := range dependencies {
			inDegree[dep]++
		}
	}

	// Find nodes with no incoming edges
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	// Process nodes
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		for _, neighbor := range graph[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycles
	if len(result) != len(graph) {
		return nil, fmt.Errorf("cyclic dependency detected")
	}

	return result, nil
}
