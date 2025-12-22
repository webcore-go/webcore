package core

import (
	"fmt"
	"reflect"

	"github.com/semanggilab/webcore-go/app/loader"
	"github.com/semanggilab/webcore-go/app/logger"
)

type LibraryLoader interface {
	ClassName() string
	Init(args ...any) (loader.Library, error)
}

type LibraryManager struct {
	Loaders   map[string]LibraryLoader
	Libraries map[string]map[string]loader.Library // Loaded libraries
}

func CreateLibraryManager(loaders map[string]LibraryLoader) *LibraryManager {
	// Avoid redundant LibraryLoader
	// newLoaders := CheckSingleLoader("LibraryLoader", loaders)

	return &LibraryManager{
		// Loaders: newLoaders,
		Loaders:   loaders,
		Libraries: make(map[string]map[string]loader.Library),
	}
}

func (lm *LibraryManager) Destroy() error {
	for name, libMap := range lm.Libraries {
		for key, library := range libMap {
			_, err := lm.unload(name, library, &libMap, key)
			if err != nil {
				logger.Warn(err.Error())
			}
		}
	}
	return nil
}

func (lm *LibraryManager) GetLoader(name string) (LibraryLoader, bool) {
	loader, ok := lm.Loaders[name]
	return loader, ok
}

func (lm *LibraryManager) GetSingleton(name string) (loader.Library, bool) {
	return lm.GetLibrary(name, true, nil)
}

func (lm *LibraryManager) GetInstance(name string, key string) (loader.Library, bool) {
	return lm.GetLibrary(name, false, &key)
}

// GetLibrary retrieves a library instance
func (lm *LibraryManager) GetLibrary(name string, singleton bool, key *string) (loader.Library, bool) {
	// Check if library type exists
	libMap, ok := lm.Libraries[name]
	if ok {
		if singleton {
			// Check if instance exists
			if ptr, ok := libMap["default"]; ok {
				return ptr, true
			}
		} else {
			// Check if instance exists
			if ptr, ok := libMap[*key]; ok {
				return ptr, true
			}
		}
	}

	return nil, false
}

func (lm *LibraryManager) LoadFromLoader(load LibraryLoader, name string, singleton bool, key *string, args ...any) (loader.Library, error) {
	// Check if library type exists
	libMap, ok := lm.Libraries[name]
	if !ok {
		library, err := load.Init(args...)
		if err != nil {
			return nil, err
		}

		// Create library map for this type
		libMap = make(map[string]loader.Library)

		// Store instance
		if singleton {
			libMap["default"] = library
		} else {
			if key == nil {
				d := "default"
				key = &d
			}
			libMap[*key] = library
		}

		lm.Libraries[name] = libMap
		// logger.Debug("LoadFromLoader: Map BARU untuk", "name", name)
		return library, nil
	}

	if singleton {
		// Check if instance exists
		if ptr, ok := libMap["default"]; ok {
			return ptr, nil
		}
	} else {
		// Check if instance exists
		if ptr, ok := libMap[*key]; ok {
			return ptr, nil
		}
	}

	library, err := load.Init(args)
	if err != nil {
		return nil, err
	}

	// Store instance
	if singleton {
		lm.Libraries[name]["default"] = library
		// logger.Debug("LoadFromLoader: Buat Instance BARU untuk", "name", name, "key", "default")
	} else {
		if key == nil {
			d := "default"
			key = &d
		}
		lm.Libraries[name][*key] = library
		// logger.Debug("LoadFromLoader: Buat Instance BARU untuk", "name", name, "key", *key)
	}

	return library, nil
}

func (lm *LibraryManager) LoadSingletonFromLoader(loader LibraryLoader, args ...any) (loader.Library, error) {
	return lm.LoadFromLoader(loader, loader.ClassName(), true, nil, args...)
}

func (lm *LibraryManager) LoadInstanceFromLoader(loader LibraryLoader, key string, args ...any) (loader.Library, error) {
	return lm.LoadFromLoader(loader, loader.ClassName(), false, &key, args...)
}

func (lm *LibraryManager) LoadSingleton(libType reflect.Type, args ...any) (loader.Library, error) {
	return lm.LoadLibrary(libType, true, nil, args...)
}

func (lm *LibraryManager) LoadInstance(libType reflect.Type, key string, args ...any) (loader.Library, error) {
	return lm.LoadLibrary(libType, false, &key, args...)
}

func (lm *LibraryManager) UnloadSingleton(libType reflect.Type) (loader.Library, error) {
	return lm.UnloadLibrary(libType, true, nil)
}

func (lm *LibraryManager) UnloadInstance(libType reflect.Type, key string) (loader.Library, error) {
	return lm.UnloadLibrary(libType, false, &key)
}

