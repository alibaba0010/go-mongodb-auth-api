# Gin Mongo AWS Project

This is a production-ready Go project using Gin, MongoDB, and Redis, designed to be deployed on AWS.

## Features

- **Web Framework**: Gin
- **Database**: MongoDB (using official Go driver)
- **Caching**: Redis
- **Configuration**: Viper (supports env vars and config files)
- **Logging**: Zap (structured logging)
- **Rate Limiting**: Redis-backed rate limiting
- **Graceful Shutdown**: Handles SIGINT and SIGTERM
- **Docker**: Dockerfile and docker-compose for local development

## Project Structure

- `cmd/api`: Main entry point
- `internal/config`: Configuration loading
- `internal/database`: Database connections (Mongo, Redis)
- `internal/handlers`: HTTP handlers
- `internal/middleware`: HTTP middleware (Logger, RateLimit, CORS)
- `internal/models`: Data models
- `internal/repository`: Data access layer
- `internal/service`: Business logic layer
- `internal/server`: Server setup and wiring
- `pkg/logger`: Logger setup

## Local Development

1.  **Prerequisites**: Go 1.23+, Docker, Docker Compose.
2.  **Run with Docker Compose**:
    ```bash
    docker-compose up --build
    ```
    The API will be available at `http://localhost:8080`.

3.  **Run Manually**:
    - Ensure MongoDB and Redis are running.
    - Update `config.yaml` or set environment variables.
    - Run:
      ```bash
      go run cmd/api/main.go
      ```

## AWS Deployment

### Option 1: AWS App Runner (Recommended for simplicity)

1.  Push your code to a GitHub repository.
2.  Go to AWS App Runner console.
3.  Create a service linked to your GitHub repo.
4.  Configure the build:
    - Runtime: Go 1.x
    - Build command: `go build -o main ./cmd/api`
    - Start command: `./main`
    - Port: 8080
5.  Set environment variables in the App Runner configuration:
    - `MONGODB_URI`: Connection string to your MongoDB Atlas or DocumentDB.
    - `REDIS_ADDR`: Address of your ElastiCache Redis.
    - `SERVER_MODE`: `release`

### Option 2: AWS ECS (Fargate)

1.  Build the Docker image and push to Amazon ECR.
    ```bash
    aws ecr create-repository --repository-name gin-mongo-aws
    docker build -t gin-mongo-aws .
    docker tag gin-mongo-aws:latest <aws_account_id>.dkr.ecr.<region>.amazonaws.com/gin-mongo-aws:latest
    aws ecr get-login-password --region <region> | docker login --username AWS --password-stdin <aws_account_id>.dkr.ecr.<region>.amazonaws.com
    docker push <aws_account_id>.dkr.ecr.<region>.amazonaws.com/gin-mongo-aws:latest
    ```
2.  Create an ECS Cluster.
3.  Create a Task Definition using the ECR image.
    - Add environment variables for MongoDB and Redis.
4.  Create a Service using the Task Definition.

## API Endpoints

- `POST /api/v1/users`: Create a user
- `GET /api/v1/users`: Get all users
- `GET /api/v1/users/:id`: Get a user by ID
- `PUT /api/v1/users/:id`: Update a user
- `DELETE /api/v1/users/:id`: Delete a user
- `GET /health`: Health check
