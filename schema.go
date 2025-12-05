package dew

type Tabler interface {
	TableName() string
}

type Table[T any] struct {
	name string
}

func NewTable[T any](name string) Table[T] {
	return Table[T]{name: name}
}

func (t Table[T]) TableName() string {
	return t.name
}

func (t Table[T]) From(db *DB) *Selector[T] {
	return &Selector[T]{
		db:        db,
		tableName: t.name,
	}
}

func (t Table[T]) Insert(db *DB) *Insertor[T] {
	return &Insertor[T]{
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

func (t Table[T]) UUIDColumn(name string) UUIDColumn {
	return UUIDColumn{
		name:  name,
		table: &t.name,
	}
}

func DefineSchema[T any, S any](tableName string, builder func(Table[T]) S) S {
	table := NewTable[T](tableName)
	return builder(table)
}
