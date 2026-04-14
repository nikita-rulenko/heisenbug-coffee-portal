# Known Issues & Flaky Tests — Bean & Brew

> **Last updated: 2026-04-14**

## How to count tests

This project reports two numbers:
- **336 test functions** — `func Test*(t *testing.T)` registered in `_test.go` files
- **~637 sub-tests** — individual cases inside table-driven tests via `t.Run()`

`go test ./...` counts functions. `go test -v ./...` shows all sub-tests.
Both numbers are valid; we always report them together: "336 functions (~637 sub-tests)".

## Potentially Flaky Tests

### TestIntegrationProductSearch
- **Risk:** Uses `LIKE '%молок%'` — depends on exact seed data and insert order
- **Root cause:** Cyrillic LIKE behavior may vary across SQLite versions
- **Mitigation:** Each test gets isolated `:memory:` DB, so no cross-test data leaks
- **Status:** Stable in current setup, but fragile if seed data changes

### TestUnitNewsItemSummary
- **History:** Was buggy — used `len()` (bytes) instead of `[]rune` for UTF-8 truncation
- **Fix:** Now correctly counts runes, handles Cyrillic properly
- **Status:** Fixed, stable

### TestUnitNewsItemSummaryNegativeMaxRunes
- **Behavior:** Panics on negative input (slice bounds out of range)
- **Status:** Expected — test verifies panic with `recover()`

## Known Bugs / Limitations

### TestAPICategoryCreateDuplicateSlug
- **Issue:** Returns HTTP 500 instead of 409 Conflict
- **Root cause:** SQLite UNIQUE constraint error is not mapped to `ErrAlreadyExists` in repository
- **Impact:** Low — duplicate slugs are an edge case
- **Fix:** Map UNIQUE constraint violation to `entity.ErrAlreadyExists` in `sqlite/category.go`

### Float Precision in Calculations
- **Where:** `TestUnitOrderCalculateTotalFloatPrecision`, `TestUnitProductApplyDiscount`
- **Issue:** IEEE 754 float arithmetic: `0.1*10 + 0.2*5` is not exactly `2.0`
- **Mitigation:** Tests use tolerance check (+-0.1)
- **Recommendation:** Consider `int64` cents for production use

### Case-Insensitive Search
- **Where:** Product search via `LIKE`
- **Issue:** SQLite LIKE is case-insensitive for ASCII but not for all Unicode
- **Impact:** Cyrillic search works because SQLite treats it as binary comparison
- **Status:** Acceptable for demo project

## Missing Test Coverage

| Area | Current | Gap |
|------|---------|-----|
| HTML page handlers (Home, Catalog, NewsFeed) | 0% | No tests at all |
| Handler package overall | 67.0% | Target: 80%+ |
| Repository package overall | 77.5% | Target: 85%+ |
| cmd/server (main) | 0.0% | Not testable (wiring only) |
| E2E browser tests | none | Not implemented |
| Concurrent access patterns | none | Not tested |
| t.Parallel() | not used | Can be added to unit tests safely |

## Related Issues
- #10 — Handler coverage improvement
- #11 — t.Parallel() for unit tests
- #12 — E2E browser tests
- #13 — Repository coverage improvement
- #14 — Overall analysis and roadmap
