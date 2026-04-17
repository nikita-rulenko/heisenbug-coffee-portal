# Каталог тестов — Bean & Brew Portal

> **Last updated: 2026-04-17**

Это **оглавление** тестов по группам, не реестр функций. Конкретные имена
тестов смотри `grep -r 'TestXxx' internal/` или прямо в файлах. Числа
функций обновляй только когда добавилась/удалилась группа (новый файл
`*_test.go`), а не на каждую новую функцию.

## Сводка
- **336 тестовых функций**, **~637 прогонов с sub-tests** (`go test -v`)
- Запуск: `go test ./... -count=1` · покрытие: `go test -cover ./...`
- Coverage: entity 100% · usecase 93.6% · repository 77.5% · handler 67.0%
- Известные баги, flaky, пробелы покрытия → `known_issues.md`
- Паттерны и антипаттерны → `test-patterns.md`

## Слой entity (53 функции)

Чистые unit-тесты доменных правил, без БД и сети.

- `product_test.go` / `product_extended_test.go` — `Product.Validate`, `ApplyDiscount` (граничные значения)
- `order_test.go` / `order_extended_test.go` — `Order.Validate`, `CalculateTotal`, переходы статусов (`CanCancel`, `CanComplete`)
- `news_test.go` / `news_extended_test.go` — `NewsItem.Validate`, `Summary` (UTF-8 через `[]rune`)
- `category_test.go` — `Category.Validate`

## Слой repository/sqlite (80 функций)

Integration-тесты против реального `:memory:` SQLite через `setupTestDB(t)`.

- `product_test.go` / `product_extended_test.go` — CRUD, фильтр по категории, LIKE-поиск, пагинация, count, `DeleteNotFound`
- `order_test.go` / `order_extended_test.go` — CRUD с items, статус-переходы, фильтр по customer_id, пагинация, `GetNotFound`
- `news_test.go` / `news_extended_test.go` — CRUD, пагинация, edge cases
- `category_test.go` / `category_extended_test.go` — CRUD, дубликат slug (см. #15), edge cases
- `testhelper_test.go` — `setupTestDB` хелпер, без `Test*` функций

## Слой usecase (97 функций)

Unit-тесты бизнес-логики через реальный SQLite-слой (быстрее моков).

- `usecase_test.go` — базовые сценарии всех use-case'ов (validate, pagination, status flow)
- `product_uc_extended_test.go` — расширенные кейсы Product UC
- `order_uc_extended_test.go` — расширенные кейсы Order UC (cancel, complete, цена из продукта)
- `news_uc_extended_test.go` — расширенные кейсы News UC
- `category_uc_extended_test.go` — расширенные кейсы Category UC

## Слой handler — API (106 функций)

API-тесты через `setupTestServer(t)` — полная связка router → handler → usecase → repo.

- `api_test.go` — flow-тесты: CRUD каждой сущности, валидация (400), not found (404), полный заказ-flow
- `category_api_extended_test.go` — расширенные API-кейсы категорий
- `product_api_extended_test.go` — расширенные API-кейсы продуктов
- `order_api_extended_test.go` — расширенные API-кейсы заказов
- `news_api_extended_test.go` — расширенные API-кейсы новостей

## Слой handler — HTML pages и auth (нет тестов)

- `pages.go` (16 методов: Home, Catalog, NewsFeed, Cart, Checkout, OrderConfirmation, SearchFragment, AdminNews/Create/Delete, Login*, Logout) — **0% покрытия**, см. #10
- `auth.go` (AuthMiddleware, AdminOnly, GetUsername, IsAdmin) — **0% покрытия**, см. #10

## Зависимости и изоляция

- Каждый integration/API-тест получает **свой** `:memory:` SQLite — нет shared state
- `setupTestDB(t)` поднимает миграции из `migrations.go`
- `setupTestServer(t)` собирает chain handlers с реальным usecase и репозиторием
- `t.Parallel()` сейчас **нигде** не используется (см. #11)

## Что в каталоге НЕТ (намеренно)

- Имён конкретных функций — это шум, ищи `grep`
- Sub-test names — `t.Run("case", ...)` живут в коде
- Багов и flaky — это в `known_issues.md`
- Coverage по конкретному файлу — это `go test -cover ./...`
