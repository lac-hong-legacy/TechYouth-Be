# VEN API - TechYouth Backend

[![CI](https://github.com/lac-hong-legacy/TechYouth-Be/actions/workflows/ci.yml/badge.svg)](https://github.com/lac-hong-legacy/TechYouth-Be/actions/workflows/ci.yml)
[![Deploy Staging](https://github.com/lac-hong-legacy/TechYouth-Be/actions/workflows/deploy-staging.yml/badge.svg)](https://github.com/lac-hong-legacy/TechYouth-Be/actions/workflows/deploy-staging.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Backend API for TechYouth learning platform built with Go, Fiber, PostgreSQL, and Redis.

## 🚀 Features

- **User Authentication**: JWT-based authentication with refresh tokens, email verification, and password reset
- **Content Management**: Characters, lessons, timelines, and educational content
- **Media Handling**: MinIO integration for video, image, and subtitle storage
- **Gamification**: XP system, achievements, spirits, and leaderboards
- **Guest Mode**: Try before you register functionality
- **Rate Limiting**: Redis-based rate limiting for API protection
- **Email Service**: MailHog integration for email notifications
- **API Documentation**: Swagger/OpenAPI documentation
- **CI/CD**: Automated testing and deployment with GitHub Actions

## 📋 Prerequisites

- Docker & Docker Compose
- Go 1.24+ (for local development)
- GitHub account (for private module access)
- Make (optional, for convenience commands)

## 🛠️ Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/lac-hong-legacy/TechYouth-Be.git
cd TechYouth-Be
```

### 2. Setup Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env and add your GitHub token
vim .env
```

Get a GitHub token from: https://github.com/settings/tokens/new

### 3. Start Services

```bash
# Using Make
make install

# Or using Docker Compose directly
docker compose up --build
```

### 4. Access Services

- **API**: http://localhost:8000
- **API Documentation**: http://localhost:8000/swagger/
- **MailHog UI**: http://localhost:8025
- **MinIO Console**: http://localhost:9001

## 📚 Documentation

- [CI/CD Documentation](CICD.md) - GitHub Actions workflows and deployment
- [Docker Setup](DOCKER_SETUP.md) - Docker and Docker Compose guide
- [API Documentation](http://localhost:8000/swagger/) - Swagger UI (when running)

## 🏗️ Architecture

```
┌─────────────────┐
│   Fiber API     │  ← Go Backend (Port 8000)
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
┌───▼───┐ ┌──▼────┐
│ Redis │ │MinIO  │  ← Cache & Storage
└───┬───┘ └───────┘
    │
┌───▼──────┐
│PostgreSQL│  ← Database
└──────────┘
```

## 🧪 Development

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

### Database Migrations

Migrations are handled automatically by GORM AutoMigrate on startup.

### Make Commands

```bash
make help            # Show available commands
make setup           # Initial setup
make build           # Build Docker images
make up              # Start services (foreground)
make up-d            # Start services (background)
make down            # Stop services
make logs            # View logs
make logs-api        # View API logs only
make restart         # Restart all services
make restart-api     # Restart API only
make clean           # Stop and remove volumes
make rebuild         # Rebuild from scratch
make shell-api       # Open shell in API container
make db-shell        # Open PostgreSQL shell
make redis-cli       # Open Redis CLI
```

## 🚢 Deployment

### GitHub Actions CI/CD

This project uses GitHub Actions for automated testing and deployment:

- **CI**: Runs on every push and PR
  - Tests with PostgreSQL and Redis
  - Linting and code quality checks
  - Security scanning
  - Docker build and push

- **CD**: Automated deployments
  - Staging: Automatic on push to `main`
  - Production: Manual approval or version tags

See [CICD.md](CICD.md) for detailed setup instructions.

### Manual Deployment

```bash
# Build for production
docker compose -f docker-compose.prod.yml build

# Deploy
docker compose -f docker-compose.prod.yml up -d

# View logs
docker compose -f docker-compose.prod.yml logs -f
```

## 📊 Monitoring

### Health Check

```bash
curl http://localhost:8000/health
```

### Logs

```bash
# View all logs
docker compose logs -f

# View specific service
docker compose logs -f ven-api
docker compose logs -f postgres
docker compose logs -f redis
```

### Database Access

```bash
# Connect to PostgreSQL
make db-shell

# Or directly
docker compose exec postgres psql -U ven_user -d ven_api
```

### Redis Access

```bash
# Connect to Redis
make redis-cli

# Or directly
docker compose exec redis redis-cli -a ven-redis-pass
```

## 🔒 Security

- JWT-based authentication with access and refresh tokens
- Password hashing with bcrypt
- Rate limiting on sensitive endpoints
- Email verification required
- Password reset with secure codes
- Session management with device tracking
- SQL injection protection via GORM
- CORS configuration
- Security headers

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Commit Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add new feature
fix: bug fix
docs: documentation update
style: code style change
refactor: code refactoring
test: add tests
chore: maintenance tasks
```

## 📝 API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh token
- `POST /api/v1/auth/logout` - Logout
- `POST /api/v1/auth/verify-email` - Verify email
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password

### Content
- `GET /api/v1/characters` - List characters
- `GET /api/v1/characters/:id` - Get character
- `GET /api/v1/lessons` - List lessons
- `GET /api/v1/lessons/:id` - Get lesson
- `GET /api/v1/timeline` - Get timeline

### User
- `GET /api/v1/user/profile` - Get user profile
- `PUT /api/v1/user/profile` - Update profile
- `GET /api/v1/user/progress` - Get progress
- `GET /api/v1/user/achievements` - Get achievements

See [Swagger Documentation](http://localhost:8000/swagger/) for complete API reference.

## 🐛 Troubleshooting

### Database Connection Error

```bash
# Restart PostgreSQL
docker compose restart postgres

# Check logs
docker compose logs postgres
```

### Redis Connection Error

```bash
# Restart Redis
docker compose restart redis

# Test connection
docker compose exec redis redis-cli -a ven-redis-pass ping
```

### Build Errors

```bash
# Clean rebuild
make clean
make rebuild

# Or
docker compose down -v
docker compose build --no-cache
docker compose up
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👥 Authors

- **lac-hong-legacy** - *Initial work*

## 🙏 Acknowledgments

- [Fiber](https://gofiber.io/) - Web framework
- [GORM](https://gorm.io/) - ORM library
- [PostgreSQL](https://www.postgresql.org/) - Database
- [Redis](https://redis.io/) - Cache
- [MinIO](https://min.io/) - Object storage
- [MailHog](https://github.com/mailhog/MailHog) - Email testing

## 📞 Support

For issues, questions, or contributions, please:
1. Check existing [Issues](https://github.com/lac-hong-legacy/TechYouth-Be/issues)
2. Create a new issue with detailed information
3. Join our community discussions

---

Made with ❤️ by the TechYouth Team

