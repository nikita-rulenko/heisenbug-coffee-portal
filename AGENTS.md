# Bean & Brew — Портал кофейни

> **Last updated: 2026-04-14**

## Обзор
Go веб-приложение для кофейни: каталог продуктов, лента новостей, управление заказами.
Создано для исследования Heisenbug 2026 по управлению AI-контекстом при тестировании.

## Архитектура
Clean Architecture, 4 слоя:
- `internal/entity/` — доменные модели (Product, Category, NewsItem, Order) + валидация + ошибки
- `internal/usecase/` — бизнес-логика, зависит только от entity + интерфейсов repository
- `internal/repository/` — интерфейсы; `sqlite/` — реализация на SQLite
- `internal/handler/` — HTTP-обработчики (JSON API + HTML-страницы)

## Стек технологий
- Go 1.25, Chi router, html/template + htmx, SQLite (WAL mode)
- Без ORM — raw SQL через `database/sql`

## Ключевые файлы
- `cmd/server/main.go` — точка входа, wiring
- `internal/repository/sqlite/migrations.go` — схема БД
- `internal/repository/sqlite/seed.go` — демо-данные (5 категорий, 17 продуктов, 3 новости)

## API
Все JSON-эндпоинты под `/api/v1/`:
- `GET/POST /products`, `GET/PUT/DELETE /products/{id}`, `GET /products/search?q=`
- `GET/POST /categories`, `GET/PUT/DELETE /categories/{id}`
- `GET/POST /news`, `GET/PUT/DELETE /news/{id}`
- `POST /orders`, `GET /orders/{id}`, `GET /orders/customer/{customerID}`
- `POST /orders/{id}/cancel`, `POST /orders/{id}/process`, `POST /orders/{id}/complete`

## Тестирование
336 тестовых функций (~637 sub-tests через table-driven) на 4 уровнях — запуск `go test ./...`.
> **Примечание:** 336 = количество `func Test*()`; ~637 = отдельные кейсы `t.Run()` внутри table-driven. Подробнее в `docs/known_issues.md`.
- **Unit** (`internal/entity/*_test.go`): валидация, ApplyDiscount, Summary, CalculateTotal, CanCancel/CanComplete
- **Integration** (`internal/repository/sqlite/*_test.go`): CRUD, поиск, пагинация, переходы статусов
- **API** (`internal/handler/*_test.go`): полные HTTP-эндпоинты через httptest
- **UseCase** (`internal/usecase/*_test.go`): бизнес-логика с реальным in-memory SQLite

### Покрытие тестами (2026-04-14)
| Слой | Coverage |
|------|----------|
| entity | 100.0% |
| usecase | 93.6% |
| repository/sqlite | 77.5% |
| handler | 67.0% |
| cmd/server | 0.0% (main — не тестируется) |

Проверить: `go test -cover ./...`

## Источники контекста
- **MD файлы**: этот файл + `docs/test-index.md`, `docs/test-context.md`, `docs/test-patterns.md`, `docs/known_issues.md`
- **Cursor rules**: `.cursor/rules/architecture.mdc`, `testing.mdc`, `github.mdc`, `mem0.mdc`, `helixir.mdc`
- **Промты онбординга**: `prompts/` — 4 промта для разных подходов к контексту
- **GitHub Issues**: [issues](https://github.com/nikita-rulenko/heisenbug-coffee-portal/issues) — трекинг работы через MCP

## Конвенции
- Именование тестов: `TestUnit*`, `TestIntegration*`, `TestAPI*`
- Предпочтительно table-driven тесты
- In-memory SQLite (`:memory:`) для изоляции тестов
- Хелпер `setupTestDB(t)` в `sqlite/testhelper_test.go`
- **GitHub Issues**: при работе по тикету оставляй комментарии о ходе работы (см. `.cursor/rules/github.mdc`)
