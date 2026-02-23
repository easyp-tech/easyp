# Package Manager

[[toc]]

EasyP provides a powerful package manager for protobuf dependencies that simplifies dependency management through a decentralized, Git-based approach. Unlike centralized solutions, EasyP works directly with Git repositories, giving you complete control over your dependencies.

## Overview

The EasyP package manager follows the **Go modules philosophy** - any Git repository can serve as a package source. This approach provides several key advantages:

- **Decentralized**: No single point of failure or control
- **Security**: Direct access to source repositories
- **Flexibility**: Support for public, private, and enterprise repositories
- **Reproducibility**: Lock files ensure consistent builds across environments
- **Performance**: Local caching minimizes network requests

### Key Features

| Feature | Description |
|---------|-------------|
| **Git-Native** | Works with any Git repository - no special server required |
| **Multiple Version Formats** | Tags, commits, pseudo-versions, latest |
| **Lock Files** | Reproducible builds with `easyp.lock` |
| **Local Caching** | Go modules-style cache architecture |
| **Vendoring Support** | Copy dependencies locally for offline builds |
| **YAML Configuration** | Simple, readable dependency declarations |

## Architecture

EasyP uses a two-tier caching system inspired by Go modules:

```
~/.easyp/
├── cache/
│   ├── download/              # Downloaded archives + checksums
│   │   └── github.com/
│   │       └── googleapis/
│   │           └── googleapis/
│   │               ├── v1.2.3.zip       # Archive
│   │               ├── v1.2.3.ziphash   # Checksum
│   │               └── v1.2.3.info      # Metadata
│   └── {git-hash}/            # Git bare repositories (internal)
└── mod/                       # Extracted, ready-to-use modules
    └── github.com/
        └── googleapis/
            └── googleapis/
                ├── v1.2.3/           # Tagged version
                │   ├── google/
                │   │   ├── api/
                │   │   └── rpc/
                │   └── ...
                └── v0.0.0-20250101123456-abc123def/  # Pseudo-version
                    ├── google/
                    └── ...
```

### Cache Location

| Environment | Location | How to Set |
|-------------|----------|------------|
| **Default** | `$HOME/.easyp` | Automatic |
| **Custom** | Any directory | Set `EASYPPATH` environment variable |
| **CI/CD** | Project-relative | `export EASYPPATH=$CI_PROJECT_DIR/.easyp` |

## Configuration

### Basic Configuration

Configure dependencies in your `easyp.yaml` file:

```yaml
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
  - github.com/bufbuild/protoc-gen-validate
```

### Advanced Configuration Examples

#### Multi-Environment Setup
```yaml
# development.easyp.yaml
deps:
  - github.com/googleapis/googleapis              # Latest for development
  - github.com/mycompany/internal-protos          # Latest internal changes
  - github.com/bufbuild/protoc-gen-validate       # Latest features

# production.easyp.yaml
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1       # Pinned
  - github.com/mycompany/internal-protos@v2.1.0                # Stable release
  - github.com/bufbuild/protoc-gen-validate@v0.10.1           # Tested version
```

#### Private Repository Setup
```yaml
deps:
  # Public dependencies
  - github.com/googleapis/googleapis@common-protos-1_3_1

  # Private company repositories
  - github.com/mycompany/auth-protos@v1.5.0
  - github.com/mycompany/common-types@v2.0.1

  # Internal GitLab
  - gitlab.company.com/platform/messaging-protos@v0.3.0
```

## Versioning Strategies

EasyP supports multiple versioning approaches to fit different development workflows:

### 1. Semantic Version Tags (Recommended for Production)

```yaml
deps:
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
  - github.com/googleapis/googleapis@common-protos-1_3_1
```

**Use when:**
- Production deployments
- Stable API consumption
- Reproducible builds required

### 2. Latest Tag (Development)

```yaml
deps:
  - github.com/googleapis/googleapis    # Uses latest available tag
  - github.com/bufbuild/protoc-gen-validate
```

**Use when:**
- Active development
- Want latest features
- Compatibility testing

### 3. Commit Hashes (Bleeding Edge)

```yaml
deps:
  - github.com/bufbuild/protoc-gen-validate@abc123def456789abcdef123456789abcdef1234
```

**Use when:**
- Need unreleased features
- Testing specific fixes
- Contributing to upstream

### 4. Pseudo-Versions (Automatic)

When EasyP can't find a suitable tag, it generates pseudo-versions automatically:

```
Format: v0.0.0-{timestamp}-{short-commit-hash}
Example: v0.0.0-20250908104020-660ec2d64e07f2fa8947527443af058b3d7169df
```

This ensures every commit can be referenced with a version-like identifier.

## Commands

### `easyp mod download`

Downloads and installs all dependencies declared in your configuration.

**What happens:**
1. **Resolves versions**: Converts tags/latest to specific commits
2. **Downloads archives**: Stores `.zip` files in cache/download
3. **Verifies checksums**: Ensures archive integrity
4. **Extracts modules**: Unpacks to cache/mod with proper structure
5. **Updates lock file**: Records exact versions and content hashes

