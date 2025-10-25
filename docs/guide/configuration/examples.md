# Configuration Examples

[[toc]]

## Overview

This page provides ready-to-use EasyP configuration examples for common scenarios. Each example is annotated with explanations and can be used as a starting point for your project.

## Basic Configurations

### Minimal Configuration

The simplest possible configuration for a basic protobuf project:

```yaml
version: v1alpha

lint:
  use: [MINIMAL]

generate:
  inputs:
    - directory: "proto"
  plugins:
    - name: go
      out: ./gen
      opts:
        paths: source_relative
```

**Use case:** Small projects, prototypes, getting started with EasyP.

### Standard Project Configuration

A balanced configuration suitable for most projects:

```yaml
version: v1alpha

lint:
  use: [DEFAULT]
  service_suffix: "Service"
  enum_zero_value_suffix: "UNSPECIFIED"
  allow_comment_ignores: true

deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1

generate:
  inputs:
    - directory: "proto"
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/company/myproject
    - name: go-grpc
      out: ./gen/go
      opts:
        paths: source_relative
        require_unimplemented_servers: false

breaking:
  against_git_ref: "main"
```

**Use case:** Standard API development with Go, moderate linting, basic dependencies.

### Strict Configuration

Maximum quality checks for production APIs:

```yaml
version: v1alpha

lint:
  use: 
    - DEFAULT
    - COMMENTS
    - UNARY_RPC
  service_suffix: "Service"
  enum_zero_value_suffix: "UNSPECIFIED"
  allow_comment_ignores: false  # No exceptions allowed

generate:
  inputs:
    - directory: "api"
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/company/api
    - name: go-grpc
      out: ./gen/go
      opts:
        paths: source_relative
        require_unimplemented_servers: true  # Force implementation

breaking:
  against_git_ref: "production"
```

**Use case:** Production APIs, public-facing services, high-quality requirements.

## Language-Specific Configurations

### Go Backend Service

Full-featured Go service with gRPC and REST gateway:

```yaml
version: v1alpha

lint:
  use: [DEFAULT]
  service_suffix: "Service"
  enum_zero_value_suffix: "UNSPECIFIED"

deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
  - github.com/bufbuild/protoc-gen-validate@v0.10.1

generate:
  inputs:
    - directory: "api"
  plugins:
    # Go protobuf generation
    - name: go
      out: ./internal/pb
      opts:
        paths: source_relative
        module: github.com/company/service
      with_imports: true
    
    # gRPC service generation
    - name: go-grpc
      out: ./internal/pb
      opts:
        paths: source_relative
        require_unimplemented_servers: false
    
    # REST gateway generation
    - name: grpc-gateway
      out: ./internal/pb
      opts:
        paths: source_relative
        generate_unbound_methods: true
        logtostderr: true
    
    # OpenAPI documentation
    - name: openapiv2
      out: ./docs/api
      opts:
        simple_operation_ids: true
        json_names_for_fields: true
        output_format: yaml
        allow_repeated_fields_in_body: true
    
    # Validation
    - name: validate-go
      out: ./internal/pb
      opts:
        paths: source_relative
        lang: go

breaking:
  against_git_ref: "main"
  ignore:
    - "internal/"
```

### Python Microservice

Configuration for Python-based microservice:

```yaml
version: v1alpha

lint:
  use: [DEFAULT]

deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1

generate:
  inputs:
    - directory: "protos"
  plugins:
    # Python protobuf generation
    - name: python
      out: ./src/generated
      opts:
        python_package: "myservice.proto"
    
    # Python gRPC generation
    - name: python-grpc
      out: ./src/generated
      opts:
        python_package: "myservice.proto"
    
    # Python type stubs
    - name: mypy
      out: ./src/generated
      opts:
        mypy_out: "./src/generated"
    
    # Python betterproto (alternative)
    - name: python-betterproto
      out: ./src/generated_betterproto
      opts:
        package: "myservice"
```

