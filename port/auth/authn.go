package auth

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type IAuthenticationManager interface {
	GetAuthenticatonHandler() fiber.Handler
}

type IAuthValidator interface {
	Name() string
	GetValue() string
	ValidateKey(ctx *fiber.Ctx) error
	VerifyUser(ctx *fiber.Ctx, userKey string, userInfo IUserAuthInfo) (bool, error)
}

type Authenticator struct {
	Loader    IStoreWrapper
	Validator IAuthValidator
}

func NewAuthenticator(validator IAuthValidator, loader IStoreWrapper) *Authenticator {
	return &Authenticator{
		Loader:    loader,
		Validator: validator,
	}
}

func (a *Authenticator) Check(ctx *fiber.Ctx) error {
	// userKey := a.Validator.GetValue()
	err := a.Loader.CheckUser(ctx, a.Validator)
	if err != nil {
		return err
	}

	userInfo := a.Loader.GetLoadedUser()
	if userInfo == nil {
		return fmt.Errorf("User not found: nil")
	}

	return nil
}

type IUserAuthInfo interface {
	GetControlType() string // 'RBAC' or 'ABAC'
}

type UserAuthInfo struct {
}

type UserAuthInfoRBAC struct {
	UserId   string   `mapstructure:"key"`         // used by Api Key and JWT
	Username *string  `mapstructure:"user"`        // used by Basic Auth
	Password *string  `mapstructure:"password"`    // used by Basic Auth
	Groups   []string `mapstructure:"groups"`      // used by JWT Auth
	Roles    []string `mapstructure:"permissions"` // combination of roles from all user groups owned by user
}

func (u1 *UserAuthInfoRBAC) GetControlType() string {
	return "RBAC"
}

func (u1 *UserAuthInfoRBAC) GetUserID() string {
	return u1.UserId
}

type PolicyABAC struct {
	Effect    string // 'Allow' or 'Deny'
	Action    string
	Condition []ConditionABAC // condition with 'AND' operator (Nested and OR operation not supported yet)
}

type ConditionABAC struct {
	Attribute string
	Operator  string
	Value     any
}

type UserAuthInfoABAC struct {
	UserAuthInfo
	UserId   string       `mapstructure:"key"`      // used by Api Key and JWT
	Username *string      `mapstructure:"user"`     // used by Basic Auth
	Password *string      `mapstructure:"password"` // used by Basic Auth
	Groups   []string     `mapstructure:"groups"`   // used by JWT Auth
	Policies []PolicyABAC `mapstructure:"policies"`
}

func (u2 *UserAuthInfoABAC) GetControlType() string {
	return "ABAC"
}

func (u2 *UserAuthInfoABAC) GetUserID() string {
	return u2.UserId
}
