package auth

import (
	"fmt"
	"slices"

	"github.com/semanggilab/webcore-go/app/loader"
	"github.com/semanggilab/webcore-go/app/logger"
)

type IAuthorizationManager interface {
	GetAuthorization() IAuthorization
}

type IAuthorization interface {
	loader.Library

	Check(user IUserAuthInfo, method string, path string) error
}

type Authorization struct {
	Loader IStoreWrapper
}

func NewAuthorization(loader IStoreWrapper) (*Authorization, error) {
	return &Authorization{
		Loader: loader,
	}, nil
}

func (a *Authorization) Check(user IUserAuthInfo, method string, path string) error {
	ok, err := a.Loader.CheckResource(method, path)
	if err != nil {
		return err
	}

	if ok {
		resourceInfo := a.Loader.GetLoadedResource()
		if resourceInfo != nil {
			return resourceInfo.IsUserPermitted(user)
		}
	}

	// defaulf permission untuk resource yang tidak memiliki permission
	return nil
}

type IResourceInfo interface {
	GetAction() string
	GetMethod() string
	GetPath() string
	GetControlType() string // 'RBAC' or 'ABAC'
	IsUserPermitted(user IUserAuthInfo) error
}

type ResourceInfoRBAC struct {
	Action         string   `mapstructure:"action"`
	Path           string   `mapstructure:"path"`
	Method         string   `mapstructure:"method"`
	PermittedRoles []string `mapstructure:"permissions"`
}

func (r1 *ResourceInfoRBAC) GetControlType() string {
	return "RBAC"
}

func (r1 *ResourceInfoRBAC) GetAction() string {
	return r1.Action
}

func (r1 *ResourceInfoRBAC) GetMethod() string {
	return r1.Method
}

func (r1 *ResourceInfoRBAC) GetPath() string {
	return r1.Path
}

func (r1 *ResourceInfoRBAC) IsUserPermitted(user IUserAuthInfo) error {
	// Ensure the user auth info is compatible (RBAC).
	if user.GetControlType() != "RBAC" {
		return fmt.Errorf("Load wrong User Access Control Type User (%s) and Resource (RBAC)", user.GetControlType())
	}

	// Type assert the user to the concrete RBAC type to access roles.
	rbacUser, ok := user.(*UserAuthInfoRBAC)
	if !ok {
		// This case should ideally not be reached if GetControlType() is 'RBAC',
		// but it's a safe check to have.
		return fmt.Errorf("RBAC properties not found in user")
	}

	// Check if any of the user's roles are in the permitted set.
	for _, userRole := range rbacUser.Roles {
		if slices.Contains(r1.PermittedRoles, userRole) {
			// The user has a permitted role, grant access.
			return nil
		}
	}

	// The user has no roles that grant access to this resource.
	return fmt.Errorf("User access denied")
}

type ResourceInfoABAC struct {
	Action            string       `mapstructure:"action"`
	Path              string       `mapstructure:"path"`
	Method            string       `mapstructure:"method"`
	PermittedPolicies []PolicyABAC `mapstructure:"policies"`
}

func (r2 *ResourceInfoABAC) GetControlType() string {
	return "ABAC"
}

func (r2 *ResourceInfoABAC) GetAction() string {
	return r2.Action
}

func (r2 *ResourceInfoABAC) GetMethod() string {
	return r2.Method
}

func (r2 *ResourceInfoABAC) GetPath() string {
	return r2.Path
}

func (r2 *ResourceInfoABAC) IsUserPermitted(user IUserAuthInfo) error {
	// Ensure the user auth info is compatible (RBAC).
	if user.GetControlType() != "RBAC" {
		return fmt.Errorf("Load wrong User Access Control Type User (%s) and Resource (RBAC)", user.GetControlType())
	}

	// Type assert the user to the concrete RBAC type to access roles.
	abacUser, ok := user.(*UserAuthInfoABAC)
	if !ok {
		// This case should ideally not be reached if GetControlType() is 'RBAC',
		// but it's a safe check to have.
		return fmt.Errorf("RBAC properties not found in user")
	}

	// Check if any of the user's roles are in the permitted set.
	for _, userPolicy := range abacUser.Policies {
		if r2.IsAccessGranted(userPolicy, r2.PermittedPolicies) {
			// The user has a permitted role, grant access.
			return nil
		}
	}

	return fmt.Errorf("User access denied")
}

func (r2 *ResourceInfoABAC) IsAccessGranted(userPolicy PolicyABAC, policies []PolicyABAC) bool {
	logger.Fatal("ABAC Policy enforcement logic not implemented yet")
	return false
}