### TypeScript/JavaScript Frontend

Configuration for web frontend applications:

```yaml
version: v1alpha

lint:
  use: [BASIC]

deps:
  - github.com/googleapis/googleapis

generate:
  inputs:
    - directory: "api"
  plugins:
    # TypeScript definitions
    - name: ts
      out: ./src/generated
      opts:
        declaration: true
        target: "es2020"
        module: "esnext"
        force_server: true
        force_client: true
    
    # gRPC-Web client
    - name: grpc-web
      out: ./src/generated
      opts:
        import_style: "typescript"
        mode: "grpcweb"
    
    # Alternative: Connect-Web (modern)
    - name: connect-web
      out: ./src/connect
      opts:
        target: "ts"
        import_extension: ".js"
```

### Multi-Language Project

Supporting multiple languages from the same proto files:

```yaml
version: v1alpha

lint:
  use: [DEFAULT]

deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1

generate:
  inputs:
    - directory: "proto"
  plugins:
    # Go
    - name: go
      out: ./gen/go
      opts:
        paths: source_relative
        module: github.com/company/api
    - name: go-grpc
      out: ./gen/go
      opts:
        paths: source_relative
    
    # Python
    - name: python
      out: ./gen/python
    - name: python-grpc
      out: ./gen/python
    
    # TypeScript
    - name: ts
      out: ./gen/typescript
      opts:
        declaration: true
    
    # Java
    - name: java
      out: ./gen/java
      opts:
        java_package: "com.company.api"
    
    # C++
    - name: cpp
      out: ./gen/cpp
      opts:
        cpp_std: "c++17"
```

## Architecture-Specific Configurations

### Microservices Architecture

Configuration for a microservices setup with shared types:

```yaml
version: v1alpha

lint:
  use: [DEFAULT]
  service_suffix: "Service"
  ignore_only:
    # Allow different conventions for external APIs
    SERVICE_SUFFIX:
      - "proto/external/"

deps:
  # Shared company protos
  - github.com/company/shared-protos@v2.0.0
  # Google common types
  - github.com/googleapis/googleapis@common-protos-1_3_1
  # Service mesh types
  - github.com/envoyproxy/protoc-gen-validate@v0.10.1

generate:
  inputs:
    # Local service definitions
    - directory: "api/service"
    
    # Shared types from dependency
    - git_repo:
        url: "github.com/company/shared-protos@v2.0.0"
        sub_directory: "types"
        out: "gen/shared"
    
    # External vendor APIs
    - git_repo:
        url: "github.com/stripe/stripe-proto"
        sub_directory: "proto"
        out: "gen/vendor/stripe"
  
  plugins:
    - name: go
      out: ./internal/pb
      opts:
        paths: source_relative
        module: github.com/company/order-service
      with_imports: false  # Don't regenerate shared types
    
    - name: go-grpc
      out: ./internal/pb
      opts:
        paths: source_relative

breaking:
  against_git_ref: "main"
  ignore:
    - "internal/"      # Internal APIs can break
    - "experimental/"  # Experimental features
```

### Monorepo Configuration

Configuration for a monorepo with multiple services:

```yaml
version: v1alpha

lint:
  use: [DEFAULT]
  service_suffix: "Service"
  # Different rules for different parts
  ignore_only:
    COMMENT_SERVICE:
      - "services/legacy/"
    SERVICE_SUFFIX:
      - "services/external/"

deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1

generate:
  inputs:
    # Common types
    - directory:
        path: "common"
        root: "proto"
    
    # Service-specific APIs
    - directory:
        path: "services/user/api"
        root: "proto"
    - directory:
        path: "services/order/api"
        root: "proto"
    - directory:
        path: "services/payment/api"
        root: "proto"
  
  plugins:
    - name: go
      out: ./gen/go
      opts:
        paths: import
        module: github.com/company/platform
      with_imports: true
    
    - name: go-grpc
      out: ./gen/go
      opts:
        paths: import

breaking:
  against_git_ref: "main"
  ignore:
    - "proto/internal/"
    - "proto/experimental/"
```

