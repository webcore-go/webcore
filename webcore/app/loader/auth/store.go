package auth

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/semanggilab/webcore-go/app/logger"
)

type IStore interface {
	GetUserAuthInfo(ctx *fiber.Ctx, validator IAuthValidator) (IUserAuthInfo, error)
	GetResourceInfo(method string, path string) (IResourceInfo, error)
}

type IStoreWrapper interface {
	CheckUser(ctx *fiber.Ctx, validator IAuthValidator) error
	GetLoadedUser() IUserAuthInfo

	CheckResource(method string, path string) (bool, error)
	GetLoadedResource() IResourceInfo
}

type IAuthStore interface {
	GetStore() IStore
}

type StoreWrapper struct {
	Store    IStore
	User     IUserAuthInfo
	Resource IResourceInfo
}

func NewStoreWrapper(store IStore) *StoreWrapper {
	return &StoreWrapper{
		Store: store,
	}
}

func (u *StoreWrapper) CheckUser(ctx *fiber.Ctx, validator IAuthValidator) error {
	userKey := validator.GetValue()
	info, err := u.Store.GetUserAuthInfo(ctx, validator) // mencari user aktif
	if err != nil {
		return fmt.Errorf("User not found: %s", userKey)
	}

	u.User = info
	return nil
}

func (u *StoreWrapper) CheckResource(method string, path string) (bool, error) {
	info, err := u.Store.GetResourceInfo(method, path) // mencari user aktif
	if err != nil {
		logger.Info(err.Error(), "method", method, "path", path)
		return false, err
	}

	u.Resource = info
	return true, nil
}

func (u *StoreWrapper) GetLoadedUser() IUserAuthInfo {
	return u.User
}

func (u *StoreWrapper) GetLoadedResource() IResourceInfo {
	return u.Resource
}
