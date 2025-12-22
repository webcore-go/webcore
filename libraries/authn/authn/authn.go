package authn

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/semanggilab/webcore-go/app/config"
	"github.com/semanggilab/webcore-go/app/core"
	"github.com/semanggilab/webcore-go/app/helper"
	"github.com/semanggilab/webcore-go/app/loader/auth"
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
	libmanager := core.Instance().LibraryManager
	lName := "auth.store:" + context.Config.Auth.Store
	loader, ok := libmanager.GetLoader(lName)
	if !ok {
		return fmt.Errorf("LibraryLoader %s tidak ditemukan", lName)
	}

	// Initialize module components
	library, err := libmanager.LoadSingletonFromLoader(loader, context, config)
	if err != nil {
		return fmt.Errorf("Library AuthStore %s tidak ditemukan %v", lName, err)
	}

	authstore := library.(auth.IAuthStore)
	storeWrapper := auth.NewStoreWrapper(authstore.GetStore())
	a.Authenticator = auth.NewAuthenticator(a.Validator, storeWrapper)

	// lzName := "authz:" + strings.ToLower(context.Config.Auth.Control)
	// zloader, ok := libmanager.GetLoader(lzName)
	// if !ok {
	// 	return fmt.Errorf("LibraryLoader tidak ditemukan %s", lzName)
	// }

	// // Initialize module components
	// zlibrary, err := libmanager.LoadSingletonFromLoader(zloader, context, config)
	// if err != nil {
	// 	return fmt.Errorf("Setup Authentication middleware %v", err)
	// }

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
			return c.Status(fiber.StatusUnauthorized).JSON(helper.APIError{
				HttpCode:  fiber.StatusUnauthorized,
				ErrorCode: 2,
				ErrorName: "UNAUTHORIZED",
				Message:   err.Error(),
			})
		}

		if err := a.Authenticator.Check(c); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(helper.APIError{
				HttpCode:  fiber.StatusUnauthorized,
				ErrorCode: 2,
				ErrorName: "UNAUTHORIZED",
				Message:   err.Error(),
			})
		}

		if err := a.Authorizer.Check(a.Authenticator.Loader.GetLoadedUser(), c.Method(), c.Path()); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(helper.APIError{
				HttpCode:  fiber.StatusUnauthorized,
				ErrorCode: 2,
				ErrorName: "UNAUTHORIZED",
				Message:   err.Error(),
			})
		}

		return c.Next()
	}
}

func (a *AuthN) Uninstall() error {
	return nil
}
