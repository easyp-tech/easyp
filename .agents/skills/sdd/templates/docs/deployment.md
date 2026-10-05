<!-- scope: Dockerfile*, docker-compose*, **/deploy/*, **/k8s/* -->
# Deployment Documentation Template

## What This Generates

- `.spec/DEPLOYMENT.md` — environments, deployment pipeline, rollout strategy, health checks, and infrastructure

## Instructions

You are a technical documentarian. Create deployment documentation for the project in the `.spec/` directory.
Analyze: Dockerfile, docker-compose files, CI/CD configs, Kubernetes manifests, Helm charts, Terraform files, deployment scripts, health check endpoints.

### Step 1: Identify Deployment Infrastructure

Check for the presence of:
- **Containers**: `Dockerfile`, `docker-compose.yml`, `.dockerignore`
- **CI/CD**: `.github/workflows/`, `.gitlab-ci.yml`, `Jenkinsfile`, `bitbucket-pipelines.yml`
- **Orchestration**: `k8s/`, `helm/`, `kustomize/`, `nomad/`
- **IaC**: `terraform/`, `pulumi/`, `cloudformation/`
- **PaaS**: `fly.toml`, `render.yaml`, `railway.json`, `vercel.json`, `netlify.toml`

### Step 2: Create .spec/DEPLOYMENT.md

#### Structure:

##### 1. Overview
- One sentence: deployment target + strategy
- ASCII diagram of the deployment pipeline:
```
push → CI (lint + test) → build image → push registry → deploy → health check
```

##### 2. Environments

| Environment | URL / Host | Purpose | Branch | Auto-deploy |
|-------------|-----------|---------|--------|-------------|
| Local | `localhost:8080` | Development | — | — |
| Staging | `staging.example.com` | Pre-production testing | `develop` | Yes |
| Production | `example.com` | Live | `main` | No (manual) |

For each environment: how to access, what branch triggers deployment, any special config.

##### 3. Docker

**Dockerfile analysis:**
- Base image and build stages (multi-stage?)
- Build arguments and their purpose
- Exposed ports
- Entry point / command
- Image size optimization notes

**Docker Compose (dev):**
- Services defined and their roles
- Volume mounts
- Network configuration
- Useful commands:
```bash
docker-compose up -d        # Start all services
docker-compose logs -f app  # Follow app logs
docker-compose down -v      # Stop and remove volumes
```

##### 4. CI/CD Pipeline

For each CI/CD workflow file:
- **File**: path to workflow
- **Trigger**: on push, PR, tag, manual
- **Steps**: ordered list of what happens
- **Secrets required**: list of secret names (not values)

Example:
```
.github/workflows/ci.yml
  Trigger: push to main, PR
  Steps: checkout → setup → lint → test → build → push image
  Secrets: DOCKER_USERNAME, DOCKER_PASSWORD, DEPLOY_KEY
```

##### 5. Rollout Strategy
- Strategy type: rolling update / blue-green / canary / recreate
- Configuration (max surge, max unavailable, canary percentage)
- Where configured (k8s deployment spec, PaaS config, script)
- Zero-downtime deployment: yes/no, how achieved

##### 6. Health Checks

| Endpoint | Type | Expected Response | Timeout |
|----------|------|-------------------|---------|
| `/health` | Liveness | `200 OK` | 5s |
| `/ready` | Readiness | `200 OK` | 10s |

- What each check verifies (DB connection, cache, external services)
- Startup probe (if applicable)
- Code reference: where health check handlers are defined

##### 7. Rollback Procedure
Step-by-step rollback instructions:
1. Identify the issue (logs, metrics, alerts)
2. Rollback command (e.g., `kubectl rollout undo`, `git revert` + redeploy, PaaS rollback)
3. Verify rollback (health check, smoke test)
4. Post-mortem actions

Include actual commands for the project's deployment target.

##### 8. Secrets Management
- Where secrets are stored: environment variables, vault, cloud secret manager, CI/CD variables
- How secrets are injected: env files, k8s secrets, mounted volumes
- Rotation policy (if documented)
- Required secrets list:

| Secret | Purpose | Where Set |
|--------|---------|-----------|
| `DATABASE_URL` | DB connection | CI/CD vars |
| `JWT_SECRET` | Token signing | Vault |

Do not include actual secret values.

##### 9. Infrastructure Requirements (if applicable)
Resource requirements per service:
| Service | CPU | Memory | Storage | Replicas |
|---------|-----|--------|---------|----------|

Determine from k8s resource limits, PaaS config, or documentation.

##### 10. Monitoring & Alerts (if applicable)
- Where to view logs (log aggregator URL, command)
- Metrics dashboard (Grafana URL, Datadog, etc.)
- Alert rules (what triggers alerts)
- On-call / escalation process (if documented)

## General Rules

- Language: English
- All commands and paths must come from actual project files
- Never include real secrets, tokens, or internal URLs — use placeholders
- If the project has no deployment config (pure library), do not generate this file
- Focus on what exists, not what should exist — document current state
- After creating, update `.spec/README.md`: add a link under the appropriate section
