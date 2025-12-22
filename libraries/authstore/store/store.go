package store

import (
	"github.com/semanggilab/webcore-go/app/loader/auth"
)

// StorageABAC is a temporary struct for unmarshaling ABAC configuration.
type StorageABAC struct {
	Users     []auth.UserAuthInfoABAC `mapstructure:"users"`
	Resources []auth.ResourceInfoABAC `mapstructure:"resources"`
}

func (c *StorageABAC) SetEnvBindings() map[string]string {
	return map[string]string{"users": "USERS", "resources": "RESOURCES"}
}

func (c *StorageABAC) SetDefaults() map[string]any {
	return map[string]any{"users": []auth.UserAuthInfoABAC{}, "resources": []auth.ResourceInfoABAC{}}
}

// StorageRBAC is a temporary struct for unmarshaling RBAC configuration.
type StorageRBAC struct {
	Users     []auth.UserAuthInfoRBAC `mapstructure:"users"`
	Resources []auth.ResourceInfoRBAC `mapstructure:"resources"`
}

func (c *StorageRBAC) SetEnvBindings() map[string]string {
	return map[string]string{"users": "USERS", "resources": "RESOURCES"}
}

func (c *StorageRBAC) SetDefaults() map[string]any {
	return map[string]any{"users": []auth.UserAuthInfoRBAC{}, "resources": []auth.ResourceInfoRBAC{}}
}

type Storage struct {
	Users     []auth.IUserAuthInfo
	Resources []auth.IResourceInfo
}

type AuthStore struct {
	Backend auth.IStore

	ControlType string
	Storage     *Storage
	Loaded      bool
}

func (y *AuthStore) SetBackend(backend auth.IStore) {
	y.Backend = backend
}

func (y *AuthStore) Install(args ...any) error {
	// Tidak melakukan apa-apa
	return nil
}

func (y *AuthStore) Uninstall() error {
	// Tidak melakukan apa-apa
	return nil
}

func (y *AuthStore) GetStore() auth.IStore {
	return y.Backend
}
