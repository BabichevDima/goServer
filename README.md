# Chirpy API

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?logo=postgresql)](https://www.postgresql.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Twitter-like microblogging service API built with Go and PostgreSQL.

## Features

- User registration and authentication (JWT)
- Create, read, update, and delete chirps (posts)
- Rate limiting and metrics tracking
- Webhook integration
- Admin dashboard

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Make (optional)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/chirpy.git
   cd chirpy
   ```

2. Set up environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your credentials
   ```

3. Run the server:
   ```bash
   go run main.go
   ```

## API Documentation

### Base URL
`http://localhost:8080`

### Authentication
All endpoints except `/api/healthz`, `/api/login`, and `/api/users` require authentication.

Include the JWT token in the Authorization header:
```
Authorization: Bearer <your_token>
```

### Endpoints

#### Users
| Method | Endpoint       | Description          |
|--------|----------------|----------------------|
| POST   | `/api/users`   | Register new user    |
| PUT    | `/api/users`   | Update user details  |
| POST   | `/api/login`   | Authenticate user    |

#### Chirps
| Method | Endpoint            | Description            |
|--------|---------------------|---------------------------|
| POST   | `/api/chirps`       | Create new chirp         |
| GET    | `/api/chirps`       | Get all chirps            |
| GET    | `/api/chirps/{id}`  | Get specific chirp       |
| DELETE | `/api/chirps/{id}`  | Delete chirp              |

#### Authentication
| Method | Endpoint          | Description            |
|--------|-------------------|---------------------------|
| POST   | `/api/refresh`     | Refresh access token      |
| POST   | `/api/revoke`      | Revoke refresh token      |

#### Admin
| Method | Endpoint          | Description            |
|--------|-------------------|---------------------------|
| GET    | `/admin/metrics`  | View usage statistics     |
| POST   | `/admin/reset`     | Reset metrics counter     |

## Data Models

### User
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "created_at": "ISO8601 timestamp",
  "updated_at": "ISO8601 timestamp",
  "is_chirpy_red": false
}
```

### Chirp
```json
{
  "id": "uuid",
  "body": "Chirp content (140 chars max)",
  "user_id": "uuid",
  "created_at": "ISO8601 timestamp",
  "updated_at": "ISO8601 timestamp"
}
```

## Example Requests

### Register User
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

### Create Chirp
```bash
curl -X POST http://localhost:8080/api/chirps \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{"body": "Hello Chirpy world!"}'
```

### Get Chirps
```bash
curl http://localhost:8080/api/chirps
```

## Deployment

### Docker
```bash
docker-compose up -d
```

### Production
For production deployments, consider:
- Using a reverse proxy (Nginx, Caddy)
- Configuring TLS/SSL
- Setting up proper logging and monitoring

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Distributed under the MIT License. See `LICENSE` for more information.

## START APP
in "WSL" terminal run command:
- go build -o out && ./out
#### or
- go run .
