# API Reference (short)

Core Dew concepts. This is a quick cheat sheet; check code/autocomplete for details.

## Dialects
- `dew.PostgreSQLDialect{}`
- `dew.SQLiteDialect{}`
- `dew.MySQLDialect{}`
- `dew.MSSQLDialect{}`

## Connection
```go
db, err := dew.Open(driver, dsn, dialect)
defer db.Close()
```

## Schema & tables
- `DefineSchema[T,S](tableName string, dialect Dialect, builder func(Table[T]) S) S`
- `Table[T]` methods: `IntColumn`, `StringColumn`, `FloatColumn`, `BoolColumn`, `TimeColumn`, `UUIDColumn`
- `Table[T].From(db)` → `*Selector[T]`
- `Table[T].Insert(db)` → `*Insertor[T]`
- `Table[T].Delete(db)` → `*Deletor[T]`

## Selector (SELECT)
- Built via `schema.From(db)` or `dew.From[T](db, schema)`
- Key methods: `Select`, `Where`, `Distinct`, `InnerJoin/LeftJoin/RightJoin`, `GroupBy`, `Having`, `OrderBy`, `Limit`, `Offset`
- Execution: `All()`, `First()`, `One()`, `Scan(&slice)`, `Exists()`, `Count()`
- Subquery: any `Selector` is an `Expression` and can be used in `InSub`, `NotInSub`, `EqSub`

## Insertor (INSERT)
- Built via `dew.Insert[T](db, schema)` or `schema.Insert(db)`
- Methods: `Columns(...)`, `Values(...)`, `Models(...)`, `Returning(...)`
- Conflicts: `OnConflict(cols...).DoNothing()` or `OnConflict(cols...).SetUpdate(col, val)`
- Execution: `Exec()`, `ToSql()`

## Deletor (DELETE)
- Built via `dew.Delete[T](db, schema)` or `schema.Delete(db)`
- Methods: `Where(...)`, `Returning(...)`
- Execution: `Exec()`, `RowsAffected()`, `Scan(dest...)`, `ToSql()`
- Requires `Where`; otherwise an error prevents full-table deletes.

## Expressions & predicates
- For columns: `Eq`, `Neq`, `Gt/Gte`, `Lt/Lte`, `In`, `NotIn`, `InSub`, `NotInSub`, `Between`, `IsNull/IsNotNull`, `Like`
- Aggregates: `Count`, `CountDistinct`, `Sum`, `Avg`, `Min`, `Max` — wrap in `dew.As(expr, "alias")` when you need named fields
- Ordering: `dew.Asc(col)`, `dew.Desc(col)`
- Raw SQL (sparingly): `dew.Raw("...")`

