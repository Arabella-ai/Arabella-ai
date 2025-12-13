# Arabella - AI Video Generation Platform

<div align="center">
  <h3>🎬 Create professional AI-generated videos using pre-built templates</h3>
  <p>A scalable, high-performance backend built with Go, designed for 1M+ concurrent users</p>
</div>

---

## 🌟 Features

- **AI Video Generation** - Generate professional videos using Gemini VEO, OpenAI Sora, and more
- **Template Library** - Pre-built templates for various use cases (Cyberpunk, Product Showcase, Vlogs, etc.)
- **Real-time Updates** - WebSocket-based progress tracking during video generation
- **User Management** - Google OAuth authentication with tiered subscription system
- **Credit System** - Flexible credit-based usage with free and premium tiers
- **Scalable Architecture** - Clean Architecture design supporting 1M+ concurrent users

## 🏗️ Architecture

This project follows **Clean Architecture** principles with four distinct layers:

```
├── cmd/                    # Application entry points
│   └── api/                # Main API server
├── internal/               # Private application code
│   ├── domain/             # Entities & business rules
│   │   ├── entity/         # Core domain entities
│   │   ├── repository/     # Repository interfaces
│   │   └── service/        # Domain service interfaces
│   ├── usecase/            # Application use cases
│   ├── interface/          # HTTP handlers, middleware
│   │   ├── http/           # REST API handlers
│   │   └── websocket/      # WebSocket handlers
│   └── infrastructure/     # External systems
│       ├── auth/           # JWT & Google OAuth
│       ├── cache/          # Redis caching
│       ├── database/       # PostgreSQL
│       ├── provider/       # AI providers
│       ├── queue/          # Job queue
│       └── repository/     # Repository implementations
├── config/                 # Configuration
├── docs/                   # Swagger documentation
├── migrations/             # Database migrations
└── scripts/                # Build & deployment scripts
```

## 🚀 Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 16+ (or use Docker)
- Redis 7+ (or use Docker)

### Development Setup

1. **Clone the repository**
```bash
git clone https://github.com/arabella/ai-studio-backend.git
cd ai-studio-backend
```

2. **Install dependencies**
```bash
make deps
```

3. **Start infrastructure services**
```bash
make docker-up
```

4. **Run database migrations**
```bash
make migrate-up
```

5. **Start the development server**
```bash
make dev
```

The API will be available at `http://localhost:8080`

### Using Docker Compose (Full Stack)

```bash
docker-compose up -d
```

This starts:
- API Server (port 8080)
- PostgreSQL (port 5432)
- Redis (port 6379)

For development tools (Redis Commander, pgAdmin):
```bash
make docker-up-dev
```

## 📚 API Documentation

### Swagger UI

When running in development mode, access Swagger UI at:
```
http://localhost:8080/swagger/index.html
```

### Key Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/google` | Authenticate with Google OAuth |
| GET | `/api/v1/templates` | List all templates |
| GET | `/api/v1/templates/:id` | Get template details |
| POST | `/api/v1/videos/generate` | Start video generation |
| GET | `/api/v1/videos/:id/status` | Get generation status |
| GET | `/api/v1/user/profile` | Get user profile |

### WebSocket

Real-time progress updates:
```
ws://localhost:8080/api/v1/ws/videos/:id?token=<access_token>
```

Events:
- `progress` - Generation progress updates
- `completed` - Video generation completed
- `error` - Generation error

## ⚙️ Configuration

### Environment Variables

Copy the example and configure:
```bash
cp .env.example .env
```

Key variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | Environment (development/production) | development |
| `SERVER_PORT` | HTTP server port | 8080 |
| `DB_HOST` | PostgreSQL host | localhost |
| `REDIS_HOST` | Redis host | localhost |
| `JWT_SECRET` | JWT signing secret | (required) |
| `GEMINI_API_KEY` | Gemini VEO API key | (optional) |
| `USE_MOCK_PROVIDER` | Use mock AI provider | true |

## 🧪 Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run short tests only
make test-short
```

## 🔧 Development

### Available Make Commands

```bash
make help          # Show all commands
make build         # Build the application
make run           # Run the application
make dev           # Run with hot-reload
make test          # Run tests
make lint          # Run linter
make swagger       # Generate Swagger docs
make migrate-up    # Run migrations
make docker-up     # Start Docker containers
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run vet
make vet
```

### Install Development Tools

```bash
make install-tools
```

This installs:
- swag (Swagger generation)
- golangci-lint (Linting)
- air (Hot-reload)
- migrate (Database migrations)

## 📦 AI Providers

| Provider | Tier | Quality | Status |
|----------|------|---------|--------|
| Gemini VEO | Premium | 4K | ✅ Implemented |
| OpenAI Sora | Premium | 4K | 🔜 Coming Soon |
| Runway Gen-3 | Standard | 1080p | 🔜 Coming Soon |
| Pika Labs | Budget | 720p | 🔜 Coming Soon |
| Mock Provider | Development | - | ✅ Implemented |

## 🔐 Security

- JWT-based authentication with access/refresh tokens
- Google OAuth 2.0 integration
- Rate limiting per user/tier
- Input validation on all endpoints
- SQL injection prevention via parameterized queries
- CORS configuration for production

## 📊 Scalability

Designed for 1,000,000+ concurrent users:

- **Horizontal Scaling** - Stateless API servers
- **Database** - PostgreSQL with read replicas & PgBouncer
- **Caching** - Redis Cluster for sessions and rate limiting
- **Job Queue** - Redis-based async job processing
- **CDN** - CloudFront/Cloudflare for video delivery

## 📝 License

MIT License - see [LICENSE](LICENSE) for details.

---

<div align="center">
  <p>Built with ❤️ by the Arabella Team</p>
</div>