**Usage:**
```bash
# Use default easyp.yaml
easyp mod download

# Use custom config file
easyp -cfg production.easyp.yaml mod download

# With custom cache location
EASYPPATH=/tmp/easyp-cache easyp mod download
```

**Example output:**
```
INFO Download package package=github.com/googleapis/googleapis version=v0.0.0-20250909114430 commit=8727b5ba
INFO Install package package=github.com/googleapis/googleapis version=v0.0.0-20250909114430 commit=8727b5ba
INFO Download package package=github.com/grpc-ecosystem/grpc-gateway version=v2.19.1 commit=a070de73
INFO Install package package=github.com/grpc-ecosystem/grpc-gateway version=v2.19.1 commit=a070de73
```

### `easyp mod vendor`

Copies all installed proto files to a local `easyp_vendor/` directory for offline usage.

**Use cases:**
- **Docker builds**: Avoid network dependencies in containers
- **Air-gapped environments**: No internet access during builds
- **Reproducible builds**: Bundle exact dependency versions
- **Performance**: Eliminate network latency in repeated builds

**Usage:**
```bash
easyp mod vendor
```

**Result structure:**
```
easyp_vendor/
├── github.com/
│   ├── googleapis/
│   │   └── googleapis/
│   │       ├── google/
│   │       │   ├── api/
│   │       │   │   ├── annotations.proto
│   │       │   │   └── http.proto
│   │       │   └── rpc/
│   │       │       └── status.proto
│   │       └── ...
│   └── grpc-ecosystem/
│       └── grpc-gateway/
│           └── protoc-gen-openapiv2/
│               └── options/
│                   └── annotations.proto
```

### `easyp mod update`

Updates module versions based on current configuration and writes resolved versions to the lock file.

**Behavior:**
- Respects version constraints in `easyp.yaml`
- Updates to latest compatible versions
- Regenerates `easyp.lock` with new versions and hashes

**Usage:**
```bash
easyp mod update
```

## Lock Files

The `easyp.lock` file ensures reproducible builds by recording exact versions and content hashes:

```
github.com/bufbuild/protoc-gen-validate v0.0.0-20250908104020-660ec2d64e07f2fa8947527443af058b3d7169df h1:ZZ5JyUkmrj9OBHM+gOCzeL5L/pAKVbsUl051yhhJTjU=
github.com/googleapis/googleapis v0.0.0-20250909114430-8727b5ba7f23fbbfddda58239e8bc6b547e05878 h1:eI+XYpPio3fxl9H5/VjW2PxlxM/7yqPjEq3oQ6jUkj4=
github.com/grpc-ecosystem/grpc-gateway v2.19.1 h1:01NNlCezvwUQ07ZvblXH0kelWq8hNl2qb44bOMcaSTQ=
```

### Lock File Format

Each line contains three components:
- **Module path**: Full repository path
- **Exact version**: Resolved version (tag or pseudo-version)
- **Content hash**: SHA256 of extracted content (`h1:` prefix)

### Best Practices

✅ **Always commit `easyp.lock`** - Ensures team consistency
✅ **Run `mod update` deliberately** - Don't auto-update in CI
✅ **Review lock changes** - Understand what's being updated
❌ **Don't edit manually** - Let EasyP manage the format

## Authentication

### Public Repositories

No setup required - works out of the box:

```yaml
deps:
  - github.com/googleapis/googleapis
  - github.com/bufbuild/protoc-gen-validate
```

### Private Repositories

#### SSH Keys (Recommended)

Configure Git to use SSH for GitHub/GitLab:

```bash
# For GitHub
git config --global url."git@github.com:".insteadOf "https://github.com/"

# For GitLab
git config --global url."git@gitlab.com:".insteadOf "https://gitlab.com/"

# For custom domains
git config --global url."git@gitlab.company.com:".insteadOf "https://gitlab.company.com/"
```

Then use normal HTTPS URLs in your config:

```yaml
deps:
  - github.com/mycompany/private-protos@v1.0.0
  - gitlab.company.com/platform/shared-types@v2.1.0
```

#### Personal Access Tokens

For HTTPS authentication:

```bash
# Method 1: Credential helper
git config --global credential.helper store
echo "https://username:token@github.com" >> ~/.git-credentials

# Method 2: URL rewriting
git config --global url."https://username:token@github.com/mycompany".insteadOf "https://github.com/mycompany"
```

#### Corporate Environments

```bash
# Configure proxy
git config --global http.proxy http://proxy.company.com:8080
git config --global https.proxy https://proxy.company.com:8080

# Configure certificates for internal Git servers
git config --global http.sslCAInfo /path/to/certificate.pem
```

## Common Workflows

### Initial Project Setup

```bash
# 1. Create configuration
cat > easyp.yaml << EOF
deps:
  - github.com/googleapis/googleapis
  - github.com/grpc-ecosystem/grpc-gateway@v2.19.1
EOF

# 2. Download dependencies
easyp mod download

# 3. Verify installation
ls ~/.easyp/mod/github.com/googleapis/googleapis/
```

