package basic

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/semanggilab/webcore-go/app/loader"
	"github.com/semanggilab/webcore-go/app/loader/auth"
	"github.com/semanggilab/webcore-go/app/logger"
	"github.com/semanggilab/webcore-go/lib/auth/authn"
)

type BasicAuthLoader struct {
	name string
}

func (a *BasicAuthLoader) SetClassName(name string) {
	a.name = name
}

func (a *BasicAuthLoader) ClassName() string {
	return a.name
}

func (a *BasicAuthLoader) Init(args ...any) (loader.Library, error) {

	authn := &authn.AuthN{}
	authn.SetValidator(&BasicAuthValidator{})
	err := authn.Install(args...)
	if err != nil {
		return nil, err
	}

	return authn, nil
}

type BasicAuthValidator struct {
	Header string
	Prefix string
	Key    string
}

func (a *BasicAuthValidator) Name() string {
	return "basic"
}

func (a *BasicAuthValidator) ValidateKey(ctx *fiber.Ctx) error {
	var apiKey string

	// Coba dapatkan dari Authorization
	authHeader := ctx.Get("Authorization")
	if authHeader == "" {
		return fmt.Errorf("Authorization header required")
	}

	// konten dimulai dengan prefiks "Basic "
	if strings.HasPrefix(authHeader, "Basic ") {
		apiKey = strings.TrimPrefix(authHeader, "Basic ")
	} else {
		return fmt.Errorf("Required prefix in Authorization header is missing")
	}

	a.Key = apiKey
	return nil
}

func (a *BasicAuthValidator) GetValue() string {
	return a.Key
}

func (a *BasicAuthValidator) GetUserPassword(userKey string) (string, string) {
	decoded, err := base64.StdEncoding.DecodeString(userKey)
	if err != nil {
		logger.Info("Basic Auth invalid key", "error", err)
	}

	credentials := string(decoded)
	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}

func (a *BasicAuthValidator) VerifyUser(ctx *fiber.Ctx, userKey string, userInfo auth.IUserAuthInfo) (bool, error) {
	if userKey == "" {
		return false, nil
	}

	username, password := a.GetUserPassword(userKey)
	if username == "" || password == "" {
		return false, nil
	}

	rbac, ok1 := userInfo.(*auth.UserAuthInfoRBAC)
	if ok1 {
		if rbac.Username != nil && *rbac.Username != username {
			return false, nil
		}

		if rbac.Password != nil && *rbac.Password != password {
			return false, nil
		}

		return true, nil
	}

	abac, ok2 := userInfo.(*auth.UserAuthInfoABAC)
	if ok2 {
		if abac.Username != nil && *abac.Username != username {
			return false, nil
		}

		if abac.Password != nil && *abac.Password != password {
			return false, nil
		}

		return true, nil
	}

	return false, nil
}
