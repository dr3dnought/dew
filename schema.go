package dew

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

type Tabler interface {
	TableName() string
}
