package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type UserLoginInfo struct {
	Username     string
	AccessToken  *string
	RefreshToken *string
	ExpiresIn    time.Duration
	RefreshIn    time.Duration
	Groups       []string
	Permissions  []string
}

type IAuthSessionStore interface {
	GetSessionStore() ISessionStore
}

type ISessionStore interface {
	Save(loginInfo *UserLoginInfo) error
	Refresh(oldAccessToken string, loginInfo *UserLoginInfo) error
	Delete(loginInfo *UserLoginInfo) error
	GetByAccessToken(accessToken string) (*UserLoginInfo, error)
	GetByRefreshToken(refreshToken string) (*UserLoginInfo, error)
	GetByUsername(username string) (*UserLoginInfo, error)
}

type IAuthSession interface {
	SetSessionStore(session ISessionStore)
	Login(ctx *fiber.Ctx, userInfo IUserAuthInfo) (*UserLoginInfo, error)
	Refresh(ctx *fiber.Ctx, userKey string) (*UserLoginInfo, error)
	Logout(ctx *fiber.Ctx, userKey string) error
}
