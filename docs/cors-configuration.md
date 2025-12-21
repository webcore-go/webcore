# CORS Configuration Documentation

This document describes how to configure Cross-Origin Resource Sharing (CORS) in the WebCoreGo framework.

## Overview

CORS is configured through the `app.cors` section in your configuration file. The middleware reads these settings and applies them to all HTTP responses.

## Configuration Structure

### Basic Configuration

```yaml
app:
  cors:
    allow_origins:
      - "http://localhost:3000"
      - "http://localhost:8080"
      - "https://yourdomain.com"
    allow_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
      - "OPTIONS"
    allow_headers:
      - "Origin"
      - "Content-Type"
      - "Accept"
      - "Authorization"
      - "X-Custom-Header"
    allow_credentials: true
    expose_headers:
      - "Content-Length"
      - "X-Request-ID"
    max_age: 86400  # 24 hours in seconds
```

### Configuration Options

#### `allow_origins`
- **Type**: `[]string`
- **Description**: List of allowed origins (domains)
- **Default**: `["*"]` (allows all origins)
- **Example**: 
  ```yaml
  allow_origins:
    - "https://example.com"
    - "https://app.example.com"
  ```

#### `allow_methods`
- **Type**: `[]string`
- **Description**: List of allowed HTTP methods
- **Default**: `["GET", "POST", "PUT", "DELETE", "OPTIONS"]`
- **Example**:
  ```yaml
  allow_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "PATCH"
    - "OPTIONS"
  ```

#### `allow_headers`
- **Type**: `[]string`
- **Description**: List of allowed headers in requests
- **Default**: `["Origin", "Content-Type", "Accept", "Authorization"]`
- **Example**:
  ```yaml
  allow_headers:
    - "Origin"
    - "Content-Type"
    - "Accept"
    - "Authorization"
    - "X-API-Key"
    - "X-Custom-Header"
  ```

#### `allow_credentials`
- **Type**: `bool`
- **Description**: Whether to allow credentials (cookies, auth headers)
- **Default**: `true`
- **Example**: `allow_credentials: true`

#### `expose_headers`
- **Type**: `[]string`
- **Description**: List of headers that can be exposed to the client
- **Default**: `["Content-Length"]`
- **Example**:
  ```yaml
  expose_headers:
    - "Content-Length"
    - "X-Request-ID"
    - "X-RateLimit-Limit"
  ```

#### `max_age`
- **Type**: `int`
- **Description**: How long the results of a preflight request can be cached (in seconds)
- **Default**: `86400` (24 hours)
- **Example**: `max_age: 3600` (1 hour)

## Environment Variables

You can override CORS settings using environment variables:

```bash
# Override origins
export APP_CORS_ALLOW_ORIGINS='["https://example.com","https://app.example.com"]'

# Override methods
export APP_CORS_ALLOW_METHODS='["GET","POST","PUT","DELETE"]'

# Override headers
export APP_CORS_ALLOW_HEADERS='["Origin","Content-Type","Authorization"]'

# Override credentials
export APP_CORS_ALLOW_CREDENTIALS=true

# Override max age
export APP_CORS_MAX_AGE=3600
```

## Usage Examples

### Development Configuration

```yaml
app:
  cors:
    allow_origins:
      - "http://localhost:3000"
      - "http://localhost:8080"
      - "http://localhost:5173"  # Vite dev server
    allow_methods:
      - "*"
    allow_headers:
      - "*"
    allow_credentials: true
    max_age: 3600  # 1 hour for development
```

### Production Configuration

```yaml
app:
  cors:
    allow_origins:
      - "https://yourapp.com"
      - "https://www.yourapp.com"
    allow_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
    allow_headers:
      - "Origin"
      - "Content-Type"
      - "Accept"
      - "Authorization"
    allow_credentials: false
    expose_headers:
      - "Content-Length"
      - "X-RateLimit-Limit"
      - "X-RateLimit-Remaining"
    max_age: 86400  # 24 hours
```

### API-Only Configuration

