# Bean & Brew — Coffee Shop Portal

## Overview
Go web application for a coffee shop with product catalog, news feed, and order management.
Built for Heisenbug 2026 research on AI test context management.

## Architecture
Clean Architecture with 4 layers:
- `internal/entity/` — domain models (Product, Category, NewsItem, Order) + validation + errors
- `internal/usecase/` — business logic, depends only on entity + repository interfaces
- `internal/repository/` — interfaces; `sqlite/` has SQLite implementation
- `internal/handler/` — HTTP handlers (JSON API + HTML pages)

## Tech Stack
- Go 1.25, Chi router, html/template + htmx, SQLite (WAL mode)
- No ORM — raw SQL with `database/sql`

## Key Files
- `cmd/server/main.go` — entry point, wiring
- `internal/repository/sqlite/migrations.go` — schema
- `internal/repository/sqlite/seed.go` — demo data (5 categories, 17 products, 3 news)

## API
All JSON endpoints under `/api/v1/`:
- `GET/POST /products`, `GET/PUT/DELETE /products/{id}`, `GET /products/search?q=`
- `GET/POST /categories`, `GET/PUT/DELETE /categories/{id}`
- `GET/POST /news`, `GET/PUT/DELETE /news/{id}`
- `POST /orders`, `GET /orders/{id}`, `GET /orders/customer/{customerID}`
- `POST /orders/{id}/cancel`, `POST /orders/{id}/process`, `POST /orders/{id}/complete`

## Testing
62 tests across 4 levels — run `go test ./...`:
- **Unit** (`internal/entity/*_test.go`): validation, ApplyDiscount, Summary, CalculateTotal, CanCancel/CanComplete
- **Integration** (`internal/repository/sqlite/*_test.go`): CRUD, search, pagination, status transitions
- **API** (`internal/handler/api_test.go`): full HTTP endpoints via httptest
- **UseCase** (`internal/usecase/usecase_test.go`): business logic with real in-memory SQLite

## Conventions
- Test names: `TestUnit*`, `TestIntegration*`, `TestAPI*`
- Table-driven tests preferred
- In-memory SQLite (`:memory:`) for test isolation
- `setupTestDB(t)` helper in `sqlite/testhelper_test.go`
