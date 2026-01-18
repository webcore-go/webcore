package yaml

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/adapter/authstore/store"
	"github.com/webcore-go/webcore/infra/config"
	appConfig "github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/port"
	"github.com/webcore-go/webcore/port/auth"
)

type YamlLoader struct {
	name string
}

func (a *YamlLoader) SetName(name string) {
	a.name = name
}

func (a *YamlLoader) Name() string {
	return a.name
}

func (l *YamlLoader) Init(args ...any) (port.Library, error) {
	config := args[1].(config.AuthConfig)
	backend, err := YamlBackend(config.Control)
	if err != nil {
		return nil, err
	}

	store := &store.AuthStore{}
	store.SetBackend(backend)
	err = store.Install(args...)
	if err != nil {
		return nil, err
	}

	return store, nil
}

type AuthStoreYAML struct {
	ControlType string
	// Validator   auth.IAuthValidator
	Storage *store.Storage
	Loaded  bool
}

func YamlBackend(control string) (*AuthStoreYAML, error) {
	y := &AuthStoreYAML{
		ControlType: control,
		Storage: &store.Storage{
			Users:     make([]auth.IUserAuthInfo, 0),
			Resources: make([]auth.IResourceInfo, 0),
		},
		Loaded: false,
	}

	switch control {
	case "ABAC":
		var tmp store.StorageABAC
		if err := appConfig.LoadConfig("access", &tmp, "access", "yaml", []string{}); err != nil {
			return nil, err
		}

		// Convert from concrete slice to interface slice
		y.Storage.Users = make([]auth.IUserAuthInfo, len(tmp.Users))
		for i := range tmp.Users {
			y.Storage.Users[i] = &tmp.Users[i]
		}
		y.Storage.Resources = make([]auth.IResourceInfo, len(tmp.Resources))
		for i := range tmp.Resources {
			y.Storage.Resources[i] = &tmp.Resources[i]
		}
	default:
		var tmp store.StorageRBAC
		if err := appConfig.LoadConfig("access", &tmp, "access", "yaml", []string{}); err != nil {
			return nil, err
		}

		y.Storage.Users = make([]auth.IUserAuthInfo, len(tmp.Users))
		for i := range tmp.Users {
			y.Storage.Users[i] = &tmp.Users[i]
		}
		y.Storage.Resources = make([]auth.IResourceInfo, len(tmp.Resources))
		for i := range tmp.Resources {
			y.Storage.Resources[i] = &tmp.Resources[i]
		}
	}

	y.Loaded = true
	return y, nil
}

func (y *AuthStoreYAML) GetUserAuthInfo(ctx *fiber.Ctx, validator auth.IAuthValidator) (auth.IUserAuthInfo, error) {
	if !y.Loaded {
		return nil, fmt.Errorf("File access.yaml gagal dimuat")
	}

	userKey := validator.GetValue()

	var err1 error
	for _, info := range y.Storage.Users {
		ok, err := validator.VerifyUser(ctx, userKey, info)
		if ok {
			if err == nil {
				return info, nil
			} else {
				err1 = err
			}
		}
	}

	if err1 != nil {
		return nil, err1
	}

	return nil, fmt.Errorf("Invalid or expired token %s", userKey)
}

func (y *AuthStoreYAML) cleanPath(infoPath string) string {
	// Remove parametes (everything after fist '/:')
	if idx := strings.Index(infoPath, "/:"); idx != -1 {
		infoPath = infoPath[:idx]
	}

	// Remove query string (everything after '?')
	if idx := strings.Index(infoPath, "?"); idx != -1 {
		infoPath = infoPath[:idx]
	}

	return infoPath
}

func (y *AuthStoreYAML) GetResourceInfo(method string, path string) (auth.IResourceInfo, error) {
	if !y.Loaded {
		return nil, fmt.Errorf("File access.yaml gagal dimuat")
	}

	for _, info := range y.Storage.Resources {
		infoPath := info.GetPath()
		cleanedInfoPath := y.cleanPath(infoPath)

		if method == info.GetMethod() && strings.HasPrefix(path, cleanedInfoPath) {
			return info, nil
		}
	}

	return nil, nil
}
