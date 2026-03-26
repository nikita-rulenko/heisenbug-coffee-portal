# Test Patterns & Anti-patterns — Bean & Brew

## Patterns Used

### 1. Table-Driven Tests
Most unit tests use subtests with `t.Run()`:
```go
tests := []struct {
    name    string
    input   Entity
    wantErr error
}{...}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { ... })
}
```
**Where:** product_test.go, order_test.go, news_test.go

### 2. Test Helpers with t.Helper()
```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sqliteRepo.Open(":memory:")
    ...
    t.Cleanup(func() { db.Close() })
    return db
}
```
**Where:** testhelper_test.go, api_test.go

### 3. In-Memory Database Isolation
Each test gets fresh `:memory:` SQLite — no cleanup needed, no data leaks.
**Benefit:** Tests can run in parallel without conflicts.

### 4. Seed Helper for Dependencies
```go
func seedCategory(t *testing.T, repo *CategoryRepo) *entity.Category { ... }
```
Creates prerequisite data for tests that need categories before products.

### 5. HTTP Test Server
```go
srv := setupTestServer(t)
defer srv.Close()
resp, _ := postJSON(srv.URL+"/api/v1/products", ...)
```
Full server stack via httptest.NewServer — tests real routing, middleware, serialization.

## Anti-patterns to Watch For

### 1. Test Interdependence
**Risk:** Tests relying on data created by other tests.
**Status:** Not present — each test creates its own data.
**If violated:** Tests may pass individually but fail in suite, or vice versa.

### 2. Hardcoded IDs
**Risk:** Assuming `ID=1` after first insert.
**Mitigation:** We use `p.ID` from Create result, not hardcoded values.

### 3. Time-Dependent Tests
**Risk:** Tests comparing `time.Now()` with stored timestamps.
**Status:** Not present, but `created_at` comparisons would be fragile.

### 4. Over-mocking
**Risk:** Mocking every dependency, losing integration coverage.
**Status:** We use real SQLite `:memory:` instead of mocks — faster and more realistic.

### 5. Testing Implementation, Not Behavior
**Risk:** Tests that break when internal implementation changes.
**Example:** Testing exact SQL query instead of result correctness.
**Mitigation:** Our tests check returned data, not internal calls.

## Known Issues

### Potential Flaky: TestIntegrationProductSearch
Uses `LIKE '%молок%'` which depends on exact data. If seed data changes
or another test inserts data with "молок" in the same DB, results change.
Currently mitigated by isolated `:memory:` per test.

### Missing: Parallel Test Execution
Tests don't use `t.Parallel()`. Could be added for unit tests safely.
Integration tests sharing DB instance should NOT be parallelized.

### Missing: Error Message Assertions
We check error type but not error messages. If an error wraps another,
`errors.Is()` handles it, but literal string comparison would break.
