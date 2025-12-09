# Dew

Lightweight type-safe ORM for Go with a fluent API and multiple dialects.

## Features

- Type-safe builders for SELECT/INSERT/DELETE
- PostgreSQL, SQLite, MySQL (placeholders handled by dialect)
- Automatic struct ↔ column mapping (snake_case)
- Subqueries, JOIN, GROUP BY, COUNT/COUNT DISTINCT, IN/IN (subquery)
- Parameterized queries to avoid SQL injection

## Docs map
- Getting Started — quick setup and CRUD basics
- Examples — runnable snippets from the demo app
- Selector — building selects, filters, joins, subqueries, ordering, DTO mapping
- Schema — defining tables/columns, dialect choice, column types
- API Reference — concise cheat sheet

## Quick Start (PostgreSQL)

```go
package main

import (
	"fmt"
	"log"

	"github.com/dr3dnought/dew"
	_ "github.com/lib/pq"
)

type User struct {
	ID    int
	Name  string
	Email string
}

var UserSchema = dew.DefineSchema("users", dew.PostgreSQLDialect{}, func(t dew.Table[User]) struct {
	dew.Table[User]
	ID    dew.IntColumn
	Name  dew.StringColumn
	Email dew.StringColumn
} {
	return struct {
		dew.Table[User]
		ID    dew.IntColumn
		Name  dew.StringColumn
		Email dew.StringColumn
	}{
		Table: t,
		ID:    t.IntColumn("id"),
		Name:  t.StringColumn("name"),
		Email: t.StringColumn("email"),
	}
})

func main() {
	dsn := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	db, err := dew.Open("postgres", dsn, dew.PostgreSQLDialect{})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table for the demo (dew doesn't manage migrations)
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL
	)`)

	// INSERT via Values
	if err := dew.Insert[User](db, UserSchema).
		Columns(UserSchema.Name, UserSchema.Email).
		Values("Alice", "alice@example.com").
		Exec(); err != nil {
		log.Fatal(err)
	}

	// SELECT
	users, err := UserSchema.From(db).
		Select(UserSchema.Name, UserSchema.Email).
		All()
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range users {
		fmt.Printf("User: %s <%s>\n", u.Name, u.Email)
	}
}
```

See also “Getting Started” and “Examples”.