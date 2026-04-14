# Индекс тестов — Bean & Brew Portal

> **Last updated: 2026-04-14**

## Сводка
- **336 тестовых функций** — `func Test*()` в коде
- **~637 sub-tests** — отдельные кейсы внутри table-driven тестов через `t.Run()`
- Запуск: `go test -v ./...` (флаг `-v` покажет sub-tests)
- Coverage: entity 100% · usecase 93.6% · repository 77.5% · handler 67.0%
- Известные проблемы и flaky тесты: см. `known_issues.md`

## Тестовые файлы

### internal/entity/product_test.go
| Тест | Тип | Покрывает |
|------|-----|-----------|
| TestUnitProductValidate | unit | Product.Validate() — name, price, category_id |
| TestUnitProductApplyDiscount | unit | Product.ApplyDiscount() — 0%, 10%, 50%, 100%, negative, >100 |

### internal/entity/order_test.go
| Тест | Тип | Покрывает |
|------|-----|-----------|
| TestUnitOrderValidate | unit | Order.Validate() — customer_id, items, quantity, product_id |
| TestUnitOrderCalculateTotal | unit | Order.CalculateTotal() — несколько товаров |
| TestUnitOrderCalculateTotalEmpty | unit | Order.CalculateTotal() — пустой список |
| TestUnitOrderCanCancel | unit | Order.CanCancel() — все 4 статуса |
| TestUnitOrderCanComplete | unit | Order.CanComplete() — все 4 статуса |

### internal/entity/news_test.go
| Тест | Тип | Покрывает |
|------|-----|-----------|
| TestUnitNewsItemValidate | unit | NewsItem.Validate() — title, content |
| TestUnitNewsItemSummary | unit | NewsItem.Summary() — обрезка, UTF-8 runes |
| TestUnitCategoryValidate | unit | Category.Validate() — name, slug |

### internal/repository/sqlite/product_test.go
| Тест | Тип | Покрывает |
|------|-----|-----------|
| TestIntegrationProductCRUD | integration | Create, GetByID, Update, Delete |
| TestIntegrationProductList | integration | Список: все, по категории, пагинация |
| TestIntegrationProductSearch | integration | LIKE-поиск по name/description |
| TestIntegrationProductCount | integration | Count: все, по категории, несуществующая |
| TestIntegrationProductDeleteNotFound | integration | Delete несуществующего продукта |

### internal/repository/sqlite/order_test.go
| Тест | Тип | Покрывает |
|------|-----|-----------|
| TestIntegrationOrderCRUD | integration | Создание с items, GetByID с items |
| TestIntegrationOrderStatusTransitions | integration | new→processing→completed |
| TestIntegrationOrderListByCustomer | integration | Фильтр по customer_id, пагинация |
| TestIntegrationOrderGetNotFound | integration | GetByID несуществующего заказа |

### internal/handler/api_test.go
| Тест | Тип | Покрывает |
|------|-----|-----------|
| TestAPICategoryCRUD | api | POST + GET /categories |
| TestAPIProductCRUD | api | POST category → POST + GET /products |
| TestAPIProductValidationError | api | 400 на пустое имя, отрицательную цену |
| TestAPINewsCRUD | api | POST + GET /news |
| TestAPIOrderFlow | api | Полный flow: category→product→order, проверка total |
| TestAPIProductNotFound | api | 404 GET /products/99999 |
| TestAPIOrderEmptyItems | api | 400 POST заказа с пустым items |

### internal/usecase/usecase_test.go
| Тест | Тип | Покрывает |
|------|-----|-----------|
| TestUnitProductUCCreateValidatesCategory | usecase | Отклоняет несуществующую категорию |
| TestUnitProductUCCreateValidatesProduct | usecase | Отклоняет пустое имя |
| TestUnitProductUCListPagination | usecase | 25 продуктов, page 1=10, page 3=5 |
| TestUnitOrderUCCreateSetsPrice | usecase | Цена из продукта, расчёт total |
| TestUnitOrderUCCancelFlow | usecase | Cancel new → cancel cancelled = ошибка |
| TestUnitOrderUCCompleteRequiresProcessing | usecase | Complete new = ошибка, process+complete = ok |
| TestUnitNewsUCPagination | usecase | 15 новостей, page 1 = 5 элементов |

## Зависимости между тестами
- Integration тесты зависят от `setupTestDB()` → `migrations.go`
- API тесты зависят от `setupTestServer()` → полная связка handlers
- UseCase тесты используют реальный SQLite через слой usecase
- Все DB тесты изолированы (отдельный `:memory:` на каждый тест)

## Расширенные тесты (добавлены 2026-04)
После масштабирования с 62 → 336 тестовых функций добавлены файлы `*_extended_test.go`:
- `internal/entity/` — расширенная валидация и edge cases для product, order, news
- `internal/repository/sqlite/` — расширенные CRUD, пагинация, edge cases для всех сущностей
- `internal/handler/` — расширенные API тесты для category, product, order, news
- `internal/usecase/` — расширенные UC тесты для category, product, order, news

## Пробелы в покрытии
- Нет тестов для HTML page handlers (PageHandler.Home, Catalog, NewsFeed)
- Нет тестов для корректности seed-данных
- Нет E2E browser тестов
- Нет тестов производительности/нагрузки
- Нет тестов конкурентного доступа
