# Breaking Changes Detection

[[toc]]

EasyP's breaking changes detection helps you maintain backward compatibility in your protobuf APIs by automatically identifying changes that could break existing clients. This is crucial for maintaining stable APIs in production environments.

## Overview

The breaking changes checker compares your current protobuf files against a previous version (typically from another Git branch) and identifies modifications that could cause compatibility issues for existing clients.

### Key Features

- **Git-based Comparison**: Compare against any Git reference (branch, tag, or commit)
- **Comprehensive Analysis**: Checks services, messages, enums, fields, and imports
- **Selective Ignore**: Skip specific directories from breaking change analysis
- **Detailed Reports**: Clear error messages with file locations and line numbers

## How It Works

The breaking changes detector follows this process:

1. **Checkout Comparison Branch**: Retrieves proto files from the specified Git reference
2. **Parse Both Versions**: Analyzes current and previous proto file structures
3. **Compare Entities**: Systematically checks all protobuf elements for breaking changes
4. **Generate Report**: Produces detailed issue reports with locations and descriptions

## Detection Level

EasyP implements **WIRE+ level** breaking change detection:
- ✅ **Full wire format compatibility** - ensures old and new versions can exchange data
- ✅ **Element deletion detection** - catches deleted services, messages, fields, etc.
- ✅ **Type safety** - detects incompatible type changes
- ❌ **Field/method renames** - currently not detected (planned for future releases)
- ❌ **File-level changes** - package moves, file options not checked yet

This provides strong compatibility guarantees while being less strict than some tools that check generated code compatibility.

## Configuration

Configure breaking changes detection in your `easyp.yaml`:

```yaml
version: v1alpha

breaking:
  # Git reference to compare against (branch, tag, or commit hash)
  against_git_ref: "main"
  
  # Directories to ignore during breaking changes analysis
  ignore:
    - "experimental"
    - "internal/proto"
    - "vendor"
```

### Configuration Options

| Option | Description | Default | Required |
|--------|-------------|---------|----------|
| `against_git_ref` | Git reference to compare against | `"master"` | No |
| `ignore` | List of directories to exclude from analysis | `[]` | No |

## Usage

### Basic Usage

Compare current changes against the main branch:

```bash
easyp breaking --against main
```

### Using Configuration File

With a custom configuration file:

```bash
easyp -cfg my-config.yaml breaking
```

### Override Git Reference

Override the configured branch:

```bash
easyp breaking --against feature/new-api
```

## Detection Level

EasyP currently implements **WIRE+ level** breaking change detection, which provides comprehensive wire format compatibility plus some generated code protections:

### Comparison with Buf Categories

| Check Type | Buf WIRE | Buf WIRE_JSON | Buf FILE | EasyP Current |
|------------|----------|---------------|----------|---------------|
| **Element Deletions** |
| Service deletion | ❌ | ❌ | ✅ | ✅ |
| RPC method deletion | ❌ | ❌ | ✅ | ✅ |
| Message deletion | ❌ | ❌ | ✅ | ✅ |
| Field deletion (by number) | ✅ | ✅ | ✅ | ✅ |
| Enum deletion | ❌ | ❌ | ✅ | ✅ |
| Enum value deletion | ✅ | ✅ | ✅ | ✅ |
| OneOf deletion | ❌ | ❌ | ✅ | ✅ |
| Import deletion | ❌ | ❌ | ✅ | ✅ |
| **Type Changes** |
| Field type change | ✅ | ✅ | ✅ | ✅ |
| RPC request/response type | ✅ | ✅ | ✅ | ✅ |
| Optional/required changes | ✅ | ✅ | ✅ | ✅ |
| **Naming (Generated Code)** |
| Field rename (same number) | ❌ | ✅ | ✅ | ❌ |
| Enum value rename | ❌ | ✅ | ✅ | ✅ |
| **File Structure** |
| Package change | ✅ | ✅ | ✅ | ❌ |
| File options (go_package, etc) | ❌ | ❌ | ✅ | ❌ |
| Moving types between files | ❌ | ❌ | ✅ | ❌ |

### What This Means

**✅ EasyP WILL detect:**
- All wire format breaking changes
- Deletion of services, methods, messages, fields
- Type changes that break serialization
- Enum value renames (same number, different name)

**❌ EasyP will NOT detect:**
- Field renames (same number, different name)
- Package name changes
- File option changes (go_package, java_package, etc.)
- Moving types between files in the same package

## Breaking Change Rules

EasyP detects the following types of breaking changes:

### Comparison with Other Tools

| Detection Level | Description | EasyP Support |
|----------------|-------------|---------------|
| **WIRE** | Wire format compatibility only | ✅ **Full support** |
| **WIRE+** | Wire + element deletion detection | ✅ **Current level** |
| **FILE** | Generated code compatibility | ❌ Partial (planned) |
## Breaking Change Rules

EasyP detects the following categories of breaking changes. Each rule has detailed documentation with examples:

### 🚨 Service and RPC Changes

