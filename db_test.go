package dew

import (
	"testing"
)

// Compile-time checks: *DB and *Tx both satisfy Querier.
var (
	_ Querier = (*DB)(nil)
	_ Querier = (*Tx)(nil)
)

func TestDB_getDialect(t *testing.T) {
	d := PostgreSQLDialect{}
	db := &DB{dialect: d}

	got := db.getDialect()
	if got != d {
		t.Errorf("DB.getDialect() = %T, want %T", got, d)
	}
}

func TestTx_getDialect(t *testing.T) {
	d := PostgreSQLDialect{}
	tx := &Tx{dialect: d}

	got := tx.getDialect()
	if got != d {
		t.Errorf("Tx.getDialect() = %T, want %T", got, d)
	}
}

// TestBuildersAcceptQuerier verifies that all builder constructors accept
// both *DB and *Tx through the Querier interface.
func TestBuildersAcceptQuerier(t *testing.T) {
	type User struct {
		ID   int    `col:"id"`
		Name string `col:"name"`
	}

	table := NewTable[User]("users", PostgreSQLDialect{})
	db := &DB{dialect: PostgreSQLDialect{}}
	tx := &Tx{dialect: PostgreSQLDialect{}}

	queriers := map[string]Querier{
		"DB": db,
		"Tx": tx,
	}

	for name, q := range queriers {
		t.Run(name, func(t *testing.T) {
			sel := table.From(q)
			if sel == nil {
				t.Fatal("Table.From returned nil")
			}

			ins := table.Insert(q)
			if ins == nil {
				t.Fatal("Table.Insert returned nil")
			}

			upd := table.Update(q)
			if upd == nil {
				t.Fatal("Table.Update returned nil")
			}

			del := table.Delete(q)
			if del == nil {
				t.Fatal("Table.Delete returned nil")
			}
		})
	}
}

// TestBuildersAcceptQuerier_Free verifies the free-standing constructor functions.
func TestBuildersAcceptQuerier_Free(t *testing.T) {
	type User struct {
		ID   int    `col:"id"`
		Name string `col:"name"`
	}

	table := NewTable[User]("users", PostgreSQLDialect{})
	tx := &Tx{dialect: PostgreSQLDialect{}}

	sel := From[User](tx, table)
	if sel == nil {
		t.Fatal("From returned nil")
	}

	ins := Insert[User](tx, table)
	if ins == nil {
		t.Fatal("Insert returned nil")
	}

	upd := Update[User](tx, table)
	if upd == nil {
		t.Fatal("Update returned nil")
	}

	del := Delete[User](tx, table)
	if del == nil {
		t.Fatal("Delete returned nil")
	}
}

// TestSelectorToSql_WithTx verifies query building works identically through *Tx.
func TestSelectorToSql_WithTx(t *testing.T) {
	type User struct {
		ID   int    `col:"id"`
		Name string `col:"name"`
	}

	table := NewTable[User]("users", PostgreSQLDialect{})

	db := &DB{dialect: PostgreSQLDialect{}}
	tx := &Tx{dialect: PostgreSQLDialect{}}

	nameCol := table.StringColumn("name")

	dbQuery, _ := table.From(db).Where(nameCol.Eq("Alice")).ToSql()
	txQuery, _ := table.From(tx).Where(nameCol.Eq("Alice")).ToSql()

	if dbQuery != txQuery {
		t.Errorf("queries differ:\n  DB: %s\n  Tx: %s", dbQuery, txQuery)
	}
}

// TestInserterToSql_WithTx verifies insert query building through *Tx.
func TestInserterToSql_WithTx(t *testing.T) {
	type User struct {
		ID   int    `col:"id"`
		Name string `col:"name"`
	}

	table := NewTable[User]("users", PostgreSQLDialect{})
	nameCol := table.StringColumn("name")

	db := &DB{dialect: PostgreSQLDialect{}}
	tx := &Tx{dialect: PostgreSQLDialect{}}

	dbQuery, _, dbErr := table.Insert(db).Columns(nameCol).Values("Alice").ToSql()
	txQuery, _, txErr := table.Insert(tx).Columns(nameCol).Values("Alice").ToSql()

	if dbErr != nil {
		t.Fatalf("DB insert build error: %v", dbErr)
	}
	if txErr != nil {
		t.Fatalf("Tx insert build error: %v", txErr)
	}
	if dbQuery != txQuery {
		t.Errorf("queries differ:\n  DB: %s\n  Tx: %s", dbQuery, txQuery)
	}
}

// TestUpdaterToSql_WithTx verifies update query building through *Tx.
func TestUpdaterToSql_WithTx(t *testing.T) {
	type User struct {
		ID   int    `col:"id"`
		Name string `col:"name"`
	}

	table := NewTable[User]("users", PostgreSQLDialect{})
	nameCol := table.StringColumn("name")
	idCol := table.IntColumn("id")

	db := &DB{dialect: PostgreSQLDialect{}}
	tx := &Tx{dialect: PostgreSQLDialect{}}

	dbQuery, _, dbErr := table.Update(db).Set(nameCol, "Bob").Where(idCol.Eq(1)).ToSql()
	txQuery, _, txErr := table.Update(tx).Set(nameCol, "Bob").Where(idCol.Eq(1)).ToSql()

	if dbErr != nil {
		t.Fatalf("DB update build error: %v", dbErr)
	}
	if txErr != nil {
		t.Fatalf("Tx update build error: %v", txErr)
	}
	if dbQuery != txQuery {
		t.Errorf("queries differ:\n  DB: %s\n  Tx: %s", dbQuery, txQuery)
	}
}

// TestDeleterToSql_WithTx verifies delete query building through *Tx.
func TestDeleterToSql_WithTx(t *testing.T) {
	type User struct {
		ID   int    `col:"id"`
		Name string `col:"name"`
	}

	table := NewTable[User]("users", PostgreSQLDialect{})
	idCol := table.IntColumn("id")

	db := &DB{dialect: PostgreSQLDialect{}}
	tx := &Tx{dialect: PostgreSQLDialect{}}

	dbQuery, _, dbErr := table.Delete(db).Where(idCol.Eq(1)).ToSql()
	txQuery, _, txErr := table.Delete(tx).Where(idCol.Eq(1)).ToSql()

	if dbErr != nil {
		t.Fatalf("DB delete build error: %v", dbErr)
	}
	if txErr != nil {
		t.Fatalf("Tx delete build error: %v", txErr)
	}
	if dbQuery != txQuery {
		t.Errorf("queries differ:\n  DB: %s\n  Tx: %s", dbQuery, txQuery)
	}
}

