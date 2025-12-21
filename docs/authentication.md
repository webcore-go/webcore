# Authentication Documentation

This document describes the authentication system in WebCoreGo, which supports both JWT and API key authentication methods.

## Overview

The authentication middleware provides flexible authentication options:

- **JWT Authentication**: Traditional token-based authentication with user roles and permissions
- **API Key Authentication**: Simple API key-based authentication for service-to-service communication

## Configuration

### API Configuration

The authentication settings are configured in the `api` section of your `config.yaml`:

```yaml
api:
  type: "jwt"  # Options: "jwt", "api_key"
  secret_key: your-secret-key-here
  expires_in: 86400  # 24 hours in seconds
  
  # API Key specific settings (only used when type is "api_key")
  api_key_header: "X-API-Key"  # Header name for API key
  api_key_prefix: ""  # Optional prefix for API key validation
```

### Environment Variables

You can also override these settings using environment variables:

```bash
# JWT Settings
API_TYPE=jwt
API_SECRET_KEY=your-secret-key-here
API_EXPIRES_IN=86400

# API Key Settings
API_API_KEY_HEADER=X-API-Key
API_API_KEY_PREFIX=
```

## Authentication Methods

### 1. JWT Authentication

JWT authentication uses JSON Web Tokens for authentication and authorization.

#### How to Use

1. **Generate a JWT Token** (typically done by an authentication service)
2. **Include in Request Headers**:
   ```
   Authorization: Bearer <your-jwt-token>
   ```

#### Example JWT Payload

```json
{
  "user_id": 123,
  "role": "admin",
  "permissions": ["read", "write", "delete"],
  "exp": 1640995200,
  "iat": 1640908800
}
```

#### JWT Claims Available in Context

After successful JWT authentication, the following claims are available in the request context:

- `user_id`: The user's unique identifier
- `user_role`: The user's role (e.g., "admin", "user")
- `user_permissions`: Array of user permissions
- `auth_type`: Set to "jwt"

### 2. API Key Authentication

API key authentication uses a simple API key for authentication, suitable for service-to-service communication.

#### How to Use

There are three ways to provide the API key:

1. **Authorization Header (Bearer format)**:
   ```
   Authorization: Bearer <your-api-key>
   ```

2. **Authorization Header (APIKey format)**:
   ```
   Authorization: APIKey <your-api-key>
   ```

3. **Custom Header** (default: `X-API-Key`):
   ```
   X-API-Key: <your-api-key>
   ```

#### API Key Configuration

You can customize the API key validation:

```yaml
api:
  type: "api_key"
  api_key_header: "X-API-Key"  # Custom header name
  api_key_prefix: "service-"    # Optional prefix for validation
```

With the prefix configuration, an API key like `service-abc123` would be validated as `abc123`.

#### Context Data Available After API Key Authentication

- `api_key`: The API key value
- `auth_type`: Set to "api_key"

## Middleware Usage

### Global Authentication

Authentication middleware is applied globally to all routes:

```go
// In internal/app/app.go
authMiddleware := middleware.NewAuthFromConfig(a.config.API)
a.fiberApp.Use(authMiddleware)
```

### Route-Specific Authentication

You can also apply authentication to specific routes:

```go
app.Get("/api/protected", middleware.NewAuthFromConfig(config.API), handler.ProtectedHandler)
```

### Role-Based Access Control

#### RoleRequired Middleware

Create middleware that requires specific roles:

```go
// Require admin role
app.Get("/api/admin", middleware.RoleRequired("admin"), adminHandler)

// Require either admin or manager role
app.Get("/api/management", middleware.RoleRequired("admin", "manager"), managementHandler)
```

#### PermissionRequired Middleware

Create middleware that requires specific permissions:

```go
// Require read permission
app.Get("/api/data", middleware.PermissionRequired("read"), dataHandler)

// Require write permission
app.Post("/api/data", middleware.PermissionRequired("write"), dataHandler)
```

