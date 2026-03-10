# Dew

[![CI](https://github.com/dr3dnought/dew/actions/workflows/ci.yml/badge.svg)](https://github.com/dr3dnought/dew/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/dr3dnought/dew/branch/dev/graph/badge.svg)](https://codecov.io/gh/dr3dnought/dew)
[![Go Reference](https://pkg.go.dev/badge/github.com/dr3dnought/dew.svg)](https://pkg.go.dev/github.com/dr3dnought/dew)
[![Go Report Card](https://goreportcard.com/badge/github.com/dr3dnought/dew)](https://goreportcard.com/report/github.com/dr3dnought/dew)

A lightweight, type-safe query builder for Go. No ORM magic, no repo layers — just queries.

**[Documentation](https://dew.xenous.org)** | **[Getting Started](https://dew.xenous.org/docs)** | **[Examples](https://dew.xenous.org/docs/examples-vs-sql)**

## Philosophy

Dew is not an ORM — it's a query builder that's expressive enough to replace the repository layer entirely. Instead of wrapping queries behind interfaces, you write them inline where you need them:

```go
user, err := Users.From(db).Where(Users.Email.Eq(email)).One(ctx)
```

Traditional repository layers add indirection without adding safety — you still write SQL-shaped code inside them. Dew gives you type-safe, composable queries that read like SQL, so the abstraction becomes unnecessary.

Every builder accepts `dew.Querier` (satisfied by both `*DB` and `*Tx`), so transaction support comes for free — pass `tx` instead of `db`, same code, no wrapper needed. Zero codegen, zero reflection at build time, just Go generics.

## Install

```bash
go get github.com/dr3dnought/dew
```

## Quick Start

```go
db, err := dew.Open("postgres", connStr, dew.PostgreSQLDialect{})

// SELECT
users, err := Users.From(db).
    Where(Users.Age.Gte(18)).
    OrderBy(dew.Desc(Users.Name)).
    Limit(10).
    All(ctx)

// INSERT
err = Users.Insert(db).
    Columns(Users.Name, Users.Email).
    Values("Alice", "alice@example.com").
    Exec(ctx)

// UPDATE
err = Users.Update(db).
    Set(Users.Name, "Bob").
    Where(Users.ID.Eq(1)).
    Exec(ctx)

// DELETE
err = Users.Delete(db).
    Where(Users.ID.Eq(1)).
    Exec(ctx)
```

## Supported Dialects

| Dialect | Placeholders |
|---------|-------------|
| PostgreSQL | `$1, $2, ...` |
| MySQL | `?, ?, ...` |
| SQLite | `?, ?, ...` |
| MSSQL | `@p1, @p2, ...` |

Works with any `database/sql` driver — `pgx`, `lib/pq`, `go-sql-driver/mysql`, `modernc/sqlite`, and more.

## Documentation

Full docs at **[dew.xenous.org](https://dew.xenous.org)** — schema definition, joins, CTEs, batch insert, error mapping, JSONB, set operations, and more.

## License

MIT - see [LICENSE](LICENSE) file for details.