### Adding New Dependencies

```bash
# 1. Edit easyp.yaml
echo "  - github.com/bufbuild/protoc-gen-validate@v0.10.1" >> easyp.yaml

# 2. Download new dependency
easyp mod download

# 3. Commit lock file changes
git add easyp.lock
git commit -m "Add protoc-gen-validate dependency"
```

### Updating Dependencies

```bash
# Update to latest compatible versions
easyp mod update

# Review changes
git diff easyp.lock

# Test with new versions
easyp generate
easyp lint

# Commit if everything works
git add easyp.lock
git commit -m "Update dependencies"
```

### Offline Development

```bash
# Vendor all dependencies
easyp mod vendor

# Now your project works offline
easyp -I easyp_vendor generate
```

## Troubleshooting

### Common Issues

#### "Repository not found" or "Authentication failed"

**Problem**: Can't access private repository
**Solution**: Check authentication setup

```bash
# Test Git access
git ls-remote https://github.com/mycompany/private-repo

# Check Git configuration
git config --list | grep url
```

#### "Version not found"

**Problem**: Specified tag/version doesn't exist
**Solution**: Check available tags

```bash
# List available tags
git ls-remote --tags https://github.com/googleapis/googleapis

# Use existing tag or commit hash
deps:
  - github.com/googleapis/googleapis@common-protos-1_3_1  # Valid tag
```

#### "Cache corruption" or "Checksum mismatch"

**Problem**: Corrupted cache files
**Solution**: Clear cache and re-download

```bash
# Clear everything
rm -rf ~/.easyp

# Or clear just downloads
rm -rf ~/.easyp/cache/download

# Re-download
easyp mod download
```

#### Network timeouts

**Problem**: Slow or unreliable network
**Solution**: Configure Git timeouts

```bash
# Increase timeout
git config --global http.lowSpeedLimit 1000
git config --global http.lowSpeedTime 300

# Use proxy if available
git config --global http.proxy http://proxy.company.com:8080
```

### Performance Optimization

#### For Large Teams

```bash
# Use shared cache server (if available)
export EASYPPATH=/shared/easyp-cache

# Or use team-specific cache
export EASYPPATH=/team-cache/easyp
```

#### For CI/CD Systems

```bash
# Use project-relative cache
export EASYPPATH=$CI_PROJECT_DIR/.easyp

# Cache between builds (GitLab CI example)
cache:
  key: easyp-$CI_COMMIT_REF_SLUG
  paths:
    - .easyp/
```

#### Cache Size Management

```bash
# Check cache usage
du -sh ~/.easyp
du -sh ~/.easyp/cache/download    # Archives only
du -sh ~/.easyp/mod               # Extracted modules

# Clean old versions (manual)
find ~/.easyp/mod -type d -name "v0.0.0-*" -mtime +30 -exec rm -rf {} \;
```

## Integration Examples

### Docker Multi-stage Build

```dockerfile
# Stage 1: Download dependencies
FROM ghcr.io/easyp-tech/easyp:latest AS deps
WORKDIR /workspace
COPY easyp.yaml easyp.lock ./
RUN easyp mod vendor

# Stage 2: Build application
FROM alpine:latest AS build
WORKDIR /app
COPY --from=deps /workspace/easyp_vendor ./easyp_vendor
COPY . .
# Use vendored dependencies for generation
RUN easyp -I easyp_vendor generate
```

### Monorepo Structure

```
my-monorepo/
├── services/
│   ├── auth-service/
│   │   └── easyp.yaml          # Service-specific deps
│   └── user-service/
│       └── easyp.yaml          # Different deps
├── shared/
│   └── common-protos/          # Internal protos
└── easyp.yaml                  # Global/shared deps
```

Each `easyp.yaml` can have different dependencies based on service needs.

## Best Practices

### Development Workflow
- ✅ Use **latest tags** during active development
- ✅ **Pin versions** for production deployments
- ✅ **Commit lock files** to ensure reproducibility
- ✅ **Review dependency updates** before merging
- ✅ **Test after updates** to catch compatibility issues

### Security
- ✅ **Pin to specific versions** in production
- ✅ **Use SSH keys** for private repositories
- ✅ **Review new dependencies** for security implications
- ✅ **Monitor for vulnerabilities** in dependencies
- ❌ **Don't embed credentials** in configuration files

### Performance
- ✅ **Cache aggressively** in CI/CD systems
- ✅ **Use vendoring** for frequently rebuilt projects
- ✅ **Clean old cache** periodically to save space
- ✅ **Use shared cache** for team environments

### Team Collaboration
- ✅ **Document authentication setup** for new team members
- ✅ **Use consistent tooling** across environments
- ✅ **Automate dependency updates** with proper testing
- ✅ **Share cache locations** when possible

The EasyP package manager provides a robust, decentralized solution for protobuf dependency management that scales from individual projects to enterprise environments while maintaining the simplicity and reliability developers expect.