## Helper Functions

The authentication package provides helper functions to access user information:

```go
// Get authentication type
authType := middleware.GetAuthType(c)

// Get user information
userID := middleware.GetUserID(c)
userRole := middleware.GetUserRole(c)
userPermissions := middleware.GetUserPermissions(c)

// Get API key (for API key authentication)
apiKey := middleware.GetAPIKey(c)
```

## Example Usage in Handlers

### Basic Authentication Check

```go
func GetUserHandler(c *fiber.Ctx) error {
    userID := middleware.GetUserID(c)
    if userID == nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "User not authenticated",
        })
    }
    
    // Continue with handler logic
    return c.JSON(fiber.Map{"user_id": userID})
}
```

### Role-Based Access Control

```go
func AdminOnlyHandler(c *fiber.Ctx) error {
    userRole := middleware.GetUserRole(c)
    if userRole == nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "User not authenticated",
        })
    }
    
    if userRole != "admin" {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
            "error": "Insufficient permissions - admin role required",
        })
    }
    
    // Continue with admin-only logic
    return c.JSON(fiber.Map{"message": "Welcome, admin!"})
}
```

### API Key Authentication

```go
func ServiceHandler(c *fiber.Ctx) error {
    authType := middleware.GetAuthType(c)
    if authType != "api_key" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "API key authentication required",
        })
    }
    
    apiKey := middleware.GetAPIKey(c)
    // Validate API key against your database/service
    // Continue with service logic
    
    return c.JSON(fiber.Map{"message": "Service request processed"})
}
```

## Security Considerations

### JWT Security

1. **Use Strong Secrets**: Always use a strong, randomly generated secret key
2. **Set Appropriate Expiration**: Set reasonable expiration times for tokens
3. **Validate Signing Method**: The middleware validates JWT signing methods
4. **Handle Token Revocation**: Implement a token blacklist if needed

### API Key Security

1. **Secure Key Generation**: Generate long, random API keys
2. **Key Rotation**: Implement API key rotation policies
3. **Scoped Permissions**: Consider adding scope/permission fields to API keys
4. **Rate Limiting**: Apply rate limiting to API key endpoints

### General Security

1. **HTTPS**: Always use HTTPS in production
2. **Header Validation**: Validate all authentication headers
3. **Error Messages**: Provide generic error messages to avoid information leakage
4. **Logging**: Log authentication attempts for security monitoring

## Migration from JWT to API Key

To switch between authentication methods, simply change the `type` in your configuration:

```yaml
# JWT Authentication
api:
  type: "jwt"
  secret_key: your-jwt-secret

# API Key Authentication  
api:
  type: "api_key"
  api_key_header: "X-API-Key"
```

The middleware will automatically handle the authentication method based on the configuration.

## Testing

### JWT Authentication Test

```bash
# Generate a JWT token (using a tool like jwt.io)
# Then make a request:
curl -H "Authorization: Bearer <your-jwt-token>" \
     http://localhost:3000/api/v1/module-a/users
```

### API Key Authentication Test

```bash
# Using Authorization header
curl -H "Authorization: Bearer <your-api-key>" \
     http://localhost:3000/api/v1/module-a/users

# Using custom header
curl -H "X-API-Key: <your-api-key>" \
     http://localhost:3000/api/v1/module-a/users
```

## Troubleshooting

### Common Issues

1. **"Authorization header required"**: Check that the Authorization header is present and properly formatted
2. **"Invalid authorization format"**: Ensure the header follows the correct format (Bearer <token>)
3. **"Invalid or expired token"**: Check JWT expiration and secret key configuration
4. **"Unsupported authentication type"**: Verify the `type` field in your configuration

### Debug Mode

Enable debug logging to troubleshoot authentication issues:

```yaml
logging:
  level: debug
```

This will provide detailed information about authentication attempts and validation.