```yaml
app:
  cors:
    allow_origins:
      - "https://api.yourapp.com"
    allow_methods:
      - "GET"
      - "POST"
    allow_headers:
      - "Origin"
      - "Content-Type"
      - "Authorization"
    allow_credentials: false
    max_age: 7200  # 2 hours
```

## Security Considerations

### 1. **Production Security**

In production, be specific about allowed origins:

```yaml
# Good: Specific origins
allow_origins:
  - "https://yourapp.com"
  - "https://www.yourapp.com"

# Bad: Wildcard with credentials
allow_origins: ["*"]
allow_credentials: true  # This combination is insecure
```

### 2. **Credentials Handling**

If you need to use credentials, avoid wildcards:

```yaml
# Good: Specific origins with credentials
allow_origins:
  - "https://yourapp.com"
allow_credentials: true

# Bad: Wildcard with credentials (insecure)
allow_origins: ["*"]
allow_credentials: true
```

### 3. **Method and Header Restrictions**

Only allow the methods and headers your API actually uses:

```yaml
# Good: Minimal allowed methods
allow_methods:
  - "GET"
  - "POST"

# Bad: Allow all methods
allow_methods: ["*"]
```

## Troubleshooting

### Common Issues

1. **CORS Error: "No 'Access-Control-Allow-Origin' header"**
   - Check that `allow_origins` includes your frontend domain
   - Verify the configuration is loaded correctly

2. **Preflight Request Issues**
   - Ensure `allow_methods` includes the HTTP method being used
   - Check that `allow_headers` includes any custom headers
   - Verify `max_age` is set appropriately

3. **Credentials Not Working**
   - Set `allow_credentials: true`
   - Avoid using `*` in `allow_origins` when credentials are enabled

### Debug Mode

Enable debug logging to see CORS configuration:

```go
// In your application setup
log.SetOutput(os.Stdout)
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

### Testing CORS

You can test your CORS configuration using curl:

```bash
# Test preflight request
curl -X OPTIONS \
  -H "Origin: https://your-frontend.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type, Authorization" \
  http://your-api.com/endpoint

# Test actual request
curl -X GET \
  -H "Origin: https://your-frontend.com" \
  -H "Authorization: Bearer token" \
  http://your-api.com/endpoint
```

## Browser Compatibility

| Feature | Chrome | Firefox | Safari | Edge |
|---------|--------|---------|--------|------|
| `allow_origins` | ✅ | ✅ | ✅ | ✅ |
| `allow_methods` | ✅ | ✅ | ✅ | ✅ |
| `allow_headers` | ✅ | ✅ | ✅ | ✅ |
| `allow_credentials` | ✅ | ✅ | ✅ | ✅ |
| `expose_headers` | ✅ | ✅ | ✅ | ✅ |
| `max_age` | ✅ | ✅ | ✅ | ✅ |

## Migration Guide

### From Hardcoded Configuration

If you're upgrading from a version with hardcoded CORS values:

1. Add the `cors` section to your `app` configuration
2. Move your hardcoded values to the appropriate fields
3. Remove any hardcoded CORS configuration from your code

### Example Migration

**Before (hardcoded):**
```go
corsConfig := cors.Config{
    AllowOrigins:     "*",
    AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
    AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
    AllowCredentials: true,
    ExposeHeaders:    "Content-Length",
}
```

**After (config-based):**
```yaml
app:
  cors:
    allow_origins: ["*"]
    allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers: ["Origin", "Content-Type", "Accept", "Authorization"]
    allow_credentials: true
    expose_headers: ["Content-Length"]
    max_age: 86400
```

## Best Practices

1. **Environment-Specific Configs**: Use different CORS settings for development and production
2. **Least Privilege**: Only allow the origins, methods, and headers you actually need
3. **Regular Reviews**: Periodically review your CORS settings for security
4. **Testing**: Test CORS thoroughly in different browsers and environments
5. **Documentation**: Document your CORS policy for development teams

## References

- [MDN Web Docs: CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [OWASP CORS Security](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)
- [Fiber CORS Middleware](https://docs.gofiber.io/middleware/cors)