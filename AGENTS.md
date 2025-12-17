# AGENTS.md

This file provides context for AI agents working with this codebase.

## Project Overview

**dew** is a type-safe SQL query builder for Go with support for PostgreSQL (and extensible to other dialects). It provides a fluent API for constructing SELECT, INSERT, UPDATE, and DELETE queries with compile-time type safety using Go generics.

## Architecture

### Core Components

| File | Purpose |
|------|---------|
| `db.go` | Database connection wrapper (`DB` struct) with dialect support |
| `dialect.go` | SQL dialect interface (placeholder styles, etc.) |
| `schema.go` | Table and schema definitions (`Table[T]`, `Tabler` interface) |
| `column.go` | Typed column definitions (`IntColumn`, `StringColumn`, etc.) with comparison methods |
| `expr.go` | Expression interface and helpers (`And`, `Or`, `Raw`, aggregates) |
| `selector.go` | SELECT query builder (`Selector[T]`) |
| `insertor.go` | INSERT query builder (`Insertor[T]`) with conflict handling |
| `updator.go` | UPDATE query builder (`Updator[T]`) |
| `deletor.go` | DELETE query builder (`Deletor[T]`) |

### Key Patterns

1. **Generics for Type Safety**
   - All builders are generic: `Selector[T]`, `Insertor[T]`, `Updator[T]`, `Deletor[T]`
   - `T` represents the model struct that maps to database rows

2. **Fluent Builder Pattern**
   - All builder methods return `*Builder[T]` for chaining
   - Example: `UserSchema.From(db).Where(...).Select(...).Limit(10).All()`

3. **Expression Interface**
   ```go
   type Expression interface {
       Sql() string
       Args() []any
   }
   ```
   - All conditions, columns, and raw expressions implement this interface
   - Enables composable query building

4. **Column Types**
   - Each column type (`IntColumn`, `StringColumn`, etc.) has type-specific methods
   - `Eq()`, `Gt()`, `In()`, `Like()`, etc. return `Expression`

5. **Schema Definition**
   ```go
   var UserSchema = dew.DefineSchema("users", dew.PostgreSQLDialect{}, func(t dew.Table[User]) struct {
       dew.Table[User]
       ID    dew.IntColumn
       Name  dew.StringColumn
   } {
       return struct { ... }{
           Table: t,
           ID:    t.IntColumn("id"),
           Name:  t.StringColumn("name"),
       }
   })
   ```

## Coding Conventions

- **Error Handling**: Return errors from `ToSql()` and execution methods; never panic
- **Safety Checks**: `DELETE` and `UPDATE` require explicit `WHERE` clause (use `Raw("1=1")` to force all)
- **Placeholder Replacement**: Use `?` in expressions; dialect converts to `$1`, `$2`, etc. for PostgreSQL
- **Context**: Execution methods accept optional `context.Context` as variadic argument

## File Organization

```
dew/
├── cmd/dew-demo/     # Demo application
├── docs/             # Documentation (MkDocs + Next.js)
├── *_test.go         # Unit tests
└── *.go              # Core library
```

## Testing

Run tests with:
```bash
go test ./...
```

Demo requires PostgreSQL:
```bash
docker run --name dew-pg -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:16
go run ./cmd/dew-demo
```

## Common Tasks

### Adding a New Column Type
1. Define struct in `column.go` with `name`, `table`, `alias` fields
2. Implement `Column` interface (`Sql()`, `Args()`, `ColumnName()`, `TableName()`, `Alias()`)
3. Add comparison methods (`Eq()`, `NotEq()`, etc.) returning `Expression`
4. Add factory method to `Table[T]` in `schema.go`

### Adding a New Query Builder
1. Create new file (e.g., `upsert.go`)
2. Define generic struct with `db`, `table`, and query-specific fields
3. Implement builder methods returning `*Builder[T]`
4. Add `ToSql()`, `Exec()`, and optionally `Scan()` methods
5. Add factory method to `Table[T]` in `schema.go`

## Dependencies

- `database/sql` (standard library)
- `github.com/lib/pq` (PostgreSQL driver, only in demo)

