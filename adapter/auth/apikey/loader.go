package apikey

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/adapter/auth/authn"
	"github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/port"
	"github.com/webcore-go/webcore/port/auth"
)

type ApiKeyLoader struct {
	name string
}

func (a *ApiKeyLoader) SetName(name string) {
	a.name = name
}

func (a *ApiKeyLoader) Name() string {
	return a.name
}

func (a *ApiKeyLoader) Init(args ...any) (port.Library, error) {
	config := args[1].(config.AuthConfig)
	authn := &authn.AuthN{}
	authn.SetValidator(NewApiKeyValidator(config))
	err := authn.Install(args...)
	if err != nil {
		return nil, err
	}

	return authn, nil
}

type ApiKeyValidator struct {
	Header string
	Prefix string
	Key    string
}

func (a *ApiKeyValidator) Name() string {
	return "apikey"
}

func (a *ApiKeyValidator) ValidateKey(ctx *fiber.Ctx) error {
	apiKey := ctx.Get(a.Header)
	if apiKey == "" {
		// Coba dapatkan dari Authorization
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			return fmt.Errorf("Authorization header required")
		}

		// konten dimulai dengan prefiks "APIKey "
		if strings.HasPrefix(authHeader, "APIKey ") {
			apiKey = strings.TrimPrefix(authHeader, "APIKey ")
		} else {
			return fmt.Errorf("Required prefix in Authorization header is missing")
		}
	}

	if a.Prefix != "" {
		if !strings.HasPrefix(apiKey, a.Prefix) {
			return fmt.Errorf("Required prefix in Authorization header is missing")
		}
		apiKey = strings.TrimPrefix(apiKey, a.Prefix)
	}

	a.Key = apiKey
	return nil
}

func (a *ApiKeyValidator) GetValue() string {
	return a.Key
}

func (a *ApiKeyValidator) VerifyUser(ctx *fiber.Ctx, userKey string, userInfo auth.IUserAuthInfo) (bool, error) {
	if userKey == "" {
		return false, nil
	}

	rbac, ok1 := userInfo.(*auth.UserAuthInfoRBAC)
	if ok1 {
		return userKey == rbac.UserId, nil
	}

	abac, ok2 := userInfo.(*auth.UserAuthInfoABAC)
	if ok2 {
		return userKey == abac.UserId, nil
	}

	return false, nil
}

func NewApiKeyValidator(config config.AuthConfig) *ApiKeyValidator {
	return &ApiKeyValidator{
		Header: config.APIKeyHeader,
		Prefix: config.APIKeyPrefix,
	}
}