| Rule | Description | Status |
|------|-------------|---------|
| [SERVICE_NO_DELETE](./rules/service-no-delete.md) | Services cannot be deleted | ✅ Implemented |
| [RPC_NO_DELETE](./rules/rpc-no-delete.md) | RPC methods cannot be deleted | ✅ Implemented |
| [RPC_SAME_REQUEST_TYPE](./rules/rpc-same-request-type.md) | RPC request types cannot be changed | ✅ Implemented |
| [RPC_SAME_RESPONSE_TYPE](./rules/rpc-same-response-type.md) | RPC response types cannot be changed | ✅ Implemented |

### 📦 Message and Field Changes

| Rule | Description | Status |
|------|-------------|---------|
| [MESSAGE_NO_DELETE](./rules/message-no-delete.md) | Messages cannot be deleted | ✅ Implemented |
| [FIELD_NO_DELETE](./rules/field-no-delete.md) | Fields cannot be deleted | ✅ Implemented |
| [FIELD_SAME_TYPE](./rules/field-same-type.md) | Field types cannot be changed | ✅ Implemented |
| [FIELD_SAME_CARDINALITY](./rules/field-same-cardinality.md) | Field optionality (optional/required) cannot be changed | ✅ Implemented |

### 🔢 Enum Changes

| Rule | Description | Status |
|------|-------------|---------|
| [ENUM_NO_DELETE](./rules/enum-no-delete.md) | Enums cannot be deleted | ✅ Implemented |
| [ENUM_VALUE_NO_DELETE](./rules/enum-value-no-delete.md) | Enum values cannot be deleted | ✅ Implemented |
| [ENUM_VALUE_SAME_NAME](./rules/enum-value-same-name.md) | Enum value names cannot be changed | ✅ Implemented |

### 🔗 OneOf Changes

| Rule | Description | Status |
|------|-------------|---------|
| [ONEOF_NO_DELETE](./rules/oneof-no-delete.md) | OneOf fields cannot be deleted | ✅ Implemented |
| [ONEOF_FIELD_NO_DELETE](./rules/oneof-field-no-delete.md) | Fields within oneofs cannot be deleted | ✅ Implemented |
| [ONEOF_FIELD_SAME_TYPE](./rules/oneof-field-same-type.md) | OneOf field types cannot be changed | ✅ Implemented |

### 📥 Import Changes

| Rule | Description | Status |
|------|-------------|---------|
| [IMPORT_NO_DELETE](./rules/import-no-delete.md) | Import statements cannot be removed | ✅ Implemented |

## Not Currently Detected

The following changes are **NOT detected** by EasyP (but may break generated code):

| Change Type | Example | Impact |
|-------------|---------|---------|
| Field renaming | `string name = 1` → `string full_name = 1` | Generated code breaks |
| Package changes | `package v1` → `package v2` | Import paths change |
| File options | `option go_package = "old"` → `option go_package = "new"` | Generated code location |
| Moving between files | Message moved to different .proto file | Import dependencies |

## Detailed Rules Documentation

For comprehensive examples and migration strategies, see the individual rule documentation:

- **Service Changes**: [SERVICE_NO_DELETE](./rules/service-no-delete.md), [RPC_NO_DELETE](./rules/rpc-no-delete.md), [RPC_SAME_REQUEST_TYPE](./rules/rpc-same-request-type.md), [RPC_SAME_RESPONSE_TYPE](./rules/rpc-same-response-type.md)
- **Message Changes**: [MESSAGE_NO_DELETE](./rules/message-no-delete.md), [FIELD_NO_DELETE](./rules/field-no-delete.md), [FIELD_SAME_TYPE](./rules/field-same-type.md), [FIELD_SAME_CARDINALITY](./rules/field-same-cardinality.md)
- **Enum Changes**: [ENUM_NO_DELETE](./rules/enum-no-delete.md), [ENUM_VALUE_NO_DELETE](./rules/enum-value-no-delete.md), [ENUM_VALUE_SAME_NAME](./rules/enum-value-same-name.md)
- **OneOf Changes**: [ONEOF_NO_DELETE](./rules/oneof-no-delete.md), [ONEOF_FIELD_NO_DELETE](./rules/oneof-field-no-delete.md), [ONEOF_FIELD_SAME_TYPE](./rules/oneof-field-same-type.md)
- **Import Changes**: [IMPORT_NO_DELETE](./rules/import-no-delete.md)

Each rule includes:
- ❌ **Bad examples** with actual breaking changes
- ✅ **Good alternatives** showing safe approaches  
- 🔧 **Migration strategies** for handling necessary changes
- 📋 **Real error messages** from EasyP

## Quick Examples

### ✅ Safe Changes (Always Allowed)
```proto
// Adding new elements is always safe
message User {
  string name = 1;
  string email = 2;
  string phone = 3;  // ✅ New field - safe
}

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc GetUserProfile(GetUserRequest) returns (UserProfile);  // ✅ New RPC - safe
}

enum Status {
  STATUS_UNSPECIFIED = 0;
  STATUS_ACTIVE = 1;
  STATUS_PENDING = 2;  // ✅ New enum value - safe
}
```

