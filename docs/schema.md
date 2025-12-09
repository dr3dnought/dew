# Schema Guide

Schemas connect Go structs to database tables and expose typed columns for builders. They also carry the dialect so placeholders are rendered correctly for the target DB.

## Defining a schema
Use `DefineSchema(tableName, dialect, builder)` to bind a Go type `T` to a table and define columns.

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

**Why dialect here?** The schema carries the dialect so all builders (`Select`, `Insert`, `Delete`) know how to emit placeholders (`$1` for Postgres, `?` for SQLite/MySQL, etc.).

## What a schema gives you
- Typed columns: `ID`, `Name`, `Email`, each with column-aware predicates (`Eq`, `Like`, `In`, `Between`, `InSub`, `IsNull`, etc.).
- Entry points:
  - `UserSchema.From(db)` → `*Selector[User]` for SELECT
  - `UserSchema.Insert(db)` → `*Insertor[User]` for INSERT
  - `UserSchema.Delete(db)` → `*Deletor[User]` for DELETE
- `TableName()` is derived from the `DefineSchema` name; no magic pluralization.

## Column helpers
Available on `dew.Table[T]` inside the builder:
- `IntColumn` — integers
- `StringColumn` — text/varchar
- `FloatColumn` — floating numeric types
- `BoolColumn` — boolean
- `TimeColumn` — time/time-with-zone values
- `UUIDColumn` — UUID as string

Each column supports predicates (`Eq`, `Neq`, `Gt/Gte`, `Lt/Lte`, `Between`, `In/NotIn`, `InSub/NotInSub`, `IsNull/IsNotNull`, `Like` for strings) and SQL generation with table qualification and optional aliasing via `.As("alias")`.

Example with aliases:
```go
userName := UserSchema.Name.As("user_name")
_ = UserSchema.From(db).
	Select(userName).
	Where(userName.Like("A%")).
	All()
```

## Multiple schemas
Define as many schemas as you need; reuse the same dialect for a given database.
```go
var OrderSchema = dew.DefineSchema("orders", dew.PostgreSQLDialect{}, func(t dew.Table[Order]) struct {
	dew.Table[Order]
	ID     dew.IntColumn
	UserID dew.IntColumn
	Total  dew.FloatColumn
	Status dew.StringColumn
} {
	return struct {
		dew.Table[Order]
		ID     dew.IntColumn
		UserID dew.IntColumn
		Total  dew.FloatColumn
		Status dew.StringColumn
	}{
		Table:  t,
		ID:     t.IntColumn("id"),
		UserID: t.IntColumn("user_id"),
		Total:  t.FloatColumn("total"),
		Status: t.StringColumn("status"),
	}
})
```

## Using schemas in queries
- Joins: `UserSchema.From(db).InnerJoin(OrderSchema, UserSchema.ID, OrderSchema.UserID)`
- Subqueries: any selector built from a schema can be used in `InSub`/`EqSub`.
- Aggregations: alias aggregates to align with DTO fields (`dew.As(dew.Sum(OrderSchema.Total), "total_amount")`).

## Safety notes
- Schema definition does **not** create tables; handle migrations separately.
- Column predicates are typed; mixing wrong types will surface at compile-time for most cases.
- Placeholders are dialect-aware; avoid hardcoding `?`/`$1`.

## When to use generic builders
You can also use `dew.From[T](db, schema)` or `dew.Insert[T](db, schema)`, but the schema methods (`From/Insert/Delete`) are the idiomatic entry points and ensure the correct table name and dialect are attached.

