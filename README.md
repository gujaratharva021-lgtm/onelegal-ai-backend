# LegalTech AI - Go Backend

Auth-only scope (Day 1): Signup, Login, JWT-protected Profile.

## Folder Structure

```
legaltech-backend/
├── cmd/server/main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── models/
│   ├── handlers/
│   ├── middleware/
│   ├── routes/
│   ├── services/
│   └── utils/
├── pkg/response/
├── docker-compose.yml
├── Dockerfile
└── .env.example
```

## Option A — Run with Docker (recommended, no local Go/Postgres needed)

```bash
cp .env.example .env
# edit .env and set a real JWT_SECRET

docker compose up --build
```

This starts Postgres + the Go backend together.
API will be available at: http://localhost:8080

## Option B — Run locally (Go + Postgres installed on your machine)

1. Make sure PostgreSQL is running and a database exists:

```sql
CREATE DATABASE legaltech_db;
```

2. Copy env file and edit values:

```bash
cp .env.example .env
```

3. Install dependencies and run:

```bash
go mod tidy
go run ./cmd/server
```

The server auto-migrates the `users` table on startup (via GORM AutoMigrate),
so you do not need to run the SQL file manually — it's there for reference only.

## Endpoints

| Method | Endpoint                  | Auth required | Description        |
|--------|----------------------------|----------------|---------------------|
| GET    | /health                    | No             | Health check        |
| POST   | /api/v1/auth/signup         | No             | Register new user   |
| POST   | /api/v1/auth/login          | No             | Login, returns JWT  |
| GET    | /api/v1/user/profile         | Yes (Bearer)   | Get logged-in user  |

### Signup body
```json
{
  "name": "Adv. Rajesh Kumar",
  "email": "rajesh@example.com",
  "password": "secret123",
  "phone": "+91 98765 43210"
}
```

### Login body
```json
{
  "email": "rajesh@example.com",
  "password": "secret123"
}
```

### Profile request
```
GET /api/v1/user/profile
Authorization: Bearer <token>
```

## Connecting the Flutter app

In the Flutter project, set `lib/core/constants/api_endpoints.dart` -> `baseUrl`
to point here. Defaults assume Android emulator: `http://10.0.2.2:8080/api/v1`.
For a physical device, use your machine's LAN IP instead of `10.0.2.2`.
