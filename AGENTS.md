# AGENTS.md

Context for AI agents working with this codebase.

## Project Overview

**dew** is a type-safe SQL query builder for Go. Supports PostgreSQL, MySQL, MSSQL, and SQLite dialects. Fluent API for SELECT, INSERT, UPDATE, and DELETE with compile-time type safety via Go 1.24 generics.

## Architecture

### Core Files

| File | Purpose |
|------|---------|
| `db.go` | `DB` and `Tx` wrappers, `Querier` interface (shared by both) |
| `dialect.go` | `Dialect` interface — placeholder styles per database |
| `schema.go` | `Table[T]`, `Tabler`, `TableRef`, `DefineSchema` |
| `column.go` | Typed columns (`IntColumn`, `StringColumn`, `BoolColumn`, `FloatColumn`, `TimeColumn`, `UUIDColumn`, `JSONBColumn[T]`) with comparison methods |
| `expr.go` | `Expression` interface, `And`, `Or`, `Raw`, aggregates (`Count`, `Sum`, etc.), `Asc`/`Desc`, `As` |
| `selector.go` | `Selector[T]` — SELECT builder, CTEs (`With`), subquery FROM (`FromSub`), `RowScanner` interface, scanning helpers |
| `inserter.go` | `Inserter[T]` — INSERT builder with `Models`/`Values`, `Batch`, `OnConflict` → `ConflictInserter[T]` |
| `updater.go` | `Updater[T]` — UPDATE builder |
| `deleter.go` | `Deleter[T]` — DELETE builder |
| `cte.go` | `CTE()`, `RecursiveCTE()` constructors |
| `set_query.go` | `SetQuery[T]` — UNION, UNION ALL, INTERSECT, EXCEPT |

### Key Patterns

1. **Generics** — All builders are `[T any]` where `T` is the model struct.

2. **Fluent chaining** — Builder methods return `*Builder[T]` for chaining:
   ```go
   Users.From(db).Where(Users.Age.Gt(18)).Select(Users.Name).Limit(10).All()
   ```

3. **Expression interface** — All conditions, columns, raw SQL implement:
   ```go
   type Expression interface {
       Sql() string
       Args() []any
   }
   ```

4. **Deferred placeholder replacement** — WHERE expressions stored as raw `Expression` objects with `?` placeholders. `buildRawQuery()` collects args; `buildQuery()` calls `replacePlaceholders()` for dialect-specific output (`$1`, `?`, `@p1`). This allows CTEs, subquery FROM, and set operations to compose correctly.

5. **RowScanner** — Structs can implement `ScanRow(*sql.Rows) error` to bypass reflection-based scanning. Detected at runtime via type assertion. Used by `All()`, `One()`, `First()`, `Scan()`, and `SetQuery.All()`.

6. **Batch insert** — `Batch(size)` on `Inserter`/`ConflictInserter` splits large inserts into multiple statements. `BatchQueries()` returns per-chunk SQL for inspection.

## Coding Conventions

- **Error handling**: Return errors from `ToSql()` and execution methods; never panic
- **Safety**: `DELETE` and `UPDATE` require explicit `WHERE` (use `Raw("1=1")` to force all)
- **Placeholders**: Use `?` in expressions; `??` escapes literal `?` (JSONB operators); dialect converts at build time
- **Context**: Execution methods accept optional `context.Context` as variadic arg
- **Querier interface**: All builders accept `Querier` (satisfied by both `*DB` and `*Tx`)

## File Organization

```
dew/
├── cmd/dew-demo/     # Demo application (requires PostgreSQL)
├── docs/dew/         # Documentation site (Fumadocs / Next.js)
├── *_test.go         # Tests (go test ./...)
└── *.go              # Core library (single package)
```

## Testing

```bash
go test ./...
```

## Common Tasks

### Adding a New Column Type
1. Define struct in `column.go` with `name`, `table`, `alias` fields
2. Implement `Column` interface (`Sql()`, `Args()`, `ColumnName()`, `TableName()`, `Alias()`)
3. Add comparison methods (`Eq()`, `NotEq()`, etc.) returning `Expression`
4. Add factory method to `Table[T]` in `schema.go`

### Adding a New Query Builder
1. Create new file
2. Define generic struct with `db Querier`, query-specific fields
3. Implement builder methods returning `*Builder[T]`
4. Add `ToSql()`, `Exec()`, and optionally `Scan()`/`All()` methods
5. Add factory method to `Table[T]` in `schema.go`

## Dependencies

- `database/sql` (standard library only)
- `github.com/lib/pq` (PostgreSQL driver, demo only)
