# Dew

A lightweight, type-safe ORM library for Go.

## Features

- Type-safe queries
- Builder pattern API
- Support for PostgreSQL, SQLite, MySQL
- Automatic struct mapping
- SQL injection protection

## Quick Start

package main

import (
    "log"
    "github.com/dr3dnought/dew"
    _ "github.com/mattn/go-sqlite3"
)

// Define your model
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
    // Open database connection
    db, err := dew.Open("sqlite3", ":memory:")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Query users
    users, err := dew.From[User](db, UserSchema).
        Where(UserSchema.Name.Eq("Alice")).
        All()
    if err != nil {
        log.Fatal(err)
    }
}