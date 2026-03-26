# Test Context — Bean & Brew Portal

## Project Domain
Coffee shop e-commerce portal with 4 core entities:

### Product
- Fields: ID, CategoryID, Name, Description, Price, ImageURL, InStock, CreatedAt, UpdatedAt
- Validation: Name required, Price >= 0, CategoryID > 0
- Business logic: ApplyDiscount(percent) returns discounted price
- DB: `products` table with `idx_products_category` index

### Category
- Fields: ID, Name, Slug, Description, SortOrder, CreatedAt
- Validation: Name required, Slug required (unique)
- Slug used for URL-friendly category identifiers
- DB: `categories` table

### NewsItem
- Fields: ID, Title, Content, Author, PublishedAt, CreatedAt, UpdatedAt
- Validation: Title required, Content required
- Business logic: Summary(maxRunes) truncates with "..." for preview
- DB: `news` table, ordered by published_at DESC

### Order
- Fields: ID, CustomerID, Status, Items, Total, CreatedAt, UpdatedAt
- OrderItem: ID, OrderID, ProductID, Quantity, Price
- Status flow: new → processing → completed (or new/processing → cancelled)
- Validation: CustomerID required, at least 1 item, quantity > 0, product_id > 0
- Business logic: CalculateTotal(), CanCancel(), CanComplete()
- DB: `orders` + `order_items` tables, transactional create

## Test Execution Environment
- Go 1.25, CGO_ENABLED=1 (for SQLite)
- SQLite in-memory (`:memory:`) — each test gets clean instance
- No external dependencies (no network, no Docker)
- Test run time: ~3 seconds total

## Test Data Strategy
- Integration tests create their own data via `seedCategory()` helper
- API tests wire full server stack and create data through HTTP
- No shared test fixtures file — data created inline

## Critical Paths for Testing
1. Product CRUD: category must exist before product
2. Order creation: product must exist, price is fetched from product
3. Order status transitions: invalid transitions return ErrInvalidStatus
4. Search: depends on LIKE query behavior with Cyrillic text
5. Pagination: offset/limit must be validated in usecase layer
