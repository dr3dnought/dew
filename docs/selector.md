# Selector Guide

`Selector` is the query builder for `SELECT` in Dew. Every schema has `From(db)` which returns `*Selector[T]`, and any selector can also be used as a subquery (it implements `Expression`).

## Creating a selector
```go
// From schema (preferred)
sel := UserSchema.From(db)

// Or generic
sel := dew.From[User](db, UserSchema)
```

## Selecting columns
```go
// All columns by default
var users []User
_ = UserSchema.From(db).All()

// Specific columns and aliases
userName := UserSchema.Name.As("user_name")
emails, _ := UserSchema.From(db).
	Select(UserSchema.ID, userName, UserSchema.Email).
	All()
```

## Filtering (`Where`)
Column predicates (work on all column types, where applicable):
- Equality/inequality: `Eq`, `Neq`
- Comparison: `Gt`, `Gte`, `Lt`, `Lte`
- Ranges: `Between(min, max)`
- Sets: `In`, `NotIn` (values), `InSub`, `NotInSub` (subqueries)
- Nullability: `IsNull`, `IsNotNull`
- Text: `Like`

Examples:
```go
// Combine multiple predicates (AND)
_ = UserSchema.From(db).
	Where(
		UserSchema.Name.Like("A%"),              // LIKE 'A%'
		UserSchema.Email.Like("%@example.com"),  // LIKE '%@example.com'
		UserSchema.ID.Gt(10),                    // id > 10
	).
	All()

// Between
_ = OrderSchema.From(db).
	Where(OrderSchema.Total.Between(100, 500)).
	All()

// In / NotIn
_ = UserSchema.From(db).
	Where(UserSchema.ID.In(1, 2, 3)).
	All()

// Combine OR/AND with helpers
_ = UserSchema.From(db).
	Where(
		dew.Or(
			UserSchema.Name.Like("A%"),
			UserSchema.Name.Like("B%"),
		),
		dew.And(
			UserSchema.Email.Like("%@example.com"),
			UserSchema.ID.Gt(10),
		),
	).
	All()
```

> `In()`/`NotIn()` with an empty slice become `1=0` / `1=1` to avoid invalid SQL.

## Ordering, pagination
```go
// Simple ordering
_ = UserSchema.From(db).
	OrderBy(dew.Desc(UserSchema.ID)).
	Limit(10).
	Offset(20).
	All()

// Multiple order clauses
_ = UserSchema.From(db).
	OrderBy(
		dew.Asc(UserSchema.Name),
		dew.Desc(UserSchema.ID),
	).
	All()

// Order by expression (e.g., aggregate alias)
totalAmount := dew.As(dew.Sum(OrderSchema.Total), "total_amount")
_ = UserSchema.From(db).
	Select(UserSchema.Name.As("user_name"), totalAmount).
	InnerJoin(OrderSchema, UserSchema.ID, OrderSchema.UserID).
	GroupBy(UserSchema.Name.As("user_name")).
	OrderBy(dew.Desc(totalAmount)).
	All()

// Reminder: First() sets LIMIT 1
first, _ := UserSchema.From(db).
	OrderBy(dew.Asc(UserSchema.ID)).
	First()
```

## Distinct
```go
// DISTINCT on all selected columns
_ = UserSchema.From(db).
	Distinct().
	Select(UserSchema.Email).
	All()

// DISTINCT on specific columns
_ = UserSchema.From(db).
	Distinct(UserSchema.Email, UserSchema.Name).
	Select(UserSchema.Email, UserSchema.Name).
	All()
```

## Joins
```go
type UserOrder struct {
	UserName string
	OrderID  int
	Total    float64
}

userName := UserSchema.Name.As("user_name")
orderID := OrderSchema.ID.As("order_id")

_ = UserSchema.From(db).
	Select(userName, orderID, OrderSchema.Total).
	InnerJoin(OrderSchema, UserSchema.ID, OrderSchema.UserID).
	OrderBy(dew.Asc(userName)).
	Scan(&[]UserOrder{})
```