### API Gateway Configuration

Configuration for an API gateway that aggregates multiple services:

```yaml
version: v1alpha

lint:
  use: [DEFAULT]

deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1

generate:
  inputs:
    # Gateway API definitions
    - directory: "api/gateway"
    
    # Import upstream service APIs
    - git_repo:
        url: "github.com/company/user-service@v1.0.0"
        sub_directory: "api"
        out: "upstream/user"
    - git_repo:
        url: "github.com/company/order-service@v1.0.0"
        sub_directory: "api"
        out: "upstream/order"
    - git_repo:
        url: "github.com/company/payment-service@v1.0.0"
        sub_directory: "api"
        out: "upstream/payment"
  
  plugins:
    # Gateway implementation
    - name: go
      out: ./internal/gateway
      opts:
        paths: source_relative
        module: github.com/company/api-gateway
    
    - name: grpc-gateway
      out: ./internal/gateway
      opts:
        paths: source_relative
        standalone: true
        grpc_api_configuration: "api/gateway/config.yaml"
    
    # Public OpenAPI spec
    - name: openapiv2
      out: ./public/openapi
      opts:
        simple_operation_ids: true
        merge_file_name: "api"
        output_format: yaml
        allow_merge: true
        merge_file_name: "merged"
```

## Development Workflow Configurations

### Development Environment

Permissive configuration for rapid development:

```yaml
version: v1alpha

lint:
  use: [MINIMAL]  # Minimal rules during development
  allow_comment_ignores: true
  ignore:
    - "experiments/"
    - "scratch/"
    - "tmp/"

deps:
  # Use latest versions during development
  - github.com/googleapis/googleapis
  - github.com/grpc-ecosystem/grpc-gateway

generate:
  inputs:
    - directory: "proto"
  plugins:
    - name: go
      out: ./gen
      opts:
        paths: source_relative
      with_imports: true  # Regenerate everything

# No breaking check during development
```

### CI/CD Pipeline Configuration

Strict configuration for CI/CD:

```yaml
version: v1alpha

lint:
  use: 
    - DEFAULT
    - COMMENTS
    - UNARY_RPC
  service_suffix: "Service"
  enum_zero_value_suffix: "UNSPECIFIED"
  allow_comment_ignores: false  # No exceptions in CI

deps:
  # Pinned versions for reproducible builds
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
  - github.com/bufbuild/protoc-gen-validate@v0.10.1

generate:
  inputs:
    - directory: "api"
  plugins:
    - name: go
      out: ./gen
      opts:
        paths: source_relative
        module: github.com/company/api
    - name: go-grpc
      out: ./gen
      opts:
        paths: source_relative
        require_unimplemented_servers: true

breaking:
  against_git_ref: "origin/main"  # Always check against remote
```

### Migration Configuration

Configuration for gradually migrating from another tool:

```yaml
version: v1alpha

lint:
  use: [MINIMAL]  # Start minimal
  # Gradually enable more rules
  # use: [MINIMAL, BASIC]
  # use: [DEFAULT]
  
  # Ignore legacy code initially
  ignore:
    - "legacy/"
    - "v1/"
  
  # Disable rules that conflict with existing code
  except:
    - FIELD_LOWER_SNAKE_CASE  # Existing uses camelCase
    - SERVICE_SUFFIX          # Services don't have suffix yet
  
  # Allow fixing issues gradually
  allow_comment_ignores: true
  ignore_only:
    PACKAGE_VERSION_SUFFIX:
      - "proto/v1/"  # v1 doesn't use version suffix

deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1

generate:
  inputs:
    - directory: "proto/v2"  # New structure
    - directory: "proto/v1"  # Legacy structure
  plugins:
    - name: go
      out: ./gen
      opts:
        paths: source_relative

breaking:
  against_git_ref: "main"
  ignore:
    - "proto/v1/"  # Don't check legacy code
```