// LoadLibrary creates or retrieves a library instance
func (lm *LibraryManager) LoadLibrary(libType reflect.Type, singleton bool, key *string, args ...any) (loader.Library, error) {
	var zero loader.Library

	// Get the type name
	if libType.Kind() == reflect.Ptr {
		libType = libType.Elem()
	}
	name := libType.Name()

	// Check if library type exists
	libMap, ok := lm.Libraries[name]
	if !ok {
		// Create new instance
		lib := reflect.New(libType).Interface()
		if library, ok := lib.(loader.Library); ok {
			err := library.Install(args...)
			if err != nil {
				return zero, err
			}

			if libConnector, ok2 := lib.(loader.Connector); ok2 {
				err = libConnector.Connect()
				if err != nil {
					return zero, err
				}
			}

			// Create library map for this type
			libMap = make(map[string]loader.Library)
			lm.Libraries[name] = libMap

			// Store instance
			if singleton {
				libMap["default"] = library
			} else {
				if key == nil {
					d := "default"
					key = &d
				}
				libMap[*key] = library
			}
			return library, nil
		}
		return zero, fmt.Errorf("type %T does not implement Library interface", lib)
	}

	if singleton {
		// Check if instance exists
		if ptr, ok := libMap["default"]; ok {
			return ptr, nil
		}
	} else {
		// Check if instance exists
		if ptr, ok := libMap[*key]; ok {
			return ptr, nil
		}
	}

	// Create new instance
	lib := reflect.New(libType).Interface()
	if library, ok := lib.(loader.Library); ok {
		err := library.Install(args...)
		if err != nil {
			return zero, err
		}

		// Store instance
		if singleton {
			libMap["default"] = library
		} else {
			if key == nil {
				d := "default"
				key = &d
			}
			libMap[*key] = library
		}
		return library, nil
	}

	return zero, fmt.Errorf("type %T does not implement Library interface", lib)
}

func (lm *LibraryManager) UnloadLibrary(libType reflect.Type, singleton bool, key *string) (loader.Library, error) {
	var zero loader.Library

	// Get the type name
	if libType.Kind() == reflect.Ptr {
		libType = libType.Elem()
	}
	name := libType.Name()

	// Check if library type exists
	libMap, ok := lm.Libraries[name]
	if !ok {
		return zero, fmt.Errorf("library type %s not found", name)
	}

	// Determine the key to use
	libKey := "default"
	if !singleton {
		if key == nil {
			return zero, fmt.Errorf("key is required for non-singleton libraries")
		}
		libKey = *key
	}

	// Check if instance exists
	library, ok := libMap[libKey]
	if !ok {
		return zero, fmt.Errorf("library instance with key %s not found", libKey)
	}

	return lm.unload(name, library, &libMap, libKey)
}

func (lm *LibraryManager) unload(name string, library loader.Library, libMap *map[string]loader.Library, libKey string) (loader.Library, error) {
	// If it's a connector, close the connection
	if libConnector, ok := library.(loader.Connector); ok {
		err := libConnector.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close connector: %v", err)
		}
	}

	// Call destroy on the library
	err := library.Uninstall()
	if err != nil {
		return nil, fmt.Errorf("failed to destroy library: %v", err)
	}

	// Remove the library from the map
	delete(*libMap, libKey)

	// If the libMap is empty, remove it entirely
	if len(*libMap) == 0 {
		delete(lm.Libraries, name)
	}

	return library, nil
}

func GetLibraryLoader(name string) (LibraryLoader, bool) {
	return Instance().LibraryManager.GetLoader(name)
}

// LoadLibrary is a convenience function that works with concrete types
func LoadLibrary[T loader.Library](singleton bool, key *string, args ...any) (T, error) {
	var zero T
	libType := reflect.TypeOf(zero)
	lib, err := Instance().LibraryManager.LoadLibrary(libType, singleton, key, args...)
	if err != nil {
		return zero, err
	}
	return lib.(T), nil
}

func Load[T loader.Library](args ...any) (T, error) {
	return LoadLibrary[T](true, nil, args...)
}

func LoadMulti[T loader.Library](key string, args ...any) (T, error) {
	return LoadLibrary[T](false, &key, args...)
}

func UnloadLibrary[T loader.Library](singleton bool, key *string, args ...any) (T, error) {
	var zero T
	libType := reflect.TypeOf(zero)
	lib, err := Instance().LibraryManager.UnloadLibrary(libType, singleton, key)
	if err != nil {
		return zero, err
	}
	return lib.(T), nil
}

func Unload[T loader.Library](args ...any) (T, error) {
	return UnloadLibrary[T](true, nil, args...)
}

func UnloadMulti[T loader.Library](key string, args ...any) (T, error) {
	return UnloadLibrary[T](false, &key, args...)
}
