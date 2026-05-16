package authn

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/adapter/authsession/session"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/out"
	"github.com/webcore-go/webcore/infra/config"
	"github.com/webcore-go/webcore/infra/logger"
	"github.com/webcore-go/webcore/port/auth"
)

type AuthN struct {
	Validator     auth.IAuthValidator
	Authenticator *auth.Authenticator
	Authorizer    *auth.Authorization
}

func NewAuthN() *AuthN {
	return &AuthN{}
}

func (a *AuthN) SetValidator(validator auth.IAuthValidator) {
	a.Validator = validator
}

// Install library
func (a *AuthN) Install(args ...any) error {
	config := args[1].(config.AuthConfig)

	if a.Validator == nil {
		return fmt.Errorf("Authentication validator is not set")
	}

	if config.Type != a.Validator.Name() {
		return fmt.Errorf("Type in Config(%s) and Validator Name(%s) does not match", config.Type, a.Validator.Name())
	}

	context := args[0].(*core.AppContext)
	/*loader, e := context.GetDefaultLibraryLoader("authstorage")
	if e != nil {
		return e
	}*/

	// Initialize AuthStore
	// library, err := context.LoadSingletonInstance(loader, context, config)
	library, err := context.StartDefaultSingletonInstance("authstorage", context, config)
	if err != nil {
		return err
	}

	logger.Info("Library Authentication Storage loaded", "storage", library)

	authstore := library.(auth.IAuthStore)
	storeWrapper := auth.NewStoreWrapper(authstore.GetStore())

	/*loader2, e := context.GetDefaultLibraryLoader("authsession")
	if e != nil {
		return e
	}*/

	// library2, err := context.LoadSingletonInstance(loader2, context, config)
	library2, err := context.StartDefaultSingletonInstance("authsession", context, config)
	if err != nil {
		return err
	}

	logger.Info("Library Authentication Session Manager loaded", "session", library2)

	authsession := library2.(*session.AuthSession)

	a.Authenticator = auth.NewAuthenticator(config, a.Validator, storeWrapper, authsession)

	// authz := zlibrary.(auth.IAuthorizationManager)
	authorizer, err := auth.NewAuthorization(storeWrapper)
	if err != nil {
		return err
	}
	a.Authorizer = authorizer

	return nil
}

func (a *AuthN) GetAuthenticatonHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := a.Validator.ValidateKey(c); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(out.Error(fiber.StatusUnauthorized, 2, "UNAUTHORIZED", err.Error()))
		}

		if err := a.Authenticator.Check(c); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(out.Error(fiber.StatusUnauthorized, 2, "UNAUTHORIZED", err.Error()))
		}

		if err := a.Authorizer.Check(a.Authenticator.AuthStore.GetLoadedUser(), c.Method(), c.Path()); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(out.Error(fiber.StatusUnauthorized, 2, "UNAUTHORIZED", err.Error()))
		}

		return c.Next()
	}
}

func (a *AuthN) Uninstall() error {
	return nil
}