## Special Use Cases

### Private Repository Configuration

Working with private Git repositories:

```yaml
version: v1alpha

lint:
  use: [DEFAULT]

# Private dependencies
# Requires Git authentication setup
deps:
  # Private GitHub
  - github.com/company/private-protos@v1.0.0
  
  # Private GitLab
  - gitlab.company.com/platform/api-definitions@v2.0.0
  
  # Private Bitbucket
  - bitbucket.org/company/shared-types@v3.0.0
  
  # On-premise Git
  - git.company.internal/protos/common@main

generate:
  inputs:
    - directory: "api"
    
    # Private remote repository
    - git_repo:
        url: "github.com/company/private-protos@v1.0.0"
        sub_directory: "public"
        out: "vendor/private"
  
  plugins:
    - name: go
      out: ./gen
      opts:
        paths: source_relative
```

### Vendor-Specific Configuration

Configuration for vendor/third-party proto files:

```yaml
version: v1alpha

lint:
  use: [DEFAULT]
  # Don't lint vendor files
  ignore:
    - "vendor/"
    - "third_party/"

deps:
  # Common vendor APIs
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
  - github.com/cncf/udpa@main
  - github.com/envoyproxy/protoc-gen-validate@v0.10.1

generate:
  inputs:
    # Your APIs
    - directory: "api"
    
    # Vendor APIs (read-only)
    - directory: "vendor"
  
  plugins:
    - name: go
      out: ./gen
      opts:
        paths: source_relative
        module: github.com/company/service
      # Don't regenerate vendor code
      with_imports: false
```

### Testing Configuration

Configuration optimized for testing:

```yaml
version: v1alpha

lint:
  use: [MINIMAL]
  # Ignore test directories
  ignore:
    - "testdata/"
    - "mocks/"
    - "fixtures/"
  # Allow test-specific patterns
  allow_comment_ignores: true

generate:
  inputs:
    - directory: "api"
    - directory: "testdata/protos"
  
  plugins:
    # Generate with test helpers
    - name: go
      out: ./gen
      opts:
        paths: source_relative
    
    # Generate mocks
    - name: go-grpc-mock
      out: ./mocks
      opts:
        paths: source_relative
        package: "mocks"
```

## Tips and Best Practices

### Configuration Management

1. **Start Simple:** Begin with minimal configuration and add complexity as needed.

2. **Version Control:** Always commit both `easyp.yaml` and `easyp.lock`.

3. **Environment-Specific:** Use different configs for dev/staging/production:
   ```bash
   easyp -cfg dev.easyp.yaml generate
   easyp -cfg prod.easyp.yaml lint
   ```

4. **Comments:** Document your configuration choices:
   ```yaml
   lint:
     # Using MINIMAL during migration phase
     # TODO: Switch to DEFAULT after fixing legacy code
     use: [MINIMAL]
   ```

5. **Gradual Adoption:** Enable rules progressively:
   ```yaml
   # Phase 1: Basic structure
   use: [MINIMAL]
   
   # Phase 2: Naming conventions
   use: [MINIMAL, BASIC]
   
   # Phase 3: Full compliance
   use: [DEFAULT, COMMENTS]
   ```

### Performance Optimization

1. **Selective Generation:** Only generate what you need:
   ```yaml
   generate:
     plugins:
       - name: go
         with_imports: false  # Don't regenerate dependencies
   ```

2. **Parallel Processing:** EasyP processes plugins in parallel by default.

3. **Caching:** Dependencies are cached locally for faster builds.

4. **Ignore Patterns:** Exclude unnecessary files from processing:
   ```yaml
   lint:
     ignore:
       - "**/*_test.proto"
       - "**/*.backup.proto"
   ```

This configuration examples guide provides templates for various use cases. Adapt these examples to your specific requirements and gradually refine your configuration as your project evolves.