## Subqueries (IN / NOT IN / comparisons)
Any selector is an `Expression`, so it can be plugged into predicates.

- `InSub(subSel)` / `NotInSub(subSel)` — column IN (SELECT ...)
- `EqSub(subSel)` / `NeqSub(subSel)` — column = (SELECT ...)

Examples:
```go
highValueOrders := OrderSchema.From(db).
	Select(OrderSchema.UserID).
	Where(OrderSchema.Total.Gt(150))

// IN (SELECT ...)
_ = UserSchema.From(db).
	Select(UserSchema.Name).
	Where(UserSchema.ID.InSub(highValueOrders)).
	Scan(&[]struct{ Name string }{})

// Equality to scalar subquery
maxOrderTotal := OrderSchema.From(db).
	Select(dew.Max(OrderSchema.Total)).
	Limit(1)

_ = OrderSchema.From(db).
	Where(OrderSchema.Total.EqSub(maxOrderTotal)).
	All()
```

## Grouping & aggregation
```go
totalOrders := dew.As(dew.CountDistinct(OrderSchema.ID), "total_orders")
totalAmount := dew.As(dew.Sum(OrderSchema.Total), "total_amount")
userName := UserSchema.Name.As("user_name")

var stats []struct {
	UserName    string
	TotalOrders int64
	TotalAmount float64
}

_ = UserSchema.From(db).
	Select(userName, totalOrders, totalAmount).
	InnerJoin(OrderSchema, UserSchema.ID, OrderSchema.UserID).
	GroupBy(userName).
	Having(totalAmount.Gt(100)). // expressions also work in HAVING
	OrderBy(dew.Desc(totalAmount)).
	Scan(&stats)
```

## Fetching results
- `All()` — returns slice `[]*T`
- `One()` — expects exactly one row
- `First()` — first row (LIMIT 1)
- `Scan(&slice)` / `Scan(&one)` — map columns to struct fields by name/alias

Example:
```go
type Row struct {
	Name  string
	Email string
}
var rows []Row
err := UserSchema.From(db).
	Select(UserSchema.Name, UserSchema.Email).
	Scan(&rows)
```

### Select into custom DTOs
You can project columns into ad-hoc structs. Field names must match column names (snake_case) or aliases.
```go
type UserEmail struct {
	UserName string // expects column/alias "user_name"
	Email    string // expects column/alias "email"
}

userName := UserSchema.Name.As("user_name")
var list []UserEmail
_ = UserSchema.From(db).
	Select(userName, UserSchema.Email).
	Scan(&list)
```

Aggregates and computed expressions should be aliased to match fields:
```go
type Stats struct {
	UserName    string
	TotalOrders int64
}

totalOrders := dew.As(dew.CountDistinct(OrderSchema.ID), "total_orders")
_ = UserSchema.From(db).
	Select(UserSchema.Name, totalOrders).
	InnerJoin(OrderSchema, UserSchema.ID, OrderSchema.UserID).
	GroupBy(UserSchema.Name).
	Scan(&[]Stats{})
```

## Existence and counts
```go
exists, _ := UserSchema.From(db).
	Where(UserSchema.Name.Eq("Alice")).
	Exists()

count, _ := UserSchema.From(db).
	Where(UserSchema.Email.Like("%@example.com")).
	Count()
```

## Cloning
Use `Clone()` to branch a selector without mutating the original (useful for Exists/Count variations).

## Raw SQL (sparingly)
`dew.Raw("...")` can be used inside `Where`/`Select`, but prefer typed builders to keep safety. Placeholders will be rewritten per dialect when you pass `Expression` values into `Where/OrderBy/Having`.

## Tips
- Always alias aggregate expressions via `dew.As` so struct fields map correctly.
- Placeholders are handled by the dialect; do not hardcode `?` or `$1`.
- For full-table scans with large data, set `Limit` or stream manually via `ScanWith`.

