package dew

type Tabler interface {
	TableName() string
}

type tableRef string

func (t tableRef) TableName() string { return string(t) }

// TableRef creates a Tabler reference for use with CTEs or other named sources.
func TableRef(name string) Tabler { return tableRef(name) }


type Table[T any] struct {
	name string
}

func NewTable[T any](name string, dialect Dialect) Table[T] {
	return Table[T]{name: name}
}

func (t Table[T]) TableName() string {
	return t.name
}

func (t Table[T]) From(db Querier) *Selector[T] {
	return &Selector[T]{
		db:        db,
		tableName: t.name,
	}
}

func (t Table[T]) Insert(db Querier) *Inserter[T] {
	return &Inserter[T]{
		db:    db,
		table: t,
	}
}

func (t Table[T]) Delete(db Querier) *Deleter[T] {
	return &Deleter[T]{
		db:    db,
		table: t,
	}
}

func (t Table[T]) Update(db Querier) *Updater[T] {
	return &Updater[T]{
		db:    db,
		table: t,
	}
}

func (t Table[T]) IntColumn(name string) IntColumn {
	return IntColumn{
		name:  name,
		table: &t.name,
	}
}

func (t Table[T]) Int64Column(name string) Int64Column {
	return Int64Column{
		name:  name,
		table: &t.name,
	}
}

func (t Table[T]) StringColumn(name string) StringColumn {
	return StringColumn{
		name:  name,
		table: &t.name,
	}
}

func (t Table[T]) BoolColumn(name string) BoolColumn {
	return BoolColumn{
		name:  name,
		table: &t.name,
	}
}

func (t Table[T]) FloatColumn(name string) FloatColumn {
	return FloatColumn{
		name:  name,
		table: &t.name,
	}
}

func (t Table[T]) TimeColumn(name string) TimeColumn {
	return TimeColumn{
		name:  name,
		table: &t.name,
	}
}

func (t Table[T]) BytesColumn(name string) BytesColumn {
	return BytesColumn{
		name:  name,
		table: &t.name,
	}
}

func (t Table[T]) Float32Column(name string) Float32Column {
	return Float32Column{
		name:  name,
		table: &t.name,
	}
}

func (t Table[T]) AnyColumn(name string) AnyColumn {
	return AnyColumn{
		name:  name,
		table: &t.name,
	}
}

func (t Table[T]) UUIDColumn(name string) UUIDColumn {
	return UUIDColumn{
		name:  name,
		table: &t.name,
	}
}

// JSONBColumn creates a typed JSONB column.
// Due to Go generics limitation, this is a standalone function, not a method.
// Usage: dew.JSONBColumn[MyType](t, "column_name")
func DefineJSONBColumn[V JSONB, T any](t Table[T], name string) JSONBColumn[V] {
	tableName := t.name
	return JSONBColumn[V]{
		name:  name,
		table: &tableName,
	}
}

func DefineSchema[T any, S any](tableName string, dialect Dialect, builder func(Table[T]) S) S {
	table := NewTable[T](tableName, dialect)
	return builder(table)
}
