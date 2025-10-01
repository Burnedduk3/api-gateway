# API Gateway - Technical Challenge

## Table of Contents
- [Overview](#overview)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [AWS Configuration](#aws-configuration)
- [Quick Start](#quick-start)
- [Token Generation](#token-generation)
- [Running the Application](#running-the-application)
- [Configuration](#configuration)
- [Hot Reload Feature](#hot-reload-feature)
- [API Documentation](#api-documentation)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)
- [Project Structure](#project-structure)

## Overview

This is an API Gateway implementation that routes requests to multiple microservices (User Service, Orders Service, and Product Service). The gateway provides:

- Dynamic routing and request forwarding
- Authentication and authorization with API tokens (stored in Redis)
- Rate limiting
- CORS handling
- Hot configuration reload (no restart required)
- Health checks and metrics endpoints

### Tech Stack
- **Language**: Go 1.25
- **Framework**: Echo (HTTP framework)
- **Database**: PostgreSQL 17.6
- **Cache**: Redis 8.2.1 (for API token storage)
- **Container**: Docker & Docker Compose
- **Configuration**: Viper (with hot reload)

## Architecture

# FALTA

### Services

| Service | Port | Description |
|---------|------|-------------|
| API Gateway | 8300 | Main entry point, routes requests, validates tokens |
| User Service | 8000 | User management |
| Orders Service | 8100 | Order processing |
| Product Service | 8200 | Product catalog |
| PostgreSQL | 5432 | Primary database |
| Redis | 6379 | API token storage |

### Authentication Flow

The API Gateway uses Redis to store and validate API tokens:

1. On first health check, the gateway generates 2 tokens:
    - Token 1: Valid token (stored in Redis)
    - Token 2: Invalid token (not stored in Redis)
2. Protected endpoints validate tokens against Redis
3. Invalid tokens or missing tokens return 401

## Prerequisites

Before running this application, ensure you have the following installed:

### Required Software
- **Docker**: Version 20.10 or higher
  ```bash
  docker --version
  ```
- **Docker Compose**: Version 2.0 or higher
  ```bash
  docker-compose --version
  ```
- **AWS CLI**: Version 2.x
  ```bash
  aws --version
  ```
- **Git**: For cloning the repository
  ```bash
  git clone <repository-url>
  cd api-gateway
  ```

### Optional (for local development)
- **Go**: Version 1.25 or higher
- **curl** or **Postman**: For API testing

## AWS Configuration

The application uses Docker images hosted in AWS ECR (Elastic Container Registry). You need to configure AWS CLI and authenticate to pull these images.

### Step 1: Install AWS CLI

**macOS:**
```bash
brew install awscli
```

**Linux:**
```bash
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
```

**Windows:**
Download and run the installer from: https://aws.amazon.com/cli/

### Step 2: Configure AWS Credentials

You need AWS credentials with ECR read permissions. Contact the project administrator for credentials.

```bash
# Configure AWS CLI
aws configure

# You'll be prompted for:
AWS Access Key ID [None]: <YOUR_ACCESS_KEY>
AWS Secret Access Key [None]: <YOUR_SECRET_KEY>
Default region name [None]: us-east-1
Default output format [None]: json
```

### Step 3: Authenticate Docker to ECR

```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 049139783164.dkr.ecr.us-east-1.amazonaws.com
```

You should see:
```
Login Succeeded
```

**Note**: This authentication expires after 12 hours. If you encounter image pull errors later, re-run this command.

## Quick Start

### Option 1: Local Development (Without API Gateway Container)

This runs backend services in Docker, allowing you to run the API Gateway locally for development:

```bash
# Start backend services
docker-compose -f docker/docker-compose-localdev.yaml up -d

# Wait for services to be healthy (30-60 seconds)
docker-compose -f docker/docker-compose-localdev.yaml ps

# Run API Gateway locally (in another terminal)
go run main.go server

# Or build and run
go build -o api-gateway .
./api-gateway server
```

### Option 2: Full Docker Environment (Recommended for Testing)

Run everything in Docker, including the API Gateway:

```bash
# Using locally built image
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml up -d

# OR using ECR image
docker-compose -f docker/docker-compose-test-container.yaml up -d
```

### Verify Services are Running

```bash
# Check all services are healthy
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml ps

# All services should show "Up" or "Up (healthy)"
```

## Token Generation

**IMPORTANT**: The API Gateway generates authentication tokens on the first health check request.

### Step 1: Hit the Health Check Endpoint

After starting all services, make a request to the health check endpoint:

```bash
# This triggers token generation
curl http://localhost:8300/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

### Step 2: Generated Tokens

After hitting the health check, the gateway automatically generates 2 test tokens:

- **Valid Token**: `key-123` (stored in Redis with value `true`)
- **Invalid Token**: `key-1234` (stored in Redis with value `false`)

### Step 3: Test Token Authentication

```bash
# Test with VALID token (should work)
curl -H "X-API-Key: key-123" http://localhost:8300/api/v1/users

# Test with INVALID token (should fail with 401/403)
curl -H "X-API-Key: key-1234" http://localhost:8300/api/v1/users

# Test without token (should fail with 401/403)
curl http://localhost:8300/api/v1/users
```

### How Token Storage Works

1. **Token Generation**: On first health check, the gateway generates 2 tokens in Redis:
   ```
   key-123  = true   (valid)
   key-1234 = false  (invalid)
   ```
2. **Token Validation**: For each protected request:
    - Gateway extracts token from `X-API-Key` header
    - Queries Redis to get the token value
    - Allows request if value is `true`, rejects if `false` or not found
3. **Token Lifecycle**: Tokens persist in Redis until manually removed or Redis is flushed

### Manual Token Management (Optional)

```bash
# Connect to Redis container
docker exec -it redis redis-cli

# Check valid token
GET key-123
# Should return: "true"

# Check invalid token
GET key-1234
# Should return: "false"

# List all keys
KEYS *

# Remove a token
DEL key-123

# Flush all tokens
FLUSHDB
```

## Running the Application

### Starting Services

```bash
# Start all services in detached mode
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml up -d

# View logs
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml logs -f

# View specific service logs
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml logs -f api-gateway
```

### Stopping Services

```bash
# Stop all services
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml down

# Stop and remove volumes (clean slate - will regenerate tokens)
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml down -v
```

### Rebuilding Services

```bash
# Rebuild specific service
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml build api-gateway

# Rebuild and restart
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml up -d --build api-gateway
```

## Configuration

### Configuration Files

Configuration files are located in `configs/`:

```
configs/
├── config.yaml                 # API Gateway config
├── api-gateway-config.yaml    # Alternative API Gateway config
├── user-config.yaml           # User service config
├── orders-config.yaml         # Orders service config
└── product-config.yaml        # Product service config
```

### Main Configuration (config.yaml)

The API Gateway is configured via `configs/config.yaml`:

```yaml
environment: development
version: 1.0.0

server:
  port: "8300"
  host: "0.0.0.0"
  read_timeout: "30s"
  write_timeout: "30s"
  path_prefix: "/api"

redis:
  host: "localhost"
  port: "6379"
  password: ""
  database: 0

backends:
  - host: "http://localhost:8000"
    id: "user"
    path_prefix: "/api/v1"
    routes:
      - id: "user-list"
        method: "GET"
        path: "/users"
        path_type: "exact"
        enabled: "true"
        auth_policy:
          type: "api"
          enabled: "true"
```

### Environment Variables

Override configuration using environment variables:

```bash
# Set environment
export PRODUCT_SERVICE_ENVIRONMENT=production

# Override server port
export PRODUCT_SERVICE_SERVER_PORT=8301

# Override Redis connection
export PRODUCT_SERVICE_REDIS_HOST=redis-prod
export PRODUCT_SERVICE_REDIS_PORT=6379
```

## Hot Reload Feature

The API Gateway supports hot configuration reload without restart! 🔥

### How to Use

1. Start the API Gateway
2. Edit `config.yaml` to add/modify/remove routes or backends
3. Save the file
4. Changes are applied automatically within ~100ms

### What Can Be Hot Reloaded

✅ **Supported (No Restart):**
- Add/remove/modify routes
- Enable/disable routes
- Add/remove backends
- Update rate limiting
- Change CORS settings
- Modify auth policies

❌ **Requires Restart:**
- Server port changes
- Server host changes
- Redis connection settings

### Example

```bash
# Start the gateway
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml up -d

# Generate tokens first
curl http://localhost:8300/health

# In another terminal, edit config
vim configs/config.yaml

# Add a new route or disable an existing one:
# enabled: "false"

# Save and watch the logs:
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml logs -f api-gateway

# You'll see:
# INFO Config file changed, reloading... file=config.yaml
# INFO Configuration reloaded successfully
```

For more details, see [docs/HOT_RELOAD.md](docs/HOT_RELOAD.md)

## API Documentation

### Base URL

```
http://localhost:8300/api
```

### Authentication

Most endpoints require an API token. Include it in the `X-API-Key` header:

```bash
curl -H "X-API-Key: key-123" http://localhost:8300/api/v1/users
```

**Remember**: Tokens are automatically generated on first health check:
- Valid token: `key-123`
- Invalid token: `key-1234`

### Health Endpoints (No Authentication Required)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Gateway health check (triggers token generation) |
| `/api/v1/health` | GET | Service health check |
| `/api/v1/health/ready` | GET | Service readiness check |
| `/api/v1/health/live` | GET | Service liveness check |
| `/metrics` | GET | Prometheus metrics (if enabled) |

### User Service Endpoints

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/api/v1/users` | GET | ✅ Yes | List all users |
| `/api/v1/users` | POST | ✅ Yes | Create user |
| `/api/v1/users/:id` | GET | ✅ Yes | Get user by ID |
| `/api/v1/users/email/:email` | GET | ✅ Yes | Get user by email |

### Orders Service Endpoints

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/api/v1/orders` | GET | ✅ Yes | List orders |
| `/api/v1/orders` | POST | ✅ Yes | Create order |
| `/api/v1/orders/:id` | GET | ✅ Yes | Get order by ID |
| `/api/v1/orders/:id/confirm` | POST | ✅ Yes | Confirm order |
| `/api/v1/orders/:id/cancel` | POST | ✅ Yes | Cancel order |
| `/api/v1/orders/:id/items` | POST | ✅ Yes | Add item to order |
| `/api/v1/orders/:id/items/:product_id` | PUT | ✅ Yes | Update order item |
| `/api/v1/orders/:id/items/:product_id` | DELETE | ✅ Yes | Remove order item |

### Product Service Endpoints

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/api/v1/products` | GET | ❌ No | List products |
| `/api/v1/products` | POST | ✅ Yes | Create product |
| `/api/v1/products/:id` | GET | ❌ No | Get product by ID |
| `/api/v1/products/sku/:sku` | GET | ❌ No | Get product by SKU |
| `/api/v1/products/:id/stock` | PATCH | ✅ Yes | Update product stock |
| `/api/v1/products/:id/price` | PATCH | ✅ Yes | Update product price |

### Example Requests

**1. Generate Tokens (Required First Step):**
```bash
curl http://localhost:8300/health
# Check logs for generated tokens
```

**2. Create a User (Requires Token):**
```bash
curl -X POST http://localhost:8300/api/v1/users \
  -H "Content-Type: application/json" \
  -H "X-API-Key: <VALID_TOKEN>" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe"
  }'
```

**3. List Products (No Token Required):**
```bash
curl http://localhost:8300/api/v1/products
```

**4. Get Order by ID (Requires Token):**
```bash
curl -H "X-API-Key: <VALID_TOKEN>" \
  http://localhost:8300/api/v1/orders/123
```

**5. Test Invalid Token:**
```bash
curl -H "X-API-Key: <INVALID_TOKEN>" \
  http://localhost:8300/api/v1/users
# Expected: 401 Unauthorized or 403 Forbidden
```

## Testing

### Running Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/domain/entities/...
```

### Running Integration Tests

```bash
# Start test environment
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml up -d

# Wait for services to be ready
sleep 30

# Generate tokens
curl http://localhost:8300/health

# Run integration tests
go test -tags=integration ./tests/integration/...

# Cleanup
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml down -v
```

### Manual Testing with curl

```bash
# 1. Start services and generate tokens
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml up -d
sleep 30
curl http://localhost:8300/health

# 2. Test authenticated endpoint with valid token
curl -H "X-API-Key: key-123" http://localhost:8300/api/v1/users

# 3. Test authenticated endpoint with invalid token
curl -H "X-API-Key: key-1234" http://localhost:8300/api/v1/users
# Expected: 401/403

# 4. Test public endpoint (no token needed)
curl http://localhost:8300/api/v1/products

# 5. Test rate limiting (send 150 rapid requests)
for i in {1..150}; do 
  curl -s http://localhost:8300/api/v1/products > /dev/null
  echo "Request $i"
done
```

## Troubleshooting

### Common Issues

#### 1. ECR Authentication Expired

**Error:**
```
Error response from daemon: pull access denied for 049139783164.dkr.ecr.us-east-1.amazonaws.com/develop/user-service
```

**Solution:**
```bash
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 049139783164.dkr.ecr.us-east-1.amazonaws.com
```

#### 2. Tokens Not Generated

**Symptom:** Authentication always fails

**Solution:**
```bash
# Make sure you hit health check first
curl http://localhost:8300/health

# Verify Redis is running
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml ps redis

# Check Redis connection
docker exec -it redis redis-cli PING
# Should return: PONG

# Verify tokens exist in Redis
docker exec -it redis redis-cli
GET key-123
# Should return: "true"
GET key-1234
# Should return: "false"
```

#### 3. Port Already in Use

**Error:**
```
Error starting userland proxy: listen tcp 0.0.0.0:8300: bind: address already in use
```

**Solution:**
```bash
# Find and kill the process using the port
lsof -ti:8300 | xargs kill -9

# Or use a different port by editing docker-compose file
```

#### 4. Redis Connection Refused

**Error:**
```
Failed to connect to Redis: connection refused
```

**Solution:**
```bash
# Check Redis is running
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml ps redis

# Check Redis logs
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml logs redis

# Restart Redis
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml restart redis

# Regenerate tokens after Redis restart
curl http://localhost:8300/health
```

#### 5. Authentication Fails with Valid Token

**Check:**
```bash
# Verify token exists in Redis
docker exec -it redis redis-cli

# Inside Redis CLI:
GET key-123
# Should return: "true"

GET key-1234
# Should return: "false"
```

**Solution:**
```bash
# Regenerate tokens by hitting health check
curl http://localhost:8300/health

# Or restart with fresh Redis
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml down -v
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml up -d
curl http://localhost:8300/health
```

#### 6. Services Not Healthy

**Check logs:**
```bash
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml logs postgres
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml logs user-service
```

**Solution:**
```bash
# Restart specific service
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml restart user-service

# Or restart all services
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml restart
```

#### 7. Configuration Not Reloading

**Check:**
- File permissions on config.yaml
- File system type (some network file systems don't support fsnotify)
- Logs for validation errors

**Solution:**
```bash
# Check file permissions
ls -la configs/config.yaml

# Watch logs for errors
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml logs -f api-gateway | grep -i reload
```

### Debug Mode

Enable debug logging:

```yaml
# In config.yaml
logging:
  level: "debug"
  format: "text"
```

### Viewing Logs

```bash
# All services
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml logs -f

# Specific service
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml logs -f api-gateway

# Last 100 lines
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml logs --tail=100 api-gateway

# Filter for tokens
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml logs api-gateway | grep -i token
```

### Clean Start

If experiencing persistent issues:

```bash
# Stop everything
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml down -v

# Remove all containers, networks, and volumes
docker system prune -a --volumes

# Re-authenticate to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 049139783164.dkr.ecr.us-east-1.amazonaws.com

# Start fresh
docker-compose -f docker/docker-compose-test-container-with-devdocker.yaml up -d

# Wait and generate tokens
sleep 30
curl http://localhost:8300/health
```

## Project Structure

```
.
├── cmd/
│   ├── root.go              # Root command
│   └── server.go            # Server command
├── configs/
│   ├── config.yaml          # Main config
│   └── *-config.yaml        # Service configs
├── docker/
│   ├── dev.Dockerfile       # Development Dockerfile
│   ├── Dockerfile           # Production Dockerfile
│   └── docker-compose-*.yaml
├── docs/
│   └── HOT_RELOAD.md        # Hot reload documentation
├── internal/
│   ├── adapters/
│   │   └── http/            # HTTP handlers
│   ├── config/
│   │   ├── config.go        # Config structs
│   │   └── watcher.go       # Hot reload watcher
│   ├── domain/
│   │   └── entities/        # Domain entities
│   ├── hotreload/
│   │   └── manager.go       # Hot reload manager
│   ├── infrastructure/      # Infrastructure layer
│   └── auth/
│       └── token.go         # Token validation (Redis)
├── pkg/
│   └── logger/              # Logger package
├── tests/
│   ├── integration/         # Integration tests
│   └── unit/               # Unit tests
├── go.mod
├── go.sum
├── main.go                  # Application entry point
└── README.md               # This file
```

## Additional Resources

- [Hot Reload Documentation](docs/HOT_RELOAD.md)
- [Contributing Guidelines](CONTRIBUTING.md)
- [License](LICENSE)

## Support

For questions or issues:
1. Check the [Troubleshooting](#troubleshooting) section
2. Review logs using the commands above
3. Contact: juandavid.juandis@gmail.com

## License

[Your License Here]

---

**Version**: 1.0.0  
**Last Updated**: January 2025  
**Author**: Juan David Cabrera Duran