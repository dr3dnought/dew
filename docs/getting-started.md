# Getting Started

## Requirements
- Go 1.21+ (any recent release is fine)
- DB driver: `github.com/lib/pq` (PostgreSQL), `github.com/mattn/go-sqlite3`, or your MySQL driver of choice

## Install
```bash
go get github.com/dr3dnought/dew
# plus a driver, e.g.:
go get github.com/lib/pq
```

## Connect & choose dialect
Dialects handle placeholders correctly (`$1` for Postgres, `?` for SQLite/MySQL).
```go
db, err := dew.Open("postgres", os.Getenv("PG_DSN"), dew.PostgreSQLDialect{})
// or sqlite:
// db, err := dew.Open("sqlite3", ":memory:", dew.SQLiteDialect{})
```

## Define schema
`DefineSchema` ties a table to a Go type and defines columns.
```go
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
```

> Dew doesn’t handle migrations — create tables yourself (SQL or migrator).

## Basic operations

### INSERT
```go
err := dew.Insert[User](db, UserSchema).
	Columns(UserSchema.Name, UserSchema.Email).
	Values("Alice", "alice@example.com").
	Exec()
```

### SELECT
```go
users, err := UserSchema.From(db).
	Select(UserSchema.ID, UserSchema.Name).
	Where(UserSchema.Name.Eq("Alice")).
	All()
```

### DELETE
```go
deleted, err := dew.Delete[User](db, UserSchema).
	Where(UserSchema.ID.Eq(42)).
	RowsAffected()
```

More examples live in the “Examples” section.
