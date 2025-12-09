# Examples

These snippets come from `cmd/dew-demo/main.go`. They run against Postgres; SQL shape is the same for other supported dialects.

## Insert batch
```go
_ = dew.Insert[User](db, UserSchema).
	Columns(UserSchema.Name, UserSchema.Email).
	Values("Alice", "alice@example.com").
	Values("Bob", "bob@example.com").
	Exec()
```

## Simple SELECT
```go
var results []struct {
	Name  string
	Email string
}
_ = UserSchema.From(db).
	Select(UserSchema.Name, UserSchema.Email).
	Where(UserSchema.Name.Eq("Alice")).
	Scan(&results)
```

## JOIN
```go
userName := UserSchema.Name.As("user_name")
orderID := OrderSchema.ID.As("order_id")

type UserOrder struct {
	UserName string
	OrderID  int
}

_ = UserSchema.From(db).
	Select(userName, orderID).
	InnerJoin(OrderSchema, UserSchema.ID, OrderSchema.UserID).
	OrderBy(dew.Asc(userName)).
	Scan(&[]UserOrder{})
```

## Subquery (IN)
```go
sub := OrderSchema.From(db).
	Select(OrderSchema.UserID).
	Where(OrderSchema.Total.Gt(150))

_ = UserSchema.From(db).
	Select(UserSchema.Name).
	Where(UserSchema.ID.InSub(sub)).
	Scan(&[]struct{ Name string }{})
```

## Aggregation + JOIN
```go
totalOrders := dew.As(dew.CountDistinct(OrderSchema.ID), "total_orders")
totalAmount := dew.As(dew.Sum(OrderSchema.Total), "total_amount")
userName := UserSchema.Name.As("user_name")

_ = UserSchema.From(db).
	Select(userName, totalOrders, totalAmount).
	InnerJoin(OrderSchema, UserSchema.ID, OrderSchema.UserID).
	GroupBy(userName).
	OrderBy(dew.Desc(totalAmount)).
	Scan(&[]struct {
		UserName    string
		TotalOrders int64
		TotalAmount float64
	}{})
```

## Delete with RowsAffected
```go
deleted, err := dew.Delete[Order](db, OrderSchema).
	Where(OrderSchema.Status.Eq("pending")).
	RowsAffected()
fmt.Println("deleted:", deleted)
```

