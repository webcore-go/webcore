package core

import (
	"fmt"
	"reflect"

	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port"
)

type LibraryLoader interface {
	SetName(name string)
	Name() string
	Init(args ...any) (port.Library, error)
}

type LibraryManager struct {
	Loaders   map[string]LibraryLoader
	Libraries map[string]map[string]port.Library // Loaded libraries
}

func CreateLibraryManager(loaders map[string]LibraryLoader) *LibraryManager {
	// setName with key
	for k, v := range loaders {
		v.SetName(k)
	}

	return &LibraryManager{
		Loaders:   loaders,
		Libraries: make(map[string]map[string]port.Library),
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

func (lm *LibraryManager) GetSingletonInstance(name string) (port.Library, bool) {
	return lm.GetLibrary(name, true, nil)
}

func (lm *LibraryManager) GetInstance(name string, key string) (port.Library, bool) {
	return lm.GetLibrary(name, false, &key)
}

// GetLibrary retrieves a library instance
func (lm *LibraryManager) GetLibrary(name string, singleton bool, key *string) (port.Library, bool) {
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

func (lm *LibraryManager) LoadFromLoader(load LibraryLoader, name string, singleton bool, key *string, args ...any) (port.Library, error) {
	// Check if library type exists
	libMap, ok := lm.Libraries[name]
	if !ok {
		library, err := load.Init(args...)
		if err != nil {
			return nil, err
		}

		// Create library map for this type
		libMap = make(map[string]port.Library)

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

	library, err := load.Init(args...)
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

func (lm *LibraryManager) LoadSingletonFromLoader(loader LibraryLoader, args ...any) (port.Library, error) {
	return lm.LoadFromLoader(loader, loader.Name(), true, nil, args...)
}

func (lm *LibraryManager) LoadInstanceFromLoader(loader LibraryLoader, key string, args ...any) (port.Library, error) {
	return lm.LoadFromLoader(loader, loader.Name(), false, &key, args...)
}

func (lm *LibraryManager) LoadSingleton(libType reflect.Type, args ...any) (port.Library, error) {
	return lm.LoadLibrary(libType, true, nil, args...)
}

func (lm *LibraryManager) LoadInstance(libType reflect.Type, key string, args ...any) (port.Library, error) {
	return lm.LoadLibrary(libType, false, &key, args...)
}

func (lm *LibraryManager) UnloadSingleton(libType reflect.Type) (port.Library, error) {
	return lm.UnloadLibrary(libType, true, nil)
}

func (lm *LibraryManager) UnloadInstance(libType reflect.Type, key string) (port.Library, error) {
	return lm.UnloadLibrary(libType, false, &key)
}

// LoadLibrary creates or retrieves a library instance
func (lm *LibraryManager) LoadLibrary(libType reflect.Type, singleton bool, key *string, args ...any) (port.Library, error) {
	var zero port.Library

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
		if library, ok := lib.(port.Library); ok {
			err := library.Install(args...)
			if err != nil {
				return zero, err
			}

			if libConnector, ok2 := lib.(port.Connector); ok2 {
				err = libConnector.Connect()
				if err != nil {
					return zero, err
				}
			}

			// Create library map for this type
			libMap = make(map[string]port.Library)
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
	if library, ok := lib.(port.Library); ok {
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

func (lm *LibraryManager) UnloadLibrary(libType reflect.Type, singleton bool, key *string) (port.Library, error) {
	var zero port.Library

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

func (lm *LibraryManager) unload(name string, library port.Library, libMap *map[string]port.Library, libKey string) (port.Library, error) {
	// If it's a connector, close the connection
	if libConnector, ok := library.(port.Connector); ok {
		err := libConnector.Disconnect()
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
func LoadLibrary[T port.Library](singleton bool, key *string, args ...any) (T, error) {
	var zero T
	libType := reflect.TypeOf(zero)
	lib, err := Instance().LibraryManager.LoadLibrary(libType, singleton, key, args...)
	if err != nil {
		return zero, err
	}
	return lib.(T), nil
}

func Load[T port.Library](args ...any) (T, error) {
	return LoadLibrary[T](true, nil, args...)
}

func LoadMulti[T port.Library](key string, args ...any) (T, error) {
	return LoadLibrary[T](false, &key, args...)
}

func UnloadLibrary[T port.Library](singleton bool, key *string, args ...any) (T, error) {
	var zero T
	libType := reflect.TypeOf(zero)
	lib, err := Instance().LibraryManager.UnloadLibrary(libType, singleton, key)
	if err != nil {
		return zero, err
	}
	return lib.(T), nil
}

func Unload[T port.Library](args ...any) (T, error) {
	return UnloadLibrary[T](true, nil, args...)
}

func UnloadMulti[T port.Library](key string, args ...any) (T, error) {
	return UnloadLibrary[T](false, &key, args...)
}
