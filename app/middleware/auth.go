package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/app/out"
)

// GetAuthType returns the authentication type from the context
func GetAuthType(c *fiber.Ctx) string {
	authType := c.Locals("auth_type")
	if authType == nil {
		return "unknown"
	}
	return authType.(string)
}

// GetUserID returns the user ID from the context
func GetUserID(c *fiber.Ctx) any {
	return c.Locals("user_id")
}

// GetUserRole returns the user role from the context
func GetUserRole(c *fiber.Ctx) any {
	return c.Locals("user_role")
}

// GetUserPermissions returns the user permissions from the context
func GetUserPermissions(c *fiber.Ctx) any {
	return c.Locals("user_permissions")
}

// GetAPIKey returns the API key from the context
func GetAPIKey(c *fiber.Ctx) string {
	apiKey := c.Locals("api_key")
	if apiKey == nil {
		return ""
	}
	return apiKey.(string)
}

// RoleRequired creates a middleware to check user roles
func RoleRequired(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authType := GetAuthType(c)

		// For API key authentication, you might want to implement different role checking logic
		if authType == "apikey" {
			// In a real implementation, you might want to check API key permissions/roles
			// For now, we'll allow access if API key is valid
			return c.Next()
		}

		// For JWT authentication, check the role claims
		userRole := GetUserRole(c)
		if userRole == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(out.Error(fiber.StatusUnauthorized, 2, "UNAUTHORIZED", "User role not found in context"))
		}

		role := userRole.(string)
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(out.Error(fiber.StatusUnauthorized, 2, "UNAUTHORIZED", "Insufficient permissions"))
	}
}

// PermissionRequired creates a middleware to check user permissions
func PermissionRequired(requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authType := GetAuthType(c)

		// For API key authentication, you might want to implement different permission checking logic
		if authType == "apikey" {
			// In a real implementation, you might want to check API key permissions
			// For now, we'll allow access if API key is valid
			return c.Next()
		}

		// For JWT authentication, check the permission claims
		userPermissions := GetUserPermissions(c)
		if userPermissions == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(out.Error(fiber.StatusUnauthorized, 2, "UNAUTHORIZED", "User permissions not found in context"))
		}

		permissions := userPermissions.([]any)
		for _, permission := range permissions {
			if permission == requiredPermission {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(out.Error(fiber.StatusUnauthorized, 2, "UNAUTHORIZED", "Insufficient permissions"))
	}
}
