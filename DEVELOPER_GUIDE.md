# Developer Quick Reference Guide

## 🚀 **New Features Added**

### 1. **Multi-Model Configuration**
```json
{
  "agents": {
    "defaults": {
      "models": [
        { "provider": "z.ai", "model": "glm-4.7" },
        { "provider": "openrouter", "model": "z.ai/glm-4.7" },
        { "provider": "anthropic", "model": "claude-3.5-sonnet" }
      ]
    }
  }
}
```

### 2. **Automatic Migration**
- Detects old format (`model` + `provider`)
- Automatically converts to new format (`models` array)
- Preserves all other config settings

### 3. **Fallback System**
- Tries model candidates in order
- Falls back to next model if one fails
- Logs fallback attempts for debugging

## 🧪 **Testing Commands**

### Run All Tests
```bash
./run_all_tests.sh
```

### Validate Test Structure
```bash
./validate_tests.sh
```

### Migration Testing
```bash
./test_migration.sh
./test_migration_syntax.sh
```

## 🔧 **Key Functions**

### Migration Functions
- `IsNewFormat(cfg)` - Check if config uses new format
- `NeedsMigration(cfg)` - Check if migration is needed
- `MigrateToNewFormat(cfg)` - Convert old to new format

### Model Functions
- `cfg.Agents.Defaults.ModelCandidates()` - Get model candidates
- `spec.ResolvedModel()` - Resolve model with provider
- `cfg.PrepareAgentModels()` - Prepare all model configurations

### Provider Functions
- `createProviderWithFallback(cfg)` - Create provider with fallback

## 📊 **Test Coverage**

| Component | Test Functions | Coverage |
|-----------|---------------|----------|
| Migration System | 4 | ✅ Complete |
| Main Application | 4 | ✅ Complete |
| Config Models | 7 | ✅ Complete |
| Config Agents | 6 | ✅ Complete |
| Subagent Profiles | 3 | ✅ Complete |
| Agent Loop | 5 | ✅ Complete |
| **Total** | **29** | **✅ Excellent** |

## 🎯 **Usage Examples**

### Check Migration Status
```go
if migrate.NeedsMigration(cfg) {
    fmt.Println("Migration needed")
}
```

### Get Model Candidates
```go
candidates := cfg.Agents.Defaults.ModelCandidates()
for _, candidate := range candidates {
    fmt.Println("Trying:", candidate)
}
```

### Migrate Config
```go
err := migrate.MigrateToNewFormat(cfg)
if err != nil {
    log.Fatal("Migration failed:", err)
}
```

## 🔍 **Debugging Tips**

### Check Config Format
```bash
picoclaw status
```

### Enable Debug Mode
```bash
picoclaw gateway --debug
```

### Check Migration Logs
```bash
# Look for these messages:
🔄 Detected old config format, upgrading...
✓ Config upgraded to new format
```

## 🚨 **Common Issues**

### Migration Not Triggered
- Check if config has both `model` and `provider` fields
- Verify config file permissions
- Check for syntax errors in config

### Model Failing
- Check API keys for all providers
- Verify model names are correct
- Check network connectivity
- Enable debug mode for detailed logs

### Display Issues
- Status command shows multiple models when available
- Falls back to single model display if only one candidate

## 📋 **Development Workflow**

1. **Make Changes**
2. **Run Tests**: `./validate_tests.sh`
3. **Check Migration**: `./test_migration.sh`
4. **Verify Build**: Ensure no syntax errors
5. **Commit Changes**

## 🎉 **Best Practices**

- Always test with `./validate_tests.sh` before committing
- Use table-driven tests for comprehensive coverage
- Include edge cases in test scenarios
- Log migration attempts for debugging
- Preserve backward compatibility in all changes