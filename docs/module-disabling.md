# Module Disabling Feature

This document describes how to disable specific modules in the WebCoreGo framework using the `disabled` configuration option.

## Overview

The module disabling feature allows you to prevent specific modules from being loaded or initialized, even if they are present in the module directories or specified in other loading methods.

## Configuration

### Basic Configuration

Add the `disabled` field to your `modules` configuration section:

```yaml
modules:
  disabled:
    - module-a
    - module-b
    - old-feature
  base_path: "./modules"
```

### Environment Variable Override

You can also override the disabled modules using environment variables:

```bash
export MODULES_DISABLED='["module-a","module-b"]'
```

## How It Works

### Disabled Module Detection

The module loader checks the `disabled` list before attempting to load modules through the following methods:

1. **LoadModuleFromGit**: Checks if the repository name (extracted from repoURL) is in the disabled list
2. **LoadModuleFromPackage**: Checks if the module path is in the disabled list
3. **AutoLoadModules**: Checks if individual module files (extracted from filenames) are in the disabled list
4. **LoadModuleFromPath**: Checks if the module name (from the module interface) is in the disabled list

### Behavior When Disabled

When a module is disabled:

- **LoadModuleFromGit**: Returns an error indicating the module is disabled
- **LoadModuleFromPackage**: Returns an error indicating the module is disabled  
- **AutoLoadModules**: Skips the module entirely and logs a warning message
- **LoadModuleFromPath**: Returns an error indicating the module is disabled

## Usage Examples

### Example 1: Disabling Specific Modules

```yaml
modules:
  disabled:
    - payment-gateway  # Disable payment processing module
    - reporting        # Disable reporting module
  base_path: "./modules"
```

### Example 2: Disabling During Development

```yaml
modules:
  disabled:
    - debug-tools      # Disable debug tools in production
  base_path: "./modules"
```

### Example 3: Conditional Disabling

```yaml
# Development config
modules:
  disabled:
    - production-only-feature
  base_path: "./modules"

# Production config  
modules:
  # No disabled modules - all modules load
  base_path: "./modules"
```

## Implementation Details

### Helper Functions

The module loader uses the following helper functions:

1. **`isModuleDisabled(moduleName string) bool`**: Checks if a module name is in the disabled list
2. **`loadModulesFromDirectoryWithDisabledCheck(dirPath, basePath string) error`**: Loads modules from a directory, skipping disabled ones

### Error Messages

When a disabled module is encountered:

```go
return fmt.Errorf("module '%s' is disabled in configuration", moduleName)
```

### Logging

When modules are skipped during auto-loading:

```go
slog.Info("Info: skipping disabled module '%s' from %s\n", moduleName, basePath)
```

## Best Practices

### 1. Use Descriptive Module Names

Choose clear, descriptive names for your modules to make disabling easier:

```yaml
modules:
  disabled:
    - payment-processing    # Good: Clear purpose
    - user-authentication  # Good: Clear purpose
    - mod1                 # Bad: Not descriptive
```

### 2. Group Related Disabling

If you need to disable multiple related modules, consider a naming convention:

```yaml
modules:
  disabled:
    - payment-*
    # This would disable: payment-gateway, payment-processor, payment-webhooks
```

### 3. Document Disabled Modules

Add comments to your configuration file explaining why modules are disabled:

```yaml
modules:
  disabled:
    - legacy-reporting     # Disabled: replaced by new-analytics module
    - debug-tools          # Disabled: security concern in production
  base_path: "./modules"
```

### 4. Use Environment-Specific Configs

Create different configuration files for different environments:

```yaml
# config/development.yaml
modules:
  disabled:
    - production-only-feature
  base_path: "./modules"

# config/production.yaml
modules:
  # No disabled modules - all modules load in production
  base_path: "./modules"
```

## Troubleshooting

### Common Issues

1. **Module still loads despite being disabled**
   - Check the module name in the disabled list matches exactly
   - Verify the configuration is loaded correctly
   - Check for typos in module names

2. **Error when trying to load disabled module manually**
   - This is expected behavior - disabled modules cannot be loaded
   - Remove the module from the disabled list or use a different loading method

3. **Module not found in disabled list**
   - Ensure the module name is spelled correctly
   - Check if the module is loaded through a different method

### Debug Mode

Enable debug logging to see which modules are being skipped:

```go
// In your application setup
log.SetOutput(os.Stdout)
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

## Migration Guide

### From Previous Versions

If you're upgrading from a version without module disabling:

1. Add the `disabled` field to your modules configuration
2. Test that your application still works as expected
3. Gradually add modules to the disabled list as needed

### Backward Compatibility

The disabling feature is fully backward compatible:

- Existing configurations without `disabled` field will work unchanged
- Modules not in the disabled list will load normally
- No breaking changes to existing APIs

## Future Enhancements

Potential future improvements:

1. **Pattern-based disabling**: Support wildcards and regex patterns in the disabled list
2. **Dependency-aware disabling**: Automatically disable dependent modules
3. **Runtime enabling/disabling**: Enable/disable modules without restarting the application
4. **Module metadata**: Add metadata to modules (e.g., "experimental", "deprecated") for easier management