### ❌ Breaking Changes (Always Detected)
```proto
// Deletions and type changes break compatibility
message User {
  string name = 1;
  // ❌ Deleted field - BREAKING
}

service UserService {
  // ❌ Deleted RPC method - BREAKING  
  rpc GetUser(GetUserRequestV2) returns (GetUserResponse);  // ❌ Changed request type - BREAKING
}
```

### Scenario 3: Breaking Changes (Currently Not Detected)

```proto
// 🟡 Breaks generated code but passes EasyP checks
message User {
  string user_name = 1;    // Renamed from "name"
  string user_email = 2;   // Renamed from "email"
}

service UserService {
  rpc GetUserProfile(GetUserRequest) returns (GetUserResponse);  // Renamed from GetUser
}
```

## Output Format

### Text Format (Default)

```
services.proto:45:1: Previously present RPC "DeleteUser" on service "UserService" was deleted. (BREAKING_CHECK)
messages.proto:15:3: Previously present field "2" with name "email" on message "User" was deleted. (BREAKING_CHECK)
```

### JSON Format

```bash
easyp breaking --against main --format json
```

```json
{
  "path": "services.proto",
  "position": {
    "line": 45,
    "column": 1
  },
  "source_name": "",
  "message": "Previously present RPC \"DeleteUser\" on service \"UserService\" was deleted.",
  "rule_name": "BREAKING_CHECK"
}
```

## Best Practices

### 1. Regular Checks
Run breaking change detection in your CI/CD pipeline:

```yaml
# GitHub Actions example
- name: Check for breaking changes
  run: easyp breaking --against origin/main
```

### 2. Branch Protection
Use breaking change checks to protect important branches:

```yaml
# Only allow PRs that don't introduce breaking changes
if: github.event_name == 'pull_request'
run: |
  easyp breaking --against origin/main
  if [ $? -eq 1 ]; then
    echo "Breaking changes detected!"
    exit 1
  fi
```

### 3. Versioning Strategy
When breaking changes are necessary:

- Create a new package version (e.g., `myservice.v2`)
- Maintain the old version during migration period
- Use deprecation notices in the old version

### 4. Ignore Patterns
Use ignore patterns wisely:

```yaml
breaking:
  ignore:
    - "experimental/**"      # Experimental APIs
    - "internal/**"          # Internal-only APIs
    - "**/testing/**"        # Test utilities
```

## Common Issues and Solutions

### Issue: "Repository does not exist"
**Solution**: Ensure you're running the command in a Git repository with the specified branch available.

### Issue: "Cannot find git ref"
**Solution**: Verify the branch/tag name exists and is accessible:
```bash
git branch -a  # List all branches
git tag        # List all tags
```

### Issue: False Positives in Generated Code
**Solution**: Add generated directories to ignore list:
```yaml
breaking:
  ignore:
    - "generated/**"
    - "**/pb/**"
```

### Issue: Large Number of Changes
**Solution**: For major refactoring, consider:
1. Creating a new API version
2. Implementing changes incrementally
3. Using feature flags for gradual rollout

## Integration Examples

### CI/CD Pipeline Integration

#### GitHub Actions

```yaml
name: API Compatibility Check

on:
  pull_request:
    branches: [ main ]

jobs:
  breaking-changes:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Fetch full history
      
      - name: Install EasyP
        run: |
          curl -sSfL https://raw.githubusercontent.com/easyp-tech/easyp/main/install.sh | sh
          
      - name: Check for breaking changes
        run: |
          ./bin/easyp breaking --against origin/main
```

#### GitLab CI

```yaml
breaking-changes:
  stage: test
  image: easyp/lint:latest
  script:
    - git fetch origin main
    - easyp breaking --against origin/main
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
```

### Pre-commit Hook

```bash
#!/bin/sh
# .git/hooks/pre-push

protected_branch='main'
current_branch=$(git symbolic-ref HEAD | sed -e 's,.*/\(.*\),\1,')

if [ $current_branch = $protected_branch ]; then
    echo "Running breaking changes check..."
    easyp breaking --against HEAD~1
    if [ $? -eq 1 ]; then
        echo "❌ Breaking changes detected. Push rejected."
        exit 1
    fi
    echo "✅ No breaking changes detected."
fi
```

## Troubleshooting

### Debug Mode
Enable debug logging for detailed information:

```bash
easyp --debug breaking --against main
```

### Manual Comparison
For complex cases, you can manually inspect the comparison:

```bash
# Compare specific files
git show main:path/to/file.proto > old_version.proto
easyp lint old_version.proto  # Validate old version
easyp lint current_file.proto  # Validate current version
```

### Performance Optimization
For large repositories:

```bash
# Limit scope to specific paths
easyp breaking --against main --path api/
```

The breaking changes detection in EasyP provides a robust foundation for maintaining API compatibility while allowing your protobuf schemas to evolve safely over time.