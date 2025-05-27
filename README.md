# TakaDoo Banking System

A robust banking system API built with Go that handles user transactions, balance management, and administrative functions.

## Features

- User authentication and authorization (with JWT and role-based access)
- Transaction management (deposits, withdrawals, transfers)
- Balance tracking in multiple currencies (EUR supported, extensible)
- Admin panel for transaction monitoring
- Historical balance queries
- RESTful API interface
- Swagger API documentation
- Automated middleware tests

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 14 or higher
- Redis (for caching, optional but recommended)
- Docker and Docker Compose (optional, for containerized deployment)

## Project Structure

```
.
├── cmd/                    # Application entry points
│   ├── api/               # Main API server
│   ├── check_admin/       # Get all admin details
│   ├── migrate/           # Database migrate command
│   └── reset_admin/       # Reset admin password in database
├── internal/              # Private application code
│   ├── middleware/        # JWT authentication and role middleware
│   ├── models/            # Data models
│   ├── repository/        # Database interactions
│   ├── service/           # Business logic
│   └── handlers/          # HTTP handlers
├── migrations/            # Database migrations
├── docs/                  # API documentation (Swagger)
└── tests/                 # (For future integration tests)
```

## Environment Setup

1. Copy the example environment file and edit as needed:
```bash
cp .env.example .env
# Edit .env with your configuration
```

**Required .env variables:**
```
ADMIN_EMAIL=admin@takadao.com
ADMIN_PASSWORD=your_admin_password
JWT_SECRET=your_jwt_secret
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_db_password
DB_NAME=takadao
REDIS_HOST=localhost
REDIS_PORT=6379
SERVER_PORT=8080
```

## Running the Application

### With Docker Compose
```bash
docker-compose up -d
```

### Without Docker
- Make sure PostgreSQL and Redis are running and configured as in your `.env` file.

### Run database migrations
```bash
go run cmd/migrate/main.go
```

### Start the API server
```bash
go run cmd/api/main.go
```

### Generate Swagger documentation
```bash
swag init -g cmd/api/main.go -o docs
```

The API will be available at `http://localhost:8080`

## API Documentation (Swagger)

- **Local Swagger UI:** [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
- **Live Swagger UI:** [https://takadao-banking-api.onrender.com/swagger/index.html](https://takadao-banking-api.onrender.com/swagger/index.html)
- To authorize, click "Authorize" and enter your JWT as: `Bearer <token>`

## API Endpoints

### Authentication

- **User Login:** `POST /api/v1/auth/user/login`
- **Admin Login:** `POST /api/v1/auth/admin/login`
- **User Register:** `POST /api/v1/auth/user/register`
- **Admin Register:** `POST /api/v1/auth/admin/register` (requires admin token)

### User Endpoints

- **Get My Profile:** `GET /api/v1/users/me` (use this instead of `/users/profile`)
- **Get Balances:** `GET /api/v1/users/balances`
- **Deposit:** `POST /api/v1/transactions/deposit`
- **Withdraw:** `POST /api/v1/transactions/withdraw`
- **Transfer:** `POST /api/v1/transactions/transfer`
- **List My Transactions:** `GET /api/v1/transactions`
- **Get My Transaction:** `GET /api/v1/transactions/{id}`

### Admin Endpoints (require Bearer token with admin role)

- **List All Users:** `GET /api/v1/admin/users`
- **Get User:** `GET /api/v1/admin/users/{id}`
- **Update User:** `PUT /api/v1/admin/users/{id}`
- **Delete User:** `DELETE /api/v1/admin/users/{id}`
- **List All Transactions:** `GET /api/v1/admin/transactions`
- **Get Transaction:** `GET /api/v1/admin/transactions/{id}`
- **Get User Balance at Time:** `GET /api/v1/admin/users/{user_id}/balance?at_time=...`

## Testing

### Run all tests
```bash
go test ./...
```

### Run middleware tests only
```bash
go test ./internal/middleware -v
```

- Tests for authentication and admin middleware are in `internal/middleware/auth_middleware_test.go`.
- Add more tests in the corresponding `*_test.go` files in each package.

## Testing Production API

You can test the live production API using the Swagger UI:

- **Live Swagger UI:** [https://takadao-banking-api.onrender.com/swagger/index.html](https://takadao-banking-api.onrender.com/swagger/index.html)

### 1. Add Admin Credentials
- Use the admin login endpoint to obtain a JWT token:
  - **Endpoint:** `POST /api/v1/auth/admin/login`
  - **Body:**
    ```json
    {
      "email": "admin@takadao.com",
      "password": "admin123"
    }
    ```
- The response will include a JWT token.

### 2. Authorize with Bearer JWT Token
- In the Swagger UI, click the "Authorize" button.
- Enter your token in the following format:
  ```
  Bearer <your_jwt_token>
  ```
- Now you can access protected admin endpoints.

## Performance Considerations

The system is designed to handle high transaction volumes with the following optimizations:

1. **Balance Calculation**: 
   - Real-time balance is maintained using a materialized view
   - Historical balances are calculated using efficient date-based indexing
   - For high-frequency transactions, we use a caching layer with Redis

2. **Database Optimization**:
   - Indexed queries for transaction lookups
   - Partitioned tables for transaction history
   - Efficient foreign key relationships

3. **Scalability**:
   - Horizontal scaling of API servers
   - Read replicas for balance queries
   - Message queue for transaction processing

## Security

- JWT-based authentication (with role-based access)
- Password hashing using bcrypt
- Input validation and sanitization

