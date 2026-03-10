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

## CTEs (Common Table Expressions)

```go
// Simple CTE
active := Users.From(db).Where(Users.Age.Gte(18))

users, err := dew.From[User](db, dew.TableRef("active_users")).
    With(dew.CTE("active_users", active)).
    Where(dew.Raw("active_users.name = ?", "Alice")).
    All(ctx)

// WITH active_users AS (SELECT * FROM users WHERE users.age >= $1)
// SELECT * FROM active_users WHERE active_users.name = $2
```

Recursive CTEs for hierarchical data:

```go
base := dew.Raw("SELECT 1 AS n")
step := dew.Raw("SELECT n + 1 FROM nums WHERE n < ?", 10)
body := dew.Union[struct{ N int }](db, base, step)

rows, err := dew.From[struct{ N int }](db, dew.TableRef("nums")).
    With(dew.RecursiveCTE("nums", body)).
    All(ctx)

// WITH RECURSIVE nums AS ((SELECT 1 AS n) UNION (SELECT n + 1 FROM nums WHERE n < $1))
// SELECT * FROM nums
```

## UNION / INTERSECT / EXCEPT

```go
admins := Users.From(db).Select(Users.Name).Where(Users.Role.Eq("admin"))
mods := Users.From(db).Select(Users.Name).Where(Users.Role.Eq("mod"))

// UNION ALL with ordering
users, err := dew.UnionAll[User](db, admins, mods).
    OrderBy(dew.Asc(dew.Raw("name"))).
    Limit(50).
    All(ctx)

// (SELECT users.name FROM users WHERE users.role = $1)
// UNION ALL
// (SELECT users.name FROM users WHERE users.role = $2)
// ORDER BY name ASC LIMIT 50
```

Chain multiple set operations:

```go
q1 := Users.From(db).Select(Users.Name).Where(Users.Age.Gt(18))
q2 := Users.From(db).Select(Users.Name).Where(Users.Age.Lt(5))
q3 := Users.From(db).Select(Users.Name).Where(Users.Name.Eq("Admin"))

sql, args, _ := dew.Union[User](db, q1, q2).UnionAll(q3).ToSql()

// (SELECT ...) UNION (SELECT ...) UNION ALL (SELECT ...)
```

## Subquery FROM

```go
sub := Users.From(db).
    Select(Users.Name, dew.As(dew.Count(), "total")).
    GroupBy(Users.Name)

results, err := dew.FromSub[Result](db, sub, "sub").
    Where(dew.Raw("sub.total > ?", 10)).
    All(ctx)

// SELECT * FROM (SELECT users.name, COUNT(*) AS total FROM users
//   GROUP BY users.name) AS sub WHERE sub.total > $1
```

Set operations can also be used as CTE bodies or subquery sources — they implement the `Expression` interface.

## Dialects

```go
dew.PostgreSQLDialect{}  // $1, $2, ...
dew.MySQLDialect{}       // ?, ?, ...
dew.MSSQLDialect{}       // @p1, @p2, ...
dew.SQLiteDialect{}      // ?, ?, ...
```

## Documentation

Full documentation: **https://dew.xenous.org**

- [Getting Started](https://dew.xenous.org/docs) — install, connect, first query
- [Schema](https://dew.xenous.org/docs/schema) — models, column types, DefineSchema
- [SELECT](https://dew.xenous.org/docs/select) — queries, joins, RowScanner, ScanWith
- [INSERT](https://dew.xenous.org/docs/insert) — single/multi/batch insert, upsert
- [UPDATE](https://dew.xenous.org/docs/update) / [DELETE](https://dew.xenous.org/docs/delete) — with RETURNING support
- [CTEs](https://dew.xenous.org/docs/cte) / [Set Operations](https://dew.xenous.org/docs/set-operations) / [Subquery FROM](https://dew.xenous.org/docs/subquery-from)
- [Error Mapping](https://dew.xenous.org/docs/error-mapping) — driver-agnostic error handling
- [Dialects & Drivers](https://dew.xenous.org/docs/dialects) — PostgreSQL, MySQL, SQLite, MSSQL

## License

MIT - see [LICENSE](LICENSE) file for details.
