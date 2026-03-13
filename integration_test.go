package dew

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// ─── Test helpers ────────────────────────────────────────────

type user struct {
	ID    int    `col:"id"`
	Name  string `col:"name"`
	Email string `col:"email"`
	Age   int    `col:"age"`
}

var (
	usersTable = NewTable[user]("users", SQLiteDialect{})
	uID        = usersTable.IntColumn("id")
	uName      = usersTable.StringColumn("name")
	uEmail     = usersTable.StringColumn("email")
	uAge       = usersTable.IntColumn("age")
)

func setupDB(t *testing.T) *DB {
	t.Helper()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db := NewDB(sqlDB, SQLiteDialect{})
	_, err = sqlDB.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		age INTEGER NOT NULL DEFAULT 0
	)`)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	return db
}

func seedUsers(t *testing.T, db *DB) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO users (name, email, age) VALUES ('Alice', 'alice@test.com', 30), ('Bob', 'bob@test.com', 25), ('Carol', 'carol@test.com', 35)")
	if err != nil {
		t.Fatal(err)
	}
}

// ─── Open ────────────────────────────────────────────────────

func TestOpen(t *testing.T) {
	db, err := Open("sqlite3", ":memory:", SQLiteDialect{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if db.getDialect() == nil {
		t.Fatal("dialect is nil")
	}
}

func TestOpen_WithErrorMapper(t *testing.T) {
	mapper := func(err error) error { return err }
	db, err := Open("sqlite3", ":memory:", SQLiteDialect{}, WithErrorMapper(mapper))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if db.errorMapper == nil {
		t.Fatal("errorMapper not set")
	}
}

// ─── Selector execution ─────────────────────────────────────

func TestSelector_All_Integration(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	results, err := usersTable.From(db).All()
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d, want 3", len(results))
	}
}

func TestSelector_All_WithContext(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	ctx := context.Background()
	results, err := usersTable.From(db).All(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d, want 3", len(results))
	}
}

func TestSelector_All_Where(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	results, err := usersTable.From(db).Where(uAge.Gt(28)).All()
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d, want 2", len(results))
	}
}

func TestSelector_All_Empty(t *testing.T) {
	db := setupDB(t)

	results, err := usersTable.From(db).All()
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("got %d, want 0", len(results))
	}
}

func TestSelector_One_Integration(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	u, err := usersTable.From(db).Where(uName.Eq("Alice")).One()
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "Alice" {
		t.Errorf("Name = %q, want Alice", u.Name)
	}
	if u.Age != 30 {
		t.Errorf("Age = %d, want 30", u.Age)
	}
}

func TestSelector_One_NotFound(t *testing.T) {
	db := setupDB(t)

	_, err := usersTable.From(db).Where(uName.Eq("Nobody")).One()
	if err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestSelector_First_Integration(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	u, err := usersTable.From(db).OrderBy(Asc(uAge)).First()
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "Bob" {
		t.Errorf("Name = %q, want Bob", u.Name)
	}
}

func TestSelector_First_NotFound(t *testing.T) {
	db := setupDB(t)

	_, err := usersTable.From(db).First()
	if err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestSelector_Count_Integration(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	count, err := usersTable.From(db).Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}

func TestSelector_Count_WithWhere(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	count, err := usersTable.From(db).Where(uAge.Gte(30)).Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestSelector_Exists_True(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	exists, err := usersTable.From(db).Where(uName.Eq("Alice")).Exists()
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("expected exists = true")
	}
}

func TestSelector_Exists_False(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	exists, err := usersTable.From(db).Where(uName.Eq("Nobody")).Exists()
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("expected exists = false")
	}
}

func TestSelector_ScanWith_Integration(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	results, err := usersTable.From(db).
		Where(uAge.Gt(28)).
		OrderBy(Asc(uAge)).
		ScanWith(func(rows *sql.Rows) (*user, error) {
			var u user
			err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Age)
			return &u, err
		})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d, want 2", len(results))
	}
	if results[0].Name != "Alice" {
		t.Errorf("first = %q, want Alice", results[0].Name)
	}
}

// ─── Selector Scan (reflection) ─────────────────────────────

func TestSelector_Scan_Struct(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	var u user
	err := usersTable.From(db).Where(uName.Eq("Bob")).Scan(&u)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "Bob" {
		t.Errorf("Name = %q, want Bob", u.Name)
	}
}

func TestSelector_Scan_Slice(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	var users []user
	err := usersTable.From(db).OrderBy(Asc(uAge)).Scan(&users)
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 3 {
		t.Fatalf("got %d, want 3", len(users))
	}
	if users[0].Name != "Bob" {
		t.Errorf("first = %q, want Bob", users[0].Name)
	}
}

func TestSelector_Scan_NoDest(t *testing.T) {
	db := setupDB(t)
	err := usersTable.From(db).Scan()
	if err == nil {
		t.Fatal("expected error for no dest")
	}
}

func TestSelector_Scan_NilDest(t *testing.T) {
	db := setupDB(t)
	err := usersTable.From(db).Scan((*user)(nil))
	if err == nil {
		t.Fatal("expected error for nil dest")
	}
}

func TestSelector_Scan_NotFound(t *testing.T) {
	db := setupDB(t)
	var u user
	err := usersTable.From(db).Where(uName.Eq("Nobody")).Scan(&u)
	if err != ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestSelector_ScanCtx(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	var u user
	err := usersTable.From(db).Where(uName.Eq("Alice")).ScanCtx(context.Background(), &u)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "Alice" {
		t.Errorf("Name = %q, want Alice", u.Name)
	}
}

// ─── RowScanner integration ─────────────────────────────────

type rowScannerUser struct {
	ID   int    `col:"id"`
	Name string `col:"name"`
}

func (u *rowScannerUser) ScanRow(rows *sql.Rows) error {
	return rows.Scan(&u.ID, &u.Name)
}

var rsTable = NewTable[rowScannerUser]("users", SQLiteDialect{})

func TestSelector_All_RowScanner(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	results, err := rsTable.From(db).Select(
		usersTable.IntColumn("id"),
		usersTable.StringColumn("name"),
	).All()
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d, want 3", len(results))
	}
	if results[0].Name == "" {
		t.Error("RowScanner didn't populate Name")
	}
}

func TestSelector_Scan_RowScanner_Slice(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	var users []rowScannerUser
	err := rsTable.From(db).Select(
		usersTable.IntColumn("id"),
		usersTable.StringColumn("name"),
	).Scan(&users)
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 3 {
		t.Fatalf("got %d, want 3", len(users))
	}
}

func TestSelector_Scan_RowScanner_One(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	var u rowScannerUser
	err := rsTable.From(db).Select(
		usersTable.IntColumn("id"),
		usersTable.StringColumn("name"),
	).Where(uName.Eq("Alice")).Scan(&u)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "Alice" {
		t.Errorf("Name = %q, want Alice", u.Name)
	}
}

// ─── Inserter execution ─────────────────────────────────────

func TestInserter_Exec_Integration(t *testing.T) {
	db := setupDB(t)

	err := usersTable.Insert(db).
		Columns(uName, uEmail, uAge).
		Values("Dave", "dave@test.com", 40).
		Exec()
	if err != nil {
		t.Fatal(err)
	}

	count, _ := usersTable.From(db).Count()
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestInserter_Exec_MultiRow(t *testing.T) {
	db := setupDB(t)

	err := usersTable.Insert(db).
		Columns(uName, uEmail, uAge).
		Values("Alice", "alice@test.com", 30).
		Values("Bob", "bob@test.com", 25).
		Exec()
	if err != nil {
		t.Fatal(err)
	}

	count, _ := usersTable.From(db).Count()
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestInserter_Exec_Models(t *testing.T) {
	db := setupDB(t)

	// Models inserts all fields including ID, so provide explicit IDs
	users := []*user{
		{ID: 1, Name: "Alice", Email: "alice@test.com", Age: 30},
		{ID: 2, Name: "Bob", Email: "bob@test.com", Age: 25},
	}

	err := usersTable.Insert(db).Models(users...).Exec()
	if err != nil {
		t.Fatal(err)
	}

	count, _ := usersTable.From(db).Count()
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestInserter_Batch_Integration(t *testing.T) {
	db := setupDB(t)

	err := usersTable.Insert(db).
		Columns(uName, uEmail, uAge).
		Values("Alice", "alice@test.com", 30).
		Values("Bob", "bob@test.com", 25).
		Values("Carol", "carol@test.com", 35).
		Batch(2).
		Exec()
	if err != nil {
		t.Fatal(err)
	}

	count, _ := usersTable.From(db).Count()
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}

func TestInserter_ScanWith_Integration(t *testing.T) {
	db := setupDB(t)

	results, err := usersTable.Insert(db).
		Columns(uName, uEmail, uAge).
		Values("Alice", "alice@test.com", 30).
		Returning(uID, uName).
		ScanWith(func(rows *sql.Rows) (*user, error) {
			var u user
			err := rows.Scan(&u.ID, &u.Name)
			return &u, err
		})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d, want 1", len(results))
	}
	if results[0].Name != "Alice" {
		t.Errorf("Name = %q, want Alice", results[0].Name)
	}
}

// ─── Updater execution ──────────────────────────────────────

func TestUpdater_Exec_Integration(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	err := usersTable.Update(db).
		Set(uName, "Alice Updated").
		Where(uName.Eq("Alice")).
		Exec()
	if err != nil {
		t.Fatal(err)
	}

	u, _ := usersTable.From(db).Where(uName.Eq("Alice Updated")).One()
	if u == nil {
		t.Fatal("updated user not found")
	}
}

func TestUpdater_RowsAffected_Integration(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	n, err := usersTable.Update(db).
		Set(uAge, 99).
		Where(uAge.Gte(30)).
		RowsAffected()
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("rowsAffected = %d, want 2", n)
	}
}

func TestUpdater_Scan_Integration(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	var id int
	var name, email string
	var age int
	err := usersTable.Update(db).
		Set(uName, "Alice v2").
		Where(uName.Eq("Alice")).
		Returning(uID, uName, uEmail, uAge).
		Scan(context.Background(), &id, &name, &email, &age)
	if err != nil {
		t.Fatal(err)
	}
	if name != "Alice v2" {
		t.Errorf("Name = %q, want Alice v2", name)
	}
}

func TestUpdater_ScanWith_Integration(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	results, err := usersTable.Update(db).
		Set(uAge, 99).
		Where(uAge.Gte(30)).
		Returning(uID, uName, uEmail, uAge).
		ScanWith(func(rows *sql.Rows) (*user, error) {
			var u user
			err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Age)
			return &u, err
		})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d, want 2", len(results))
	}
	if results[0].Age != 99 {
		t.Errorf("Age = %d, want 99", results[0].Age)
	}
}

// ─── Deleter execution ──────────────────────────────────────

func TestDeleter_Exec_Integration(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	err := usersTable.Delete(db).
		Where(uName.Eq("Bob")).
		Exec()
	if err != nil {
		t.Fatal(err)
	}

	count, _ := usersTable.From(db).Count()
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestDeleter_RowsAffected_Integration(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	n, err := usersTable.Delete(db).
		Where(uAge.Lt(35)).
		RowsAffected()
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("rowsAffected = %d, want 2", n)
	}
}

func TestDeleter_Scan_Integration(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	var id int
	var name, email string
	var age int
	err := usersTable.Delete(db).
		Where(uName.Eq("Alice")).
		Returning(uID, uName, uEmail, uAge).
		Scan(context.Background(), &id, &name, &email, &age)
	if err != nil {
		t.Fatal(err)
	}
	if name != "Alice" {
		t.Errorf("Name = %q, want Alice", name)
	}
}

func TestDeleter_ScanWith_Integration(t *testing.T) {
	db := setupDB(t)
	seedUsers(t, db)

	results, err := usersTable.Delete(db).
		Where(uAge.Gte(30)).
		Returning(uID, uName, uEmail, uAge).
		ScanWith(func(rows *sql.Rows) (*user, error) {
			var u user
			err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Age)
			return &u, err
		})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d, want 2", len(results))
	}
}

// ─── Transactions ────────────────────────────────────────────

func TestTx_Begin_Commit(t *testing.T) {
	db := setupDB(t)

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	err = usersTable.Insert(tx).
		Columns(uName, uEmail, uAge).
		Values("TxUser", "tx@test.com", 20).
		Exec()
	if err != nil {
		t.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}

	u, err := usersTable.From(db).Where(uName.Eq("TxUser")).One()
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "TxUser" {
		t.Errorf("Name = %q, want TxUser", u.Name)
	}
}

func TestTx_Begin_Rollback(t *testing.T) {
	db := setupDB(t)

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}

	err = usersTable.Insert(tx).
		Columns(uName, uEmail, uAge).
		Values("Ghost", "ghost@test.com", 0).
		Exec()
	if err != nil {
		t.Fatal(err)
	}

	err = tx.Rollback()
	if err != nil {
		t.Fatal(err)
	}

	count, _ := usersTable.From(db).Count()
	if count != 0 {
		t.Errorf("count = %d, want 0 after rollback", count)
	}
}

func TestTx_BeginTx(t *testing.T) {
	db := setupDB(t)

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	err = usersTable.Insert(tx).
		Columns(uName, uEmail, uAge).
		Values("TxUser", "tx@test.com", 20).
		Exec()
	if err != nil {
		t.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}

	count, _ := usersTable.From(db).Count()
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestTx_ErrorMapper_Inherited(t *testing.T) {
	called := false
	mapper := func(err error) error {
		called = true
		return err
	}

	sqlDB, _ := sql.Open("sqlite3", ":memory:")
	defer sqlDB.Close()
	db := NewDB(sqlDB, SQLiteDialect{}, WithErrorMapper(mapper))

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	if tx.errorMapper == nil {
		t.Fatal("tx should inherit errorMapper")
	}

	// Trigger mapError
	tx.mapError(sql.ErrNoRows)
	if !called {
		t.Fatal("mapper should have been called on tx")
	}
}

// ─── SetQuery execution ─────────────────────────────────────

// NOTE: SetQuery (UNION/INTERSECT/EXCEPT) wraps each SELECT in parens,
// which SQLite doesn't support. These are tested via ToSql in builder_test.go
// and work with PostgreSQL/MySQL. Skipping execution test for SQLite.

// ─── Error mapping integration ──────────────────────────────

func TestErrorMapper_Select_Integration(t *testing.T) {
	mapped := false
	mapper := func(err error) error {
		mapped = true
		return ErrNotFound
	}

	sqlDB, _ := sql.Open("sqlite3", ":memory:")
	defer sqlDB.Close()
	db := NewDB(sqlDB, SQLiteDialect{}, WithErrorMapper(mapper))

	// Query a non-existent table — should trigger error mapper
	_, err := usersTable.From(db).All()
	if err == nil {
		t.Fatal("expected error")
	}
	if !mapped {
		t.Fatal("error mapper should have been called")
	}
}

func TestErrorMapper_Insert_Integration(t *testing.T) {
	mapped := false
	mapper := func(err error) error {
		mapped = true
		return err
	}

	sqlDB, _ := sql.Open("sqlite3", ":memory:")
	defer sqlDB.Close()
	db := NewDB(sqlDB, SQLiteDialect{}, WithErrorMapper(mapper))

	// Insert into non-existent table
	err := usersTable.Insert(db).
		Columns(uName, uEmail, uAge).
		Values("Test", "test@test.com", 1).
		Exec()
	if err == nil {
		t.Fatal("expected error")
	}
	if !mapped {
		t.Fatal("error mapper should have been called")
	}
}
