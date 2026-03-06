package dew

import (
	"reflect"
	"testing"
)

// Shared test schema — no DB connection needed.

type testUser struct {
	ID    int    `col:"id"`
	Name  string `col:"name"`
	Email string `col:"email"`
	Age   int    `col:"age"`
}

var (
	testTable = NewTable[testUser]("users", PostgreSQLDialect{})
	tID       = testTable.IntColumn("id")
	tName     = testTable.StringColumn("name")
	tEmail    = testTable.StringColumn("email")
	tAge      = testTable.IntColumn("age")

	pgDB    = &DB{dialect: PostgreSQLDialect{}}
	myDB    = &DB{dialect: MySQLDialect{}}
	mssqlDB = &DB{dialect: MSSQLDialect{}}
)

// ─── Selector ────────────────────────────────────────────────

func TestSelector_BasicSelect(t *testing.T) {
	sql, _ := testTable.From(pgDB).ToSql()
	want := "SELECT * FROM users"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_SelectColumns(t *testing.T) {
	sql, _ := testTable.From(pgDB).Select(tID, tName).ToSql()
	want := "SELECT users.id, users.name FROM users"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_Where(t *testing.T) {
	tests := []struct {
		name     string
		db       Querier
		wantSQL  string
		wantArgs []any
	}{
		{
			name:     "PostgreSQL",
			db:       pgDB,
			wantSQL:  "SELECT * FROM users WHERE users.id = $1",
			wantArgs: []any{1},
		},
		{
			name:     "MySQL",
			db:       myDB,
			wantSQL:  "SELECT * FROM users WHERE users.id = ?",
			wantArgs: []any{1},
		},
		{
			name:     "MSSQL",
			db:       mssqlDB,
			wantSQL:  "SELECT * FROM users WHERE users.id = @p1",
			wantArgs: []any{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := testTable.From(tt.db).Where(tID.Eq(1)).ToSql()
			if sql != tt.wantSQL {
				t.Errorf("sql = %q, want %q", sql, tt.wantSQL)
			}
			if !reflect.DeepEqual(args, tt.wantArgs) {
				t.Errorf("args = %v, want %v", args, tt.wantArgs)
			}
		})
	}
}

func TestSelector_WhereMultiple(t *testing.T) {
	sql, args := testTable.From(pgDB).
		Where(tName.Eq("Alice"), tAge.Gt(18)).
		ToSql()

	wantSQL := "SELECT * FROM users WHERE users.name = $1 AND users.age > $2"
	wantArgs := []any{"Alice", 18}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_WhereAndOr(t *testing.T) {
	sql, args := testTable.From(pgDB).
		Where(Or(tName.Eq("Alice"), tName.Eq("Bob"))).
		ToSql()

	wantSQL := "SELECT * FROM users WHERE (users.name = $1 OR users.name = $2)"
	wantArgs := []any{"Alice", "Bob"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_Limit(t *testing.T) {
	sql, _ := testTable.From(pgDB).Limit(10).ToSql()
	want := "SELECT * FROM users LIMIT 10"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_Offset(t *testing.T) {
	sql, _ := testTable.From(pgDB).Limit(10).Offset(20).ToSql()
	want := "SELECT * FROM users LIMIT 10 OFFSET 20"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_OrderBy(t *testing.T) {
	sql, _ := testTable.From(pgDB).OrderBy(Desc(tAge), Asc(tName)).ToSql()
	want := "SELECT * FROM users ORDER BY users.age DESC, users.name ASC"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_Distinct(t *testing.T) {
	// Distinct with explicit empty slice triggers DISTINCT *
	sel := testTable.From(pgDB)
	sel.distinctColumns = []Column{} // explicitly empty, not nil
	sql, _ := sel.ToSql()
	want := "SELECT DISTINCT * FROM users"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_DistinctColumns(t *testing.T) {
	sql, _ := testTable.From(pgDB).Distinct(tName, tEmail).ToSql()
	want := "SELECT DISTINCT users.name, users.email FROM users"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_GroupBy(t *testing.T) {
	sql, _ := testTable.From(pgDB).
		Select(tName, Count()).
		GroupBy(tName).
		ToSql()
	want := "SELECT users.name, COUNT(*) FROM users GROUP BY users.name"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_Having(t *testing.T) {
	sql, args := testTable.From(pgDB).
		Select(tName, Count()).
		GroupBy(tName).
		Having(Raw("COUNT(*) > ?", 5)).
		ToSql()
	wantSQL := "SELECT users.name, COUNT(*) FROM users GROUP BY users.name HAVING COUNT(*) > $1"
	wantArgs := []any{5}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_InnerJoin(t *testing.T) {
	ordersTable := NewTable[testUser]("orders", PostgreSQLDialect{})
	orderUserID := ordersTable.IntColumn("user_id")

	sql, _ := testTable.From(pgDB).
		InnerJoin(ordersTable, tID, orderUserID).
		ToSql()

	want := "SELECT * FROM users INNER JOIN orders ON users.id = orders.user_id"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_LeftJoin(t *testing.T) {
	ordersTable := NewTable[testUser]("orders", PostgreSQLDialect{})
	orderUserID := ordersTable.IntColumn("user_id")

	sql, _ := testTable.From(pgDB).
		LeftJoin(ordersTable, tID, orderUserID).
		ToSql()

	want := "SELECT * FROM users LEFT JOIN orders ON users.id = orders.user_id"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_Complex(t *testing.T) {
	sql, args := testTable.From(pgDB).
		Select(tName, tAge).
		Where(tAge.Gte(18), tName.NotEq("Admin")).
		OrderBy(Asc(tName)).
		Limit(50).
		Offset(10).
		ToSql()

	wantSQL := "SELECT users.name, users.age FROM users WHERE users.age >= $1 AND users.name != $2 ORDER BY users.name ASC LIMIT 50 OFFSET 10"
	wantArgs := []any{18, "Admin"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_SubqueryExpr(t *testing.T) {
	sql, args := testTable.From(pgDB).
		Where(tID.InSub(
			testTable.From(pgDB).Select(tID).Where(tAge.Gt(21)),
		)).
		ToSql()

	wantSQL := "SELECT * FROM users WHERE users.id IN (SELECT users.id FROM users WHERE users.age > $1)"
	wantArgs := []any{21}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_Clone(t *testing.T) {
	base := testTable.From(pgDB).Where(tAge.Gt(18))
	clone := base.Clone()

	// mutate clone
	clone.Where(tName.Eq("Alice"))

	baseSql, baseArgs := base.ToSql()
	cloneSql, cloneArgs := clone.ToSql()

	if baseSql == cloneSql {
		t.Errorf("clone mutation affected base: both = %q", baseSql)
	}
	if len(baseArgs) == len(cloneArgs) && len(cloneArgs) > 1 {
		t.Errorf("clone args leaked to base")
	}
}

// ─── Insertor ────────────────────────────────────────────────

func TestInsertor_SingleRow(t *testing.T) {
	sql, args, err := testTable.Insert(pgDB).
		Columns(tName, tEmail).
		Values("Alice", "alice@test.com").
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "INSERT INTO users (name, email) VALUES ($1, $2)"
	wantArgs := []any{"Alice", "alice@test.com"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestInsertor_MultiRow(t *testing.T) {
	sql, args, err := testTable.Insert(pgDB).
		Columns(tName, tEmail).
		Values("Alice", "a@test.com").
		Values("Bob", "b@test.com").
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "INSERT INTO users (name, email) VALUES ($1, $2), ($3, $4)"
	wantArgs := []any{"Alice", "a@test.com", "Bob", "b@test.com"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestInsertor_MySQL(t *testing.T) {
	sql, args, err := testTable.Insert(myDB).
		Columns(tName).
		Values("Alice").
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "INSERT INTO users (name) VALUES (?)"
	wantArgs := []any{"Alice"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestInsertor_Models(t *testing.T) {
	u := &testUser{ID: 1, Name: "Alice", Email: "a@test.com", Age: 30}
	sql, args, err := testTable.Insert(pgDB).Models(u).ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "INSERT INTO users (id, name, email, age) VALUES ($1, $2, $3, $4)"
	wantArgs := []any{1, "Alice", "a@test.com", 30}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestInsertor_Returning(t *testing.T) {
	sql, _, err := testTable.Insert(pgDB).
		Columns(tName).
		Values("Alice").
		Returning(tID).
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "INSERT INTO users (name) VALUES ($1) RETURNING users.id"
	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
}

func TestInsertor_OnConflictDoNothing(t *testing.T) {
	sql, _, err := testTable.Insert(pgDB).
		Columns(tName, tEmail).
		Values("Alice", "a@test.com").
		OnConflict(tEmail).
		DoNothing().
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "INSERT INTO users (name, email) VALUES ($1, $2) ON CONFLICT (email) DO NOTHING"
	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
}

func TestInsertor_OnConflictDoUpdate(t *testing.T) {
	sql, args, err := testTable.Insert(pgDB).
		Columns(tName, tEmail).
		Values("Alice", "a@test.com").
		OnConflict(tEmail).
		SetUpdate(tName, "Alice Updated").
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantArgs := []any{"Alice", "a@test.com", "Alice Updated"}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}

	// SQL should contain ON CONFLICT ... DO UPDATE SET
	if !contains(sql, "ON CONFLICT (email) DO UPDATE SET") {
		t.Errorf("sql missing conflict clause: %q", sql)
	}
	if !contains(sql, "name = $3") {
		t.Errorf("sql missing set clause: %q", sql)
	}
}

func TestInsertor_ErrorNoData(t *testing.T) {
	_, _, err := testTable.Insert(pgDB).ToSql()
	if err == nil {
		t.Fatal("expected error for insert with no data")
	}
}

func TestInsertor_ErrorColumnsWithModels(t *testing.T) {
	u := &testUser{Name: "Alice"}
	_, _, err := testTable.Insert(pgDB).Columns(tName).Models(u).ToSql()
	if err == nil {
		t.Fatal("expected error for Columns + Models")
	}
}

func TestInsertor_ErrorValuesAndModels(t *testing.T) {
	u := &testUser{Name: "Alice"}
	_, _, err := testTable.Insert(pgDB).
		Columns(tName).
		Values("Alice").
		Models(u).
		ToSql()
	if err == nil {
		t.Fatal("expected error for Values + Models")
	}
}

func TestInsertor_ErrorMismatchedValues(t *testing.T) {
	_, _, err := testTable.Insert(pgDB).
		Columns(tName, tEmail).
		Values("Alice"). // only 1 value for 2 columns
		ToSql()
	if err == nil {
		t.Fatal("expected error for mismatched values/columns")
	}
}

// ─── Updator ─────────────────────────────────────────────────

func TestUpdator_Basic(t *testing.T) {
	sql, args, err := testTable.Update(pgDB).
		Set(tName, "Bob").
		Where(tID.Eq(1)).
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "UPDATE users SET name = $1 WHERE users.id = $2"
	wantArgs := []any{"Bob", 1}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestUpdator_MultipleSet(t *testing.T) {
	sql, args, err := testTable.Update(pgDB).
		Set(tName, "Bob").
		Set(tAge, 25).
		Where(tID.Eq(1)).
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "UPDATE users SET name = $1, age = $2 WHERE users.id = $3"
	wantArgs := []any{"Bob", 25, 1}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestUpdator_SetExpression(t *testing.T) {
	sql, args, err := testTable.Update(pgDB).
		Set(tAge, Raw("age + ?", 1)).
		Where(tID.Eq(1)).
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "UPDATE users SET age = age + $1 WHERE users.id = $2"
	wantArgs := []any{1, 1}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestUpdator_Returning(t *testing.T) {
	sql, _, err := testTable.Update(pgDB).
		Set(tName, "Bob").
		Where(tID.Eq(1)).
		Returning(tID, tName).
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "UPDATE users SET name = $1 WHERE users.id = $2 RETURNING id, name"
	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
}

func TestUpdator_MySQL(t *testing.T) {
	sql, args, err := testTable.Update(myDB).
		Set(tName, "Bob").
		Where(tID.Eq(1)).
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "UPDATE users SET name = ? WHERE users.id = ?"
	wantArgs := []any{"Bob", 1}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestUpdator_ErrorNoSet(t *testing.T) {
	_, _, err := testTable.Update(pgDB).Where(tID.Eq(1)).ToSql()
	if err == nil {
		t.Fatal("expected error for update with no Set")
	}
}

func TestUpdator_ErrorNoWhere(t *testing.T) {
	_, _, err := testTable.Update(pgDB).Set(tName, "Bob").ToSql()
	if err == nil {
		t.Fatal("expected error for update with no Where")
	}
}

func TestUpdator_Clone(t *testing.T) {
	base := testTable.Update(pgDB).Set(tName, "Bob").Where(tID.Eq(1))
	clone := base.Clone()
	clone.Set(tAge, 30)

	baseSql, _, _ := base.ToSql()
	cloneSql, _, _ := clone.ToSql()

	if baseSql == cloneSql {
		t.Errorf("clone mutation affected base: both = %q", baseSql)
	}
}

// ─── Deletor ─────────────────────────────────────────────────

func TestDeletor_Basic(t *testing.T) {
	sql, args, err := testTable.Delete(pgDB).
		Where(tID.Eq(1)).
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "DELETE FROM users WHERE users.id = $1"
	wantArgs := []any{1}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestDeletor_MultipleWhere(t *testing.T) {
	sql, args, err := testTable.Delete(pgDB).
		Where(tName.Eq("Alice"), tAge.Lt(18)).
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "DELETE FROM users WHERE users.name = $1 AND users.age < $2"
	wantArgs := []any{"Alice", 18}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestDeletor_Returning(t *testing.T) {
	sql, _, err := testTable.Delete(pgDB).
		Where(tID.Eq(1)).
		Returning(tID, tName).
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "DELETE FROM users WHERE users.id = $1 RETURNING id, name"
	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
}

func TestDeletor_MySQL(t *testing.T) {
	sql, args, err := testTable.Delete(myDB).
		Where(tID.Eq(1)).
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "DELETE FROM users WHERE users.id = ?"
	wantArgs := []any{1}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestDeletor_ForceDeleteAll(t *testing.T) {
	sql, _, err := testTable.Delete(pgDB).
		Where(Raw("1=1")).
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "DELETE FROM users WHERE 1=1"
	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
}

func TestDeletor_ErrorNoWhere(t *testing.T) {
	_, _, err := testTable.Delete(pgDB).ToSql()
	if err == nil {
		t.Fatal("expected error for delete with no Where")
	}
}

func TestDeletor_Clone(t *testing.T) {
	base := testTable.Delete(pgDB).Where(tID.Eq(1))
	clone := base.Clone()
	clone.Where(tName.Eq("Alice"))

	baseSql, _, _ := base.ToSql()
	cloneSql, _, _ := clone.ToSql()

	if baseSql == cloneSql {
		t.Errorf("clone mutation affected base: both = %q", baseSql)
	}
}

// ─── CTE ─────────────────────────────────────────────────────

func TestSelector_CTE_Basic(t *testing.T) {
	sub := testTable.From(pgDB).Select(tID, tName).Where(tAge.Gt(18))
	sql, args := From[testUser](pgDB, TableRef("adults")).
		With(CTE("adults", sub)).
		Where(Raw("adults.name = ?", "Alice")).
		ToSql()

	wantSQL := "WITH adults AS (SELECT users.id, users.name FROM users WHERE users.age > $1) SELECT * FROM adults WHERE adults.name = $2"
	wantArgs := []any{18, "Alice"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_CTE_Recursive(t *testing.T) {
	base := Raw("SELECT 1 AS n")
	recursive := Raw("SELECT n + 1 FROM nums WHERE n < ?", 10)
	cteBody := Union[struct{ N int }](pgDB, base, recursive)

	sql, args := From[struct{ N int }](pgDB, TableRef("nums")).
		With(RecursiveCTE("nums", cteBody)).
		ToSql()

	wantSQL := "WITH RECURSIVE nums AS ((SELECT 1 AS n) UNION (SELECT n + 1 FROM nums WHERE n < $1)) SELECT * FROM nums"
	wantArgs := []any{10}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_CTE_MultipleArgs(t *testing.T) {
	cte1 := testTable.From(pgDB).Where(tAge.Gt(18))
	cte2 := testTable.From(pgDB).Where(tName.Eq("Bob"))

	sql, args := From[testUser](pgDB, TableRef("a")).
		With(CTE("a", cte1), CTE("b", cte2)).
		Where(Raw("1=1")).
		ToSql()

	wantSQL := "WITH a AS (SELECT * FROM users WHERE users.age > $1), b AS (SELECT * FROM users WHERE users.name = $2) SELECT * FROM a WHERE 1=1"
	wantArgs := []any{18, "Bob"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

// ─── Subquery FROM ───────────────────────────────────────────

func TestSelector_FromSub(t *testing.T) {
	sub := testTable.From(pgDB).Select(tName, tAge).Where(tAge.Gt(21))
	sql, args := FromSub[testUser](pgDB, sub, "sub").
		Where(Raw("sub.age < ?", 30)).
		ToSql()

	wantSQL := "SELECT * FROM (SELECT users.name, users.age FROM users WHERE users.age > $1) AS sub WHERE sub.age < $2"
	wantArgs := []any{21, 30}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

// ─── UNION / INTERSECT / EXCEPT ──────────────────────────────

func TestSetQuery_Union(t *testing.T) {
	left := testTable.From(pgDB).Select(tName).Where(tAge.Gt(18))
	right := testTable.From(pgDB).Select(tName).Where(tName.Eq("Admin"))

	sql, args := Union[testUser](pgDB, left, right).ToSql()

	wantSQL := "(SELECT users.name FROM users WHERE users.age > $1) UNION (SELECT users.name FROM users WHERE users.name = $2)"
	wantArgs := []any{18, "Admin"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSetQuery_UnionAll(t *testing.T) {
	left := testTable.From(pgDB).Select(tName).Where(tAge.Gt(18))
	right := testTable.From(pgDB).Select(tName).Where(tAge.Lt(5))

	sql, args := UnionAll[testUser](pgDB, left, right).ToSql()

	wantSQL := "(SELECT users.name FROM users WHERE users.age > $1) UNION ALL (SELECT users.name FROM users WHERE users.age < $2)"
	wantArgs := []any{18, 5}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSetQuery_Intersect(t *testing.T) {
	left := testTable.From(pgDB).Select(tName).Where(tAge.Gt(18))
	right := testTable.From(pgDB).Select(tName).Where(tName.Eq("Alice"))

	sql, args := Intersect[testUser](pgDB, left, right).ToSql()

	wantSQL := "(SELECT users.name FROM users WHERE users.age > $1) INTERSECT (SELECT users.name FROM users WHERE users.name = $2)"
	wantArgs := []any{18, "Alice"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSetQuery_Except(t *testing.T) {
	left := testTable.From(pgDB).Select(tName).Where(tAge.Gt(18))
	right := testTable.From(pgDB).Select(tName).Where(tName.Eq("Admin"))

	sql, args := Except[testUser](pgDB, left, right).ToSql()

	wantSQL := "(SELECT users.name FROM users WHERE users.age > $1) EXCEPT (SELECT users.name FROM users WHERE users.name = $2)"
	wantArgs := []any{18, "Admin"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSetQuery_UnionChain(t *testing.T) {
	q1 := testTable.From(pgDB).Select(tName).Where(tAge.Gt(18))
	q2 := testTable.From(pgDB).Select(tName).Where(tAge.Lt(5))
	q3 := testTable.From(pgDB).Select(tName).Where(tName.Eq("Admin"))

	sql, args := Union[testUser](pgDB, q1, q2).UnionAll(q3).ToSql()

	wantSQL := "(SELECT users.name FROM users WHERE users.age > $1) UNION (SELECT users.name FROM users WHERE users.age < $2) UNION ALL (SELECT users.name FROM users WHERE users.name = $3)"
	wantArgs := []any{18, 5, "Admin"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSetQuery_OrderByLimitOffset(t *testing.T) {
	left := testTable.From(pgDB).Select(tName).Where(tAge.Gt(18))
	right := testTable.From(pgDB).Select(tName).Where(tAge.Lt(5))

	sql, args := UnionAll[testUser](pgDB, left, right).
		OrderBy(Asc(Raw("name"))).
		Limit(10).
		Offset(5).
		ToSql()

	wantSQL := "(SELECT users.name FROM users WHERE users.age > $1) UNION ALL (SELECT users.name FROM users WHERE users.age < $2) ORDER BY name ASC LIMIT 10 OFFSET 5"
	wantArgs := []any{18, 5}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSetQuery_AsCTEBody(t *testing.T) {
	left := testTable.From(pgDB).Select(tName, tAge).Where(tAge.Gt(18))
	right := testTable.From(pgDB).Select(tName, tAge).Where(tName.Eq("Admin"))
	combined := Union[testUser](pgDB, left, right)

	sql, args := From[testUser](pgDB, TableRef("combined")).
		With(CTE("combined", combined)).
		Where(Raw("combined.age > ?", 25)).
		ToSql()

	wantSQL := "WITH combined AS ((SELECT users.name, users.age FROM users WHERE users.age > $1) UNION (SELECT users.name, users.age FROM users WHERE users.name = $2)) SELECT * FROM combined WHERE combined.age > $3"
	wantArgs := []any{18, "Admin", 25}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_CTE_NoArgs(t *testing.T) {
	sub := testTable.From(pgDB).Select(tID, tName)
	sql, args := From[testUser](pgDB, TableRef("all_users")).
		With(CTE("all_users", sub)).
		ToSql()

	wantSQL := "WITH all_users AS (SELECT users.id, users.name FROM users) SELECT * FROM all_users"

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if len(args) != 0 {
		t.Errorf("args = %v, want empty", args)
	}
}

func TestSetQuery_MySQL(t *testing.T) {
	left := testTable.From(myDB).Select(tName).Where(tAge.Gt(18))
	right := testTable.From(myDB).Select(tName).Where(tName.Eq("Admin"))

	sql, args := Union[testUser](myDB, left, right).ToSql()

	wantSQL := "(SELECT users.name FROM users WHERE users.age > ?) UNION (SELECT users.name FROM users WHERE users.name = ?)"
	wantArgs := []any{18, "Admin"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

// ─── helpers ─────────────────────────────────────────────────

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
