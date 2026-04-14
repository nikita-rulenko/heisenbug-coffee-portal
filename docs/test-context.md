# Контекст тестирования — Bean & Brew Portal

> **Last updated: 2026-04-14**

## Доменная область
Портал кофейни с e-commerce функциональностью, 4 основные сущности:

### Product (Продукт)
- Поля: ID, CategoryID, Name, Description, Price, ImageURL, InStock, CreatedAt, UpdatedAt
- Валидация: Name обязателен, Price >= 0, CategoryID > 0
- Бизнес-логика: ApplyDiscount(percent) — возвращает цену со скидкой
- БД: таблица `products`, индекс `idx_products_category`

### Category (Категория)
- Поля: ID, Name, Slug, Description, SortOrder, CreatedAt
- Валидация: Name обязателен, Slug обязателен (уникальный)
- Slug используется для URL-friendly идентификаторов категорий
- БД: таблица `categories`

### NewsItem (Новость)
- Поля: ID, Title, Content, Author, PublishedAt, CreatedAt, UpdatedAt
- Валидация: Title обязателен, Content обязателен
- Бизнес-логика: Summary(maxRunes) — обрезает с "..." для превью
- БД: таблица `news`, сортировка по published_at DESC

### Order (Заказ)
- Поля: ID, CustomerID, Status, Items, Total, CreatedAt, UpdatedAt
- OrderItem: ID, OrderID, ProductID, Quantity, Price
- Статусы: new → processing → completed (или new/processing → cancelled)
- Валидация: CustomerID обязателен, минимум 1 товар, quantity > 0, product_id > 0
- Бизнес-логика: CalculateTotal(), CanCancel(), CanComplete()
- БД: таблицы `orders` + `order_items`, транзакционное создание

## Среда выполнения тестов
- Go 1.25, CGO_ENABLED=1 (для SQLite)
- SQLite in-memory (`:memory:`) — каждый тест получает чистый экземпляр
- Нет внешних зависимостей (нет сети, нет Docker)
- Время прогона: ~4 секунды
- 336 тестовых функций (`func Test*()`), ~637 sub-tests (`t.Run()` внутри table-driven)
- Известные проблемы и flaky тесты: см. `known_issues.md`

## Покрытие тестами (2026-04-14)
| Пакет | Coverage |
|-------|----------|
| internal/entity | 100.0% |
| internal/usecase | 93.6% |
| internal/repository/sqlite | 77.5% |
| internal/handler | 67.0% |
| cmd/server | 0.0% |

## Стратегия тестовых данных
- Integration тесты создают свои данные через хелпер `seedCategory()`
- API тесты поднимают полный стек сервера и создают данные через HTTP
- Нет общего файла с фикстурами — данные создаются inline

## Критические пути для тестирования
1. Product CRUD: категория должна существовать до создания продукта
2. Создание заказа: продукт должен существовать, цена берётся из продукта
3. Переходы статусов заказа: невалидные переходы возвращают ErrInvalidStatus
4. Поиск: зависит от поведения LIKE с кириллическим текстом
5. Пагинация: offset/limit должны валидироваться в слое usecase
