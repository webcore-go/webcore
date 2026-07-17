package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type UserLoginInfo struct {
	Username     string        `json:"user_id"`
	AccessToken  *string       `json:"access_token"`
	RefreshToken *string       `json:"refresh_token"`
	ExpiresIn    time.Duration `json:"expires_in"`
	RefreshIn    time.Duration `json:"refresh_in"`
	Groups       []string      `json:"user_role"`
	Permissions  []string      `json:"user_permissions"`
}

type IAuthSessionStore interface {
	GetSessionStore() ISessionStore
}

type ISessionStore interface {
	Save(loginInfo *UserLoginInfo) error
	Refresh(oldAccessToken string, oldRefreshToken string, loginInfo *UserLoginInfo, refreshIn time.Duration) error
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
