# Test Index — Bean & Brew Portal

> **Last updated: 2026-04-13**

## Summary
- Total test functions: 336 (~637 sub-tests with table-driven)
- Run: `go test -v ./...`
- Coverage: entity 100% · usecase 93.6% · repository 77.5% · handler 67.0%

## Test Files

### internal/entity/product_test.go
| Test | Type | Covers |
|------|------|--------|
| TestUnitProductValidate | unit | Product.Validate() — name, price, category_id |
| TestUnitProductApplyDiscount | unit | Product.ApplyDiscount() — 0%, 10%, 50%, 100%, negative, >100 |

### internal/entity/order_test.go
| Test | Type | Covers |
|------|------|--------|
| TestUnitOrderValidate | unit | Order.Validate() — customer_id, items, quantity, product_id |
| TestUnitOrderCalculateTotal | unit | Order.CalculateTotal() — multiple items |
| TestUnitOrderCalculateTotalEmpty | unit | Order.CalculateTotal() — empty items |
| TestUnitOrderCanCancel | unit | Order.CanCancel() — all 4 statuses |
| TestUnitOrderCanComplete | unit | Order.CanComplete() — all 4 statuses |

### internal/entity/news_test.go
| Test | Type | Covers |
|------|------|--------|
| TestUnitNewsItemValidate | unit | NewsItem.Validate() — title, content |
| TestUnitNewsItemSummary | unit | NewsItem.Summary() — truncation, UTF-8 runes |
| TestUnitCategoryValidate | unit | Category.Validate() — name, slug |

### internal/repository/sqlite/product_test.go
| Test | Type | Covers |
|------|------|--------|
| TestIntegrationProductCRUD | integration | Create, GetByID, Update, Delete |
| TestIntegrationProductList | integration | List all, by category, pagination |
| TestIntegrationProductSearch | integration | LIKE search by name/description |
| TestIntegrationProductCount | integration | Count all, by category, non-existent |
| TestIntegrationProductDeleteNotFound | integration | Delete non-existent product |

### internal/repository/sqlite/order_test.go
| Test | Type | Covers |
|------|------|--------|
| TestIntegrationOrderCRUD | integration | Create with items, GetByID with items |
| TestIntegrationOrderStatusTransitions | integration | new→processing→completed |
| TestIntegrationOrderListByCustomer | integration | Filter by customer_id, pagination |
| TestIntegrationOrderGetNotFound | integration | GetByID non-existent order |

### internal/handler/api_test.go
| Test | Type | Covers |
|------|------|--------|
| TestAPICategoryCRUD | api | POST + GET /categories |
| TestAPIProductCRUD | api | POST category → POST + GET /products |
| TestAPIProductValidationError | api | 400 on empty name, negative price |
| TestAPINewsCRUD | api | POST + GET /news |
| TestAPIOrderFlow | api | Full order: category→product→order, check total |
| TestAPIProductNotFound | api | 404 GET /products/99999 |
| TestAPIOrderEmptyItems | api | 400 POST order with empty items |

### internal/usecase/usecase_test.go
| Test | Type | Covers |
|------|------|--------|
| TestUnitProductUCCreateValidatesCategory | usecase | Rejects non-existent category |
| TestUnitProductUCCreateValidatesProduct | usecase | Rejects empty name |
| TestUnitProductUCListPagination | usecase | 25 products, page 1=10, page 3=5 |
| TestUnitOrderUCCreateSetsPrice | usecase | Price from product, total calculation |
| TestUnitOrderUCCancelFlow | usecase | Cancel new → cancel cancelled = error |
| TestUnitOrderUCCompleteRequiresProcessing | usecase | Complete new = error, process+complete = ok |
| TestUnitNewsUCPagination | usecase | 15 news, page 1 = 5 items |

## Dependencies Between Tests
- Integration tests depend on `setupTestDB()` → `migrations.go`
- API tests depend on `setupTestServer()` → full handler wiring
- UseCase tests use real SQLite but through usecase layer
- All DB tests are isolated (separate `:memory:` instance per test)

## Extended Tests (added 2026-04)
After scaling from 62 → 336 test functions, the following `*_extended_test.go` files were added:
- `internal/entity/` — product, order, news extended validation and edge cases
- `internal/repository/sqlite/` — extended CRUD, pagination, edge cases for all entities
- `internal/handler/` — category, product, order, news API extended tests
- `internal/usecase/` — category, product, order, news UC extended tests

## Coverage Gaps
- No tests for HTML page handlers (PageHandler.Home, Catalog, NewsFeed)
- No tests for seed data correctness
- No E2E browser tests
- No performance/load tests
- No tests for concurrent access patterns
