# Dew

[![CI](https://github.com/dr3dnought/dew/actions/workflows/ci.yml/badge.svg)](https://github.com/dr3dnought/dew/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/dr3dnought/dew/branch/dev/graph/badge.svg)](https://codecov.io/gh/dr3dnought/dew)
[![Go Reference](https://pkg.go.dev/badge/github.com/dr3dnought/dew.svg)](https://pkg.go.dev/github.com/dr3dnought/dew)
[![Go Report Card](https://goreportcard.com/badge/github.com/dr3dnought/dew)](https://goreportcard.com/report/github.com/dr3dnought/dew)

A lightweight, type-safe query builder for Go. No ORM magic, no repo layers — just queries.

## Features

- **Type-safe** — typed columns, compile-time checked queries
- **Zero codegen** — generics only, no build step
- **Composable** — builder pattern with `Clone()` for query branching
- **Multi-dialect** — PostgreSQL, MySQL, MSSQL, SQLite
- **Transactions** — `Querier` interface works with both `*DB` and `*Tx`
- **JSONB** — first-class PostgreSQL JSONB column support

## Installation

```bash
go get github.com/dr3dnought/dew
```

## Quick Start

Define your schema once:

```go
type User struct {
    ID    int
    Name  string
    Email string
    Age   int
}

var Users = dew.DefineSchema("users", dew.PostgreSQLDialect{}, func(t dew.Table[User]) struct {
    dew.Table[User]
    ID    dew.IntColumn
    Name  dew.StringColumn
    Email dew.StringColumn
    Age   dew.IntColumn
} {
    return struct {
        dew.Table[User]
        ID    dew.IntColumn
        Name  dew.StringColumn
        Email dew.StringColumn
        Age   dew.IntColumn
    }{
        Table: t,
        ID:    t.IntColumn("id"),
        Name:  t.StringColumn("name"),
        Email: t.StringColumn("email"),
        Age:   t.IntColumn("age"),
    }
})
```

Query anywhere — no repository layer needed:

```go
db, err := dew.Open("postgres", connStr, dew.PostgreSQLDialect{})
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// SELECT
users, err := Users.From(db).
    Where(Users.Age.Gte(18)).
    OrderBy(dew.Desc(Users.Name)).
    Limit(10).
    All(ctx)

// INSERT
err = dew.Insert[User](db, Users).
    Values(&User{Name: "Alice", Email: "alice@example.com", Age: 30}).
    Exec(ctx)

// UPDATE
err = dew.Update[User](db, Users).
    Set(Users.Name, "Bob").
    Where(Users.ID.Eq(1)).
    Exec(ctx)

// DELETE
err = dew.Delete[User](db, Users).
    Where(Users.ID.Eq(1)).
    Exec(ctx)
```

## Transactions

All builders accept `Querier`, so they work with both `*DB` and `*Tx`:

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

err = dew.Insert[User](tx, Users).
    Values(&User{Name: "Alice", Age: 30}).
    Exec(ctx)
if err != nil {
    return err
}

return tx.Commit()
```

## Dialects

```go
dew.PostgreSQLDialect{}  // $1, $2, ...
dew.MySQLDialect{}       // ?, ?, ...
dew.MSSQLDialect{}       // @p1, @p2, ...
dew.SQLiteDialect{}      // ?, ?, ...
```

## Documentation

Full documentation: https://dr3dnought.github.io/dew/

## License

MIT - see [LICENSE](LICENSE) file for details.
