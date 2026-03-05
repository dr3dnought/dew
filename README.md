# Dew

[![CI](https://github.com/dr3dnought/dew/actions/workflows/ci.yml/badge.svg)](https://github.com/dr3dnought/dew/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/dr3dnought/dew/branch/main/graph/badge.svg)](https://codecov.io/gh/dr3dnought/dew)
[![Go Reference](https://pkg.go.dev/badge/github.com/dr3dnought/dew.svg)](https://pkg.go.dev/github.com/dr3dnought/dew)
[![Go Report Card](https://goreportcard.com/badge/github.com/dr3dnought/dew)](https://goreportcard.com/report/github.com/dr3dnought/dew)

A lightweight, type-safe query builder for Go.

## Features

- Type-safe queries with builder pattern
- Support for PostgreSQL, SQLite, MySQL
- Automatic struct mapping
- SQL injection protection

## Installation

```bash
go get github.com/dr3dnought/dew
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/dr3dnought/dew"
    _ "github.com/mattn/go-sqlite3"
)

// Define model
type User struct {
    ID    int
    Name  string
    Email string
}

// Define schema
var UserSchema = dew.DefineSchema("users", func(t dew.Table[User]) struct {
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
    db, err := dew.Open("sqlite3", ":memory:")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    users, err := dew.From[User](db, UserSchema).
        Where(UserSchema.Name.Eq("Alice")).
        All()
}
```

## Documentation

Full documentation: https://dr3dnought.github.io/dew/

## License

MIT - see [LICENSE](LICENSE) file for details.

