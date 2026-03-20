package dew

import (
	"database/sql"
	"errors"
	"fmt"
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
	sql, _, _ := testTable.From(pgDB).ToSql()
	want := "SELECT * FROM users"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_SelectColumns(t *testing.T) {
	sql, _, _ := testTable.From(pgDB).Select(tID, tName).ToSql()
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
			sql, args, _ := testTable.From(tt.db).Where(tID.Eq(1)).ToSql()
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
	sql, args, _ := testTable.From(pgDB).
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
	sql, args, _ := testTable.From(pgDB).
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
	sql, _, _ := testTable.From(pgDB).Limit(10).ToSql()
	want := "SELECT * FROM users LIMIT 10"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_Offset(t *testing.T) {
	sql, _, _ := testTable.From(pgDB).Limit(10).Offset(20).ToSql()
	want := "SELECT * FROM users LIMIT 10 OFFSET 20"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_OrderBy(t *testing.T) {
	sql, _, _ := testTable.From(pgDB).OrderBy(Desc(tAge), Asc(tName)).ToSql()
	want := "SELECT * FROM users ORDER BY users.age DESC, users.name ASC"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_Distinct(t *testing.T) {
	// Distinct with explicit empty slice triggers DISTINCT *
	sel := testTable.From(pgDB)
	sel.distinctColumns = []Column{} // explicitly empty, not nil
	sql, _, _ := sel.ToSql()
	want := "SELECT DISTINCT * FROM users"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_DistinctColumns(t *testing.T) {
	sql, _, _ := testTable.From(pgDB).Distinct(tName, tEmail).ToSql()
	want := "SELECT DISTINCT users.name, users.email FROM users"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_GroupBy(t *testing.T) {
	sql, _, _ := testTable.From(pgDB).
		Select(tName, Count()).
		GroupBy(tName).
		ToSql()
	want := "SELECT users.name, COUNT(*) FROM users GROUP BY users.name"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_Having(t *testing.T) {
	sql, args, _ := testTable.From(pgDB).
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

	sql, _, _ := testTable.From(pgDB).
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

	sql, _, _ := testTable.From(pgDB).
		LeftJoin(ordersTable, tID, orderUserID).
		ToSql()

	want := "SELECT * FROM users LEFT JOIN orders ON users.id = orders.user_id"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_Complex(t *testing.T) {
	sql, args, _ := testTable.From(pgDB).
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
	sql, args, _ := testTable.From(pgDB).
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

	baseSql, baseArgs, _ := base.ToSql()
	cloneSql, cloneArgs, _ := clone.ToSql()

	if baseSql == cloneSql {
		t.Errorf("clone mutation affected base: both = %q", baseSql)
	}
	if len(baseArgs) == len(cloneArgs) && len(cloneArgs) > 1 {
		t.Errorf("clone args leaked to base")
	}
}

func TestSelector_RightJoin(t *testing.T) {
	ordersTable := NewTable[testUser]("orders", PostgreSQLDialect{})
	orderUserID := ordersTable.IntColumn("user_id")

	sql, _, _ := testTable.From(pgDB).
		RightJoin(ordersTable, tID, orderUserID).
		ToSql()

	want := "SELECT * FROM users RIGHT JOIN orders ON users.id = orders.user_id"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_MultipleJoins(t *testing.T) {
	ordersTable := NewTable[testUser]("orders", PostgreSQLDialect{})
	orderUserID := ordersTable.IntColumn("user_id")
	paymentsTable := NewTable[testUser]("payments", PostgreSQLDialect{})
	paymentOrderID := paymentsTable.IntColumn("order_id")
	orderID := ordersTable.IntColumn("id")

	sql, _, _ := testTable.From(pgDB).
		InnerJoin(ordersTable, tID, orderUserID).
		LeftJoin(paymentsTable, orderID, paymentOrderID).
		ToSql()

	want := "SELECT * FROM users INNER JOIN orders ON users.id = orders.user_id LEFT JOIN payments ON orders.id = payments.order_id"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_DistinctWithWhere(t *testing.T) {
	sql, args, _ := testTable.From(pgDB).
		Distinct(tName).
		Where(tAge.Gt(18)).
		ToSql()

	wantSQL := "SELECT DISTINCT users.name FROM users WHERE users.age > $1"
	wantArgs := []any{18}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_ScanWith_ToSql(t *testing.T) {
	sql, args, _ := testTable.From(pgDB).
		Select(tName, tEmail).
		Where(tAge.Gt(21)).
		ToSql()

	wantSQL := "SELECT users.name, users.email FROM users WHERE users.age > $1"
	wantArgs := []any{21}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_MSSQL(t *testing.T) {
	sql, args, _ := testTable.From(mssqlDB).
		Select(tName).
		Where(tAge.Gt(18), tName.NotEq("Admin")).
		OrderBy(Asc(tName)).
		Limit(10).
		ToSql()

	wantSQL := "SELECT users.name FROM users WHERE users.age > @p1 AND users.name != @p2 ORDER BY users.name ASC LIMIT 10"
	wantArgs := []any{18, "Admin"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_CloneWithCTE(t *testing.T) {
	sub := testTable.From(pgDB).Where(tAge.Gt(18))
	base := From[testUser](pgDB, TableRef("adults")).
		With(CTE("adults", sub))

	clone := base.Clone()
	clone.Where(tName.Eq("Alice"))

	baseSql, baseArgs, _ := base.ToSql()
	cloneSql, cloneArgs, _ := clone.ToSql()

	wantBase := "WITH adults AS (SELECT * FROM users WHERE users.age > $1) SELECT * FROM adults"
	if baseSql != wantBase {
		t.Errorf("base sql = %q, want %q", baseSql, wantBase)
	}
	if len(baseArgs) != 1 {
		t.Errorf("base args = %v, want 1 arg", baseArgs)
	}

	wantClone := "WITH adults AS (SELECT * FROM users WHERE users.age > $1) SELECT * FROM adults WHERE users.name = $2"
	if cloneSql != wantClone {
		t.Errorf("clone sql = %q, want %q", cloneSql, wantClone)
	}
	if len(cloneArgs) != 2 {
		t.Errorf("clone args = %v, want 2 args", cloneArgs)
	}
}

func TestSelector_CloneWithFromSub(t *testing.T) {
	sub := testTable.From(pgDB).Select(tName, tAge).Where(tAge.Gt(21))
	base := FromSub[testUser](pgDB, sub, "sub")

	clone := base.Clone()
	clone.Where(Raw("sub.age < ?", 30))

	baseSql, baseArgs, _ := base.ToSql()
	cloneSql, cloneArgs, _ := clone.ToSql()

	wantBase := "SELECT * FROM (SELECT users.name, users.age FROM users WHERE users.age > $1) AS sub"
	if baseSql != wantBase {
		t.Errorf("base sql = %q, want %q", baseSql, wantBase)
	}
	if len(baseArgs) != 1 {
		t.Errorf("base args = %v, want 1 arg", baseArgs)
	}

	wantClone := "SELECT * FROM (SELECT users.name, users.age FROM users WHERE users.age > $1) AS sub WHERE sub.age < $2"
	if cloneSql != wantClone {
		t.Errorf("clone sql = %q, want %q", cloneSql, wantClone)
	}
	if len(cloneArgs) != 2 {
		t.Errorf("clone args = %v, want 2 args", cloneArgs)
	}
}

func TestSelector_SetQueryAsFromSub(t *testing.T) {
	left := testTable.From(pgDB).Select(tName).Where(tAge.Gt(18))
	right := testTable.From(pgDB).Select(tName).Where(tName.Eq("Admin"))
	combined := Union[testUser](pgDB, left, right)

	sql, args, _ := FromSub[testUser](pgDB, combined, "staff").
		Where(Raw("1=1")).
		ToSql()

	wantSQL := "SELECT * FROM ((SELECT users.name FROM users WHERE users.age > $1) UNION (SELECT users.name FROM users WHERE users.name = $2)) AS staff WHERE 1=1"
	wantArgs := []any{18, "Admin"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_WhereEmpty(t *testing.T) {
	// Empty expression should be skipped
	sql, _, _ := testTable.From(pgDB).
		Where(Raw("")).
		ToSql()

	want := "SELECT * FROM users"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_LimitOffset(t *testing.T) {
	sql, _, _ := testTable.From(pgDB).Offset(5).ToSql()
	// offset without limit
	want := "SELECT * FROM users OFFSET 5"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_SelectWithAlias(t *testing.T) {
	sql, _, _ := testTable.From(pgDB).
		Select(tName, As(Count(), "total")).
		GroupBy(tName).
		ToSql()

	want := "SELECT users.name, COUNT(*) AS total FROM users GROUP BY users.name"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_NestedAndOr(t *testing.T) {
	sql, args, _ := testTable.From(pgDB).
		Where(And(
			tAge.Gte(18),
			Or(
				tName.Eq("Alice"),
				tName.Eq("Bob"),
			),
		)).
		ToSql()

	wantSQL := "SELECT * FROM users WHERE (users.age >= $1 AND (users.name = $2 OR users.name = $3))"
	wantArgs := []any{18, "Alice", "Bob"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_HavingMultiple(t *testing.T) {
	sql, args, _ := testTable.From(pgDB).
		Select(tName, Count()).
		GroupBy(tName).
		Having(Raw("COUNT(*) > ?", 5), Raw("COUNT(*) < ?", 100)).
		ToSql()

	wantSQL := "SELECT users.name, COUNT(*) FROM users GROUP BY users.name HAVING COUNT(*) > $1 AND COUNT(*) < $2"
	wantArgs := []any{5, 100}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_HavingWithAggComparison(t *testing.T) {
	sql, args, _ := testTable.From(pgDB).
		Select(tName, Count()).
		GroupBy(tName).
		Having(Count().Gt(5), Sum(tAge).Lte(100)).
		ToSql()

	wantSQL := "SELECT users.name, COUNT(*) FROM users GROUP BY users.name HAVING COUNT(*) > $1 AND SUM(users.age) <= $2"
	wantArgs := []any{5, 100}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestSelector_GroupByMultiple(t *testing.T) {
	sql, _, _ := testTable.From(pgDB).
		Select(tName, tAge, Count()).
		GroupBy(tName, tAge).
		ToSql()

	want := "SELECT users.name, users.age, COUNT(*) FROM users GROUP BY users.name, users.age"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_JoinWithWhere(t *testing.T) {
	ordersTable := NewTable[testUser]("orders", PostgreSQLDialect{})
	orderUserID := ordersTable.IntColumn("user_id")
	orderTotal := ordersTable.FloatColumn("total")

	sql, args, _ := testTable.From(pgDB).
		Select(tName, orderTotal).
		InnerJoin(ordersTable, tID, orderUserID).
		Where(orderTotal.Gt(100.0)).
		ToSql()

	wantSQL := "SELECT users.name, orders.total FROM users INNER JOIN orders ON users.id = orders.user_id WHERE orders.total > $1"
	wantArgs := []any{100.0}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

// ─── Inserter ────────────────────────────────────────────────

func TestInserter_SingleRow(t *testing.T) {
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

func TestInserter_MultiRow(t *testing.T) {
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

func TestInserter_MySQL(t *testing.T) {
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

func TestInserter_Models(t *testing.T) {
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

func TestInserter_Returning(t *testing.T) {
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

func TestInserter_OnConflictDoNothing(t *testing.T) {
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

func TestInserter_OnConflictDoUpdate(t *testing.T) {
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

func TestInserter_ErrorNoData(t *testing.T) {
	_, _, err := testTable.Insert(pgDB).ToSql()
	if err == nil {
		t.Fatal("expected error for insert with no data")
	}
}

func TestInserter_ErrorColumnsWithModels(t *testing.T) {
	u := &testUser{Name: "Alice"}
	_, _, err := testTable.Insert(pgDB).Columns(tName).Models(u).ToSql()
	if err == nil {
		t.Fatal("expected error for Columns + Models")
	}
}

func TestInserter_ErrorValuesAndModels(t *testing.T) {
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

func TestInserter_ErrorMismatchedValues(t *testing.T) {
	_, _, err := testTable.Insert(pgDB).
		Columns(tName, tEmail).
		Values("Alice"). // only 1 value for 2 columns
		ToSql()
	if err == nil {
		t.Fatal("expected error for mismatched values/columns")
	}
}

// ─── Batch Insert ───────────────────────────────────────────

func TestInserter_BatchModels(t *testing.T) {
	users := []*testUser{
		{ID: 1, Name: "Alice", Email: "a@test.com", Age: 20},
		{ID: 2, Name: "Bob", Email: "b@test.com", Age: 25},
		{ID: 3, Name: "Carol", Email: "c@test.com", Age: 30},
		{ID: 4, Name: "Dave", Email: "d@test.com", Age: 35},
		{ID: 5, Name: "Eve", Email: "e@test.com", Age: 40},
	}

	queries, args, err := testTable.Insert(pgDB).
		Models(users...).
		Batch(2).
		BatchQueries()

	if err != nil {
		t.Fatal(err)
	}

	if len(queries) != 3 {
		t.Fatalf("got %d queries, want 3", len(queries))
	}

	// Chunk 1: 2 rows
	wantSQL := "INSERT INTO users (id, name, email, age) VALUES ($1, $2, $3, $4), ($5, $6, $7, $8)"
	if queries[0] != wantSQL {
		t.Errorf("chunk 0 sql = %q, want %q", queries[0], wantSQL)
	}
	wantArgs := []any{1, "Alice", "a@test.com", 20, 2, "Bob", "b@test.com", 25}
	if !reflect.DeepEqual(args[0], wantArgs) {
		t.Errorf("chunk 0 args = %v, want %v", args[0], wantArgs)
	}

	// Chunk 2: 2 rows
	if queries[1] != wantSQL {
		t.Errorf("chunk 1 sql = %q, want %q", queries[1], wantSQL)
	}
	wantArgs = []any{3, "Carol", "c@test.com", 30, 4, "Dave", "d@test.com", 35}
	if !reflect.DeepEqual(args[1], wantArgs) {
		t.Errorf("chunk 1 args = %v, want %v", args[1], wantArgs)
	}

	// Chunk 3: 1 row (remainder)
	wantSQL = "INSERT INTO users (id, name, email, age) VALUES ($1, $2, $3, $4)"
	if queries[2] != wantSQL {
		t.Errorf("chunk 2 sql = %q, want %q", queries[2], wantSQL)
	}
	wantArgs = []any{5, "Eve", "e@test.com", 40}
	if !reflect.DeepEqual(args[2], wantArgs) {
		t.Errorf("chunk 2 args = %v, want %v", args[2], wantArgs)
	}
}

func TestInserter_BatchValues(t *testing.T) {
	queries, args, err := testTable.Insert(pgDB).
		Columns(tName, tEmail).
		Values("Alice", "a@test.com").
		Values("Bob", "b@test.com").
		Values("Carol", "c@test.com").
		Batch(2).
		BatchQueries()

	if err != nil {
		t.Fatal(err)
	}

	if len(queries) != 2 {
		t.Fatalf("got %d queries, want 2", len(queries))
	}

	// Chunk 1: 2 rows
	wantSQL := "INSERT INTO users (name, email) VALUES ($1, $2), ($3, $4)"
	if queries[0] != wantSQL {
		t.Errorf("chunk 0 sql = %q, want %q", queries[0], wantSQL)
	}
	wantArgs := []any{"Alice", "a@test.com", "Bob", "b@test.com"}
	if !reflect.DeepEqual(args[0], wantArgs) {
		t.Errorf("chunk 0 args = %v, want %v", args[0], wantArgs)
	}

	// Chunk 2: 1 row
	wantSQL = "INSERT INTO users (name, email) VALUES ($1, $2)"
	if queries[1] != wantSQL {
		t.Errorf("chunk 1 sql = %q, want %q", queries[1], wantSQL)
	}
	wantArgs = []any{"Carol", "c@test.com"}
	if !reflect.DeepEqual(args[1], wantArgs) {
		t.Errorf("chunk 1 args = %v, want %v", args[1], wantArgs)
	}
}

func TestInserter_BatchExact(t *testing.T) {
	// Batch size equals row count — should produce exactly 1 query
	queries, _, err := testTable.Insert(pgDB).
		Columns(tName).
		Values("Alice").
		Values("Bob").
		Batch(2).
		BatchQueries()

	if err != nil {
		t.Fatal(err)
	}

	if len(queries) != 1 {
		t.Fatalf("got %d queries, want 1", len(queries))
	}

	wantSQL := "INSERT INTO users (name) VALUES ($1), ($2)"
	if queries[0] != wantSQL {
		t.Errorf("sql = %q, want %q", queries[0], wantSQL)
	}
}

func TestInserter_BatchNoBatch(t *testing.T) {
	// BatchQueries without Batch() returns single query
	queries, args, err := testTable.Insert(pgDB).
		Columns(tName).
		Values("Alice").
		BatchQueries()

	if err != nil {
		t.Fatal(err)
	}

	if len(queries) != 1 {
		t.Fatalf("got %d queries, want 1", len(queries))
	}

	wantSQL := "INSERT INTO users (name) VALUES ($1)"
	if queries[0] != wantSQL {
		t.Errorf("sql = %q, want %q", queries[0], wantSQL)
	}
	if !reflect.DeepEqual(args[0], []any{"Alice"}) {
		t.Errorf("args = %v, want %v", args[0], []any{"Alice"})
	}
}

func TestInserter_BatchOnConflict(t *testing.T) {
	users := []*testUser{
		{ID: 1, Name: "Alice", Email: "a@test.com", Age: 20},
		{ID: 2, Name: "Bob", Email: "b@test.com", Age: 25},
		{ID: 3, Name: "Carol", Email: "c@test.com", Age: 30},
	}

	ci := testTable.Insert(pgDB).
		Models(users...).
		OnConflict(tEmail).
		DoNothing().
		Batch(2)

	// ConflictInserter doesn't have BatchQueries, so test via ToSql for the full set
	// and verify the batch size is set
	q, _, err := ci.ToSql()
	if err != nil {
		t.Fatal(err)
	}

	// ToSql returns full (non-batched) query
	if !contains(q, "ON CONFLICT (email) DO NOTHING") {
		t.Errorf("expected ON CONFLICT clause, got %q", q)
	}
	if !contains(q, "($1, $2, $3, $4)") {
		t.Errorf("expected first row placeholders, got %q", q)
	}
}

func TestInserter_BatchMySQL(t *testing.T) {
	queries, args, err := testTable.Insert(myDB).
		Columns(tName).
		Values("Alice").
		Values("Bob").
		Values("Carol").
		Batch(2).
		BatchQueries()

	if err != nil {
		t.Fatal(err)
	}

	if len(queries) != 2 {
		t.Fatalf("got %d queries, want 2", len(queries))
	}

	wantSQL := "INSERT INTO users (name) VALUES (?), (?)"
	if queries[0] != wantSQL {
		t.Errorf("chunk 0 sql = %q, want %q", queries[0], wantSQL)
	}
	wantArgs := []any{"Alice", "Bob"}
	if !reflect.DeepEqual(args[0], wantArgs) {
		t.Errorf("chunk 0 args = %v, want %v", args[0], wantArgs)
	}

	wantSQL = "INSERT INTO users (name) VALUES (?)"
	if queries[1] != wantSQL {
		t.Errorf("chunk 1 sql = %q, want %q", queries[1], wantSQL)
	}
}

// ─── Updater ─────────────────────────────────────────────────

func TestUpdater_Basic(t *testing.T) {
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

func TestUpdater_MultipleSet(t *testing.T) {
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

func TestUpdater_SetExpression(t *testing.T) {
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

func TestUpdater_Returning(t *testing.T) {
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

func TestUpdater_MySQL(t *testing.T) {
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

func TestUpdater_ErrorNoSet(t *testing.T) {
	_, _, err := testTable.Update(pgDB).Where(tID.Eq(1)).ToSql()
	if err == nil {
		t.Fatal("expected error for update with no Set")
	}
}

func TestUpdater_ErrorNoWhere(t *testing.T) {
	_, _, err := testTable.Update(pgDB).Set(tName, "Bob").ToSql()
	if err == nil {
		t.Fatal("expected error for update with no Where")
	}
}

func TestUpdater_Clone(t *testing.T) {
	base := testTable.Update(pgDB).Set(tName, "Bob").Where(tID.Eq(1))
	clone := base.Clone()
	clone.Set(tAge, 30)

	baseSql, _, _ := base.ToSql()
	cloneSql, _, _ := clone.ToSql()

	if baseSql == cloneSql {
		t.Errorf("clone mutation affected base: both = %q", baseSql)
	}
}

// ─── Deleter ─────────────────────────────────────────────────

func TestDeleter_Basic(t *testing.T) {
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

func TestDeleter_MultipleWhere(t *testing.T) {
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

func TestDeleter_Returning(t *testing.T) {
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

func TestDeleter_MySQL(t *testing.T) {
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

func TestDeleter_ForceDeleteAll(t *testing.T) {
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

func TestDeleter_ErrorNoWhere(t *testing.T) {
	_, _, err := testTable.Delete(pgDB).ToSql()
	if err == nil {
		t.Fatal("expected error for delete with no Where")
	}
}

func TestDeleter_Clone(t *testing.T) {
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
	sql, args, _ := From[testUser](pgDB, TableRef("adults")).
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

	sql, args, _ := From[struct{ N int }](pgDB, TableRef("nums")).
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

	sql, args, _ := From[testUser](pgDB, TableRef("a")).
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
	sql, args, _ := FromSub[testUser](pgDB, sub, "sub").
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

	sql, args, _ := Union[testUser](pgDB, left, right).ToSql()

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

	sql, args, _ := UnionAll[testUser](pgDB, left, right).ToSql()

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

	sql, args, _ := Intersect[testUser](pgDB, left, right).ToSql()

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

	sql, args, _ := Except[testUser](pgDB, left, right).ToSql()

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

	sql, args, _ := Union[testUser](pgDB, q1, q2).UnionAll(q3).ToSql()

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

	sql, args, _ := UnionAll[testUser](pgDB, left, right).
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

	sql, args, _ := From[testUser](pgDB, TableRef("combined")).
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
	sql, args, _ := From[testUser](pgDB, TableRef("all_users")).
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

	sql, args, _ := Union[testUser](myDB, left, right).ToSql()

	wantSQL := "(SELECT users.name FROM users WHERE users.age > ?) UNION (SELECT users.name FROM users WHERE users.name = ?)"
	wantArgs := []any{18, "Admin"}

	if sql != wantSQL {
		t.Errorf("sql = %q, want %q", sql, wantSQL)
	}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

// ─── RowScanner ─────────────────────────────────────────────

type scannerUser struct {
	ID   int    `col:"id"`
	Name string `col:"name"`
}

func (u *scannerUser) ScanRow(rows *sql.Rows) error {
	return rows.Scan(&u.ID, &u.Name)
}

func TestRowScanner_Detected(t *testing.T) {
	var zero scannerUser
	if _, ok := any(&zero).(RowScanner); !ok {
		t.Fatal("*scannerUser should implement RowScanner")
	}
}

func TestRowScanner_NotDetected(t *testing.T) {
	var zero testUser
	if _, ok := any(&zero).(RowScanner); ok {
		t.Fatal("*testUser should NOT implement RowScanner")
	}
}

// ─── ErrorMapper ─────────────────────────────────────────────

func TestDB_MapError_Nil(t *testing.T) {
	db := &DB{dialect: PostgreSQLDialect{}}
	// No error mapper — should return original error
	if got := db.mapError(nil); got != nil {
		t.Errorf("mapError(nil) = %v, want nil", got)
	}
	err := fmt.Errorf("some error")
	if got := db.mapError(err); got != err {
		t.Errorf("mapError(err) = %v, want %v", got, err)
	}
}

func TestDB_MapError_WithMapper(t *testing.T) {
	mapper := func(err error) error {
		if err.Error() == "unique_violation" {
			return ErrUniqueViolation
		}
		return err
	}
	db := &DB{dialect: PostgreSQLDialect{}, errorMapper: mapper}

	// Mapped error
	err := fmt.Errorf("unique_violation")
	if got := db.mapError(err); !errors.Is(got, ErrUniqueViolation) {
		t.Errorf("mapError() = %v, want ErrUniqueViolation", got)
	}

	// Unmapped error passes through
	other := fmt.Errorf("other error")
	if got := db.mapError(other); got != other {
		t.Errorf("mapError() = %v, want %v", got, other)
	}

	// Nil passes through even with mapper
	if got := db.mapError(nil); got != nil {
		t.Errorf("mapError(nil) = %v, want nil", got)
	}
}

func TestTx_MapError(t *testing.T) {
	mapper := func(err error) error {
		if err.Error() == "fk_violation" {
			return ErrForeignKey
		}
		return err
	}
	tx := &Tx{dialect: PostgreSQLDialect{}, errorMapper: mapper}

	err := fmt.Errorf("fk_violation")
	if got := tx.mapError(err); !errors.Is(got, ErrForeignKey) {
		t.Errorf("tx.mapError() = %v, want ErrForeignKey", got)
	}
}

func TestWithErrorMapper_Option(t *testing.T) {
	called := false
	mapper := func(err error) error {
		called = true
		return err
	}
	db := &DB{dialect: PostgreSQLDialect{}}
	opt := WithErrorMapper(mapper)
	opt(db)

	if db.errorMapper == nil {
		t.Fatal("errorMapper should be set")
	}
	db.mapError(fmt.Errorf("test"))
	if !called {
		t.Fatal("mapper should have been called")
	}
}

func TestNewDB_WithErrorMapper(t *testing.T) {
	mapper := func(err error) error { return ErrNotFound }
	db := NewDB(nil, PostgreSQLDialect{}, WithErrorMapper(mapper))

	if db.errorMapper == nil {
		t.Fatal("errorMapper should be set via NewDB option")
	}
	err := fmt.Errorf("anything")
	if got := db.mapError(err); !errors.Is(got, ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", got)
	}
}

func TestDB_MapError_SentinelErrors(t *testing.T) {
	mapper := func(err error) error {
		msg := err.Error()
		switch msg {
		case "unique":
			return ErrUniqueViolation
		case "fk":
			return ErrForeignKey
		case "check":
			return ErrCheckViolation
		case "notnull":
			return ErrNotNull
		case "notfound":
			return ErrNotFound
		default:
			return err
		}
	}
	db := &DB{dialect: PostgreSQLDialect{}, errorMapper: mapper}

	tests := []struct {
		input string
		want  error
	}{
		{"unique", ErrUniqueViolation},
		{"fk", ErrForeignKey},
		{"check", ErrCheckViolation},
		{"notnull", ErrNotNull},
		{"notfound", ErrNotFound},
	}

	for _, tt := range tests {
		got := db.mapError(errors.New(tt.input))
		if !errors.Is(got, tt.want) {
			t.Errorf("mapError(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// ─── Row Locking ─────────────────────────────────────────────

func TestSelector_ForUpdate(t *testing.T) {
	sql, _, _ := testTable.From(pgDB).Where(tID.Eq(1)).ForUpdate().ToSql()
	want := "SELECT * FROM users WHERE users.id = $1 FOR UPDATE"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_ForUpdate_NoWait(t *testing.T) {
	sql, _, _ := testTable.From(pgDB).Where(tID.Eq(1)).ForUpdate().NoWait().ToSql()
	want := "SELECT * FROM users WHERE users.id = $1 FOR UPDATE NOWAIT"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_ForUpdate_SkipLocked(t *testing.T) {
	sql, _, _ := testTable.From(pgDB).Where(tID.Eq(1)).ForUpdate().SkipLocked().ToSql()
	want := "SELECT * FROM users WHERE users.id = $1 FOR UPDATE SKIP LOCKED"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_ForShare(t *testing.T) {
	sql, _, _ := testTable.From(pgDB).Where(tID.Eq(1)).ForShare().ToSql()
	want := "SELECT * FROM users WHERE users.id = $1 FOR SHARE"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_ForUpdate_WithLimitOffset(t *testing.T) {
	sql, _, _ := testTable.From(pgDB).Where(tAge.Gt(18)).Limit(10).Offset(5).ForUpdate().ToSql()
	want := "SELECT * FROM users WHERE users.age > $1 LIMIT 10 OFFSET 5 FOR UPDATE"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_ForUpdate_MySQL(t *testing.T) {
	sql, _, _ := testTable.From(myDB).Where(tID.Eq(1)).ForUpdate().ToSql()
	want := "SELECT * FROM users WHERE users.id = ? FOR UPDATE"
	if sql != want {
		t.Errorf("got %q, want %q", sql, want)
	}
}

func TestSelector_ForUpdate_Clone(t *testing.T) {
	base := testTable.From(pgDB).Where(tID.Eq(1)).ForUpdate()
	clone := base.Clone()

	baseSql, _, _ := base.ToSql()
	cloneSql, _, _ := clone.ToSql()

	if baseSql != cloneSql {
		t.Errorf("clone sql = %q, want %q", cloneSql, baseSql)
	}

	// Mutate clone — should not affect base
	clone.ForShare()
	baseSql2, _, _ := base.ToSql()
	cloneSql2, _, _ := clone.ToSql()

	if baseSql2 == cloneSql2 {
		t.Errorf("clone mutation affected base: both = %q", baseSql2)
	}
}

// ─── Schema Factory Methods ─────────────────────────────────

func TestSchemaFactoryMethods(t *testing.T) {
	table := NewTable[testUser]("items", PostgreSQLDialect{})

	t.Run("Int64Column", func(t *testing.T) {
		col := table.Int64Column("big_id")
		if col.ColumnName() != "big_id" {
			t.Errorf("ColumnName = %q", col.ColumnName())
		}
		if col.TableName() != "items" {
			t.Errorf("TableName = %q", col.TableName())
		}
		if col.Sql() != "items.big_id" {
			t.Errorf("Sql = %q", col.Sql())
		}
	})

	t.Run("BoolColumn", func(t *testing.T) {
		col := table.BoolColumn("active")
		if col.Sql() != "items.active" {
			t.Errorf("Sql = %q", col.Sql())
		}
	})

	t.Run("TimeColumn", func(t *testing.T) {
		col := table.TimeColumn("created_at")
		if col.Sql() != "items.created_at" {
			t.Errorf("Sql = %q", col.Sql())
		}
	})

	t.Run("DecimalColumn", func(t *testing.T) {
		col := table.DecimalColumn("price")
		if col.Sql() != "items.price" {
			t.Errorf("Sql = %q", col.Sql())
		}
	})

	t.Run("BytesColumn", func(t *testing.T) {
		col := table.BytesColumn("data")
		if col.Sql() != "items.data" {
			t.Errorf("Sql = %q", col.Sql())
		}
	})

	t.Run("Float32Column", func(t *testing.T) {
		col := table.Float32Column("rating")
		if col.Sql() != "items.rating" {
			t.Errorf("Sql = %q", col.Sql())
		}
	})

	t.Run("AnyColumn", func(t *testing.T) {
		col := table.AnyColumn("meta")
		if col.Sql() != "items.meta" {
			t.Errorf("Sql = %q", col.Sql())
		}
	})

	t.Run("UUIDColumn", func(t *testing.T) {
		col := table.UUIDColumn("uuid")
		if col.Sql() != "items.uuid" {
			t.Errorf("Sql = %q", col.Sql())
		}
	})
}

type Status string

func TestDefineEnumColumn(t *testing.T) {
	table := NewTable[testUser]("users", PostgreSQLDialect{})
	col := DefineEnumColumn[Status](table, "status")

	if col.ColumnName() != "status" {
		t.Errorf("ColumnName = %q", col.ColumnName())
	}
	if col.TableName() != "users" {
		t.Errorf("TableName = %q", col.TableName())
	}
	if col.Sql() != "users.status" {
		t.Errorf("Sql = %q", col.Sql())
	}
}

func TestDefineSchema(t *testing.T) {
	type Item struct {
		ID   int
		Name string
	}

	schema := DefineSchema("items", PostgreSQLDialect{}, func(t Table[Item]) struct {
		Table[Item]
		ID   IntColumn
		Name StringColumn
	} {
		return struct {
			Table[Item]
			ID   IntColumn
			Name StringColumn
		}{
			Table: t,
			ID:    t.IntColumn("id"),
			Name:  t.StringColumn("name"),
		}
	})

	if schema.TableName() != "items" {
		t.Errorf("TableName = %q", schema.TableName())
	}
	if schema.ID.Sql() != "items.id" {
		t.Errorf("ID.Sql = %q", schema.ID.Sql())
	}
	if schema.Name.Sql() != "items.name" {
		t.Errorf("Name.Sql = %q", schema.Name.Sql())
	}

	// Verify builders work via schema
	sql, _, _ := schema.From(pgDB).Where(schema.ID.Eq(1)).ToSql()
	want := "SELECT * FROM items WHERE items.id = $1"
	if sql != want {
		t.Errorf("sql = %q, want %q", sql, want)
	}
}

// ─── Column Accessor Methods (newer types) ──────────────────

func TestDecimalColumn_Accessors(t *testing.T) {
	tbl := "products"
	col := DecimalColumn{name: "price", table: &tbl}

	if col.Args() != nil {
		t.Errorf("Args = %v, want nil", col.Args())
	}
	if col.ColumnName() != "price" {
		t.Errorf("ColumnName = %q", col.ColumnName())
	}
	if col.TableName() != "products" {
		t.Errorf("TableName = %q", col.TableName())
	}
	if col.Alias() != nil {
		t.Errorf("Alias = %v, want nil", col.Alias())
	}

	aliased := col.As("unit_price")
	if aliased.Alias() == nil || *aliased.Alias() != "unit_price" {
		t.Errorf("aliased.Alias = %v", aliased.Alias())
	}
	if aliased.Sql() != "unit_price" {
		t.Errorf("aliased.Sql = %q", aliased.Sql())
	}

	// No table
	noTbl := DecimalColumn{name: "price"}
	if noTbl.TableName() != "" {
		t.Errorf("noTbl.TableName = %q", noTbl.TableName())
	}
	if noTbl.Sql() != "price" {
		t.Errorf("noTbl.Sql = %q", noTbl.Sql())
	}

	// EqSub / NotEqSub
	sub := Raw("SELECT MIN(price) FROM products")
	if col.EqSub(sub).Sql() != "products.price = (SELECT MIN(price) FROM products)" {
		t.Errorf("EqSub Sql = %q", col.EqSub(sub).Sql())
	}
	if col.NotEqSub(sub).Sql() != "products.price != (SELECT MIN(price) FROM products)" {
		t.Errorf("NotEqSub Sql = %q", col.NotEqSub(sub).Sql())
	}
	if col.InSub(sub).Sql() != "products.price IN (SELECT MIN(price) FROM products)" {
		t.Errorf("InSub Sql = %q", col.InSub(sub).Sql())
	}
	if col.NotInSub(sub).Sql() != "products.price NOT IN (SELECT MIN(price) FROM products)" {
		t.Errorf("NotInSub Sql = %q", col.NotInSub(sub).Sql())
	}
}

func TestFloat32Column_Accessors(t *testing.T) {
	tbl := "sensors"
	col := Float32Column{name: "value", table: &tbl}

	if col.Args() != nil {
		t.Errorf("Args = %v, want nil", col.Args())
	}
	if col.ColumnName() != "value" {
		t.Errorf("ColumnName = %q", col.ColumnName())
	}
	if col.TableName() != "sensors" {
		t.Errorf("TableName = %q", col.TableName())
	}
	if col.Alias() != nil {
		t.Errorf("Alias = %v, want nil", col.Alias())
	}

	noTbl := Float32Column{name: "value"}
	if noTbl.TableName() != "" {
		t.Errorf("noTbl.TableName = %q", noTbl.TableName())
	}

	// IsNull / IsNotNull
	if col.IsNull().Sql() != "sensors.value IS NULL" {
		t.Errorf("IsNull = %q", col.IsNull().Sql())
	}
	if col.IsNotNull().Sql() != "sensors.value IS NOT NULL" {
		t.Errorf("IsNotNull = %q", col.IsNotNull().Sql())
	}

	// EqSub / NotEqSub / InSub / NotInSub
	sub := Raw("SELECT 1")
	if col.EqSub(sub).Sql() != "sensors.value = (SELECT 1)" {
		t.Errorf("EqSub = %q", col.EqSub(sub).Sql())
	}
	if col.NotEqSub(sub).Sql() != "sensors.value != (SELECT 1)" {
		t.Errorf("NotEqSub = %q", col.NotEqSub(sub).Sql())
	}
	if col.InSub(sub).Sql() != "sensors.value IN (SELECT 1)" {
		t.Errorf("InSub = %q", col.InSub(sub).Sql())
	}
	if col.NotInSub(sub).Sql() != "sensors.value NOT IN (SELECT 1)" {
		t.Errorf("NotInSub = %q", col.NotInSub(sub).Sql())
	}
}

func TestBytesColumn_Accessors(t *testing.T) {
	tbl := "files"
	col := BytesColumn{name: "data", table: &tbl}

	if col.Args() != nil {
		t.Errorf("Args = %v, want nil", col.Args())
	}
	if col.ColumnName() != "data" {
		t.Errorf("ColumnName = %q", col.ColumnName())
	}
	if col.Alias() != nil {
		t.Errorf("Alias = %v, want nil", col.Alias())
	}

	noTbl := BytesColumn{name: "data"}
	if noTbl.TableName() != "" {
		t.Errorf("noTbl.TableName = %q", noTbl.TableName())
	}
	if noTbl.Sql() != "data" {
		t.Errorf("noTbl.Sql = %q", noTbl.Sql())
	}

	aliased := col.As("file_data")
	if aliased.Alias() == nil || *aliased.Alias() != "file_data" {
		t.Errorf("aliased.Alias = %v", aliased.Alias())
	}

	// NotEq / IsNotNull
	if col.NotEq([]byte("x")).Sql() != "files.data != ?" {
		t.Errorf("NotEq = %q", col.NotEq([]byte("x")).Sql())
	}
	if col.IsNotNull().Sql() != "files.data IS NOT NULL" {
		t.Errorf("IsNotNull = %q", col.IsNotNull().Sql())
	}
}

func TestEnumColumn_Accessors(t *testing.T) {
	tbl := "users"
	col := EnumColumn[Status]{name: "status", table: &tbl}

	if col.Args() != nil {
		t.Errorf("Args = %v, want nil", col.Args())
	}
	if col.ColumnName() != "status" {
		t.Errorf("ColumnName = %q", col.ColumnName())
	}
	if col.TableName() != "users" {
		t.Errorf("TableName = %q", col.TableName())
	}
	if col.Alias() != nil {
		t.Errorf("Alias = %v, want nil", col.Alias())
	}

	noTbl := EnumColumn[Status]{name: "status"}
	if noTbl.TableName() != "" {
		t.Errorf("noTbl.TableName = %q", noTbl.TableName())
	}
}

func TestInt64Column_Accessors(t *testing.T) {
	tbl := "orders"
	col := Int64Column{name: "amount", table: &tbl}

	if col.Args() != nil {
		t.Errorf("Args = %v, want nil", col.Args())
	}
	if col.ColumnName() != "amount" {
		t.Errorf("ColumnName = %q", col.ColumnName())
	}
	if col.Alias() != nil {
		t.Errorf("Alias = %v, want nil", col.Alias())
	}

	noTbl := Int64Column{name: "amount"}
	if noTbl.TableName() != "" {
		t.Errorf("noTbl.TableName = %q", noTbl.TableName())
	}

	// NotEq / Gte / Lte / IsNotNull / NotIn / EqSub / NotEqSub / InSub / NotInSub
	if col.NotEq(5).Sql() != "orders.amount != ?" {
		t.Errorf("NotEq = %q", col.NotEq(5).Sql())
	}
	if col.Gte(10).Sql() != "orders.amount >= ?" {
		t.Errorf("Gte = %q", col.Gte(10).Sql())
	}
	if col.Lte(100).Sql() != "orders.amount <= ?" {
		t.Errorf("Lte = %q", col.Lte(100).Sql())
	}
	if col.IsNotNull().Sql() != "orders.amount IS NOT NULL" {
		t.Errorf("IsNotNull = %q", col.IsNotNull().Sql())
	}
	if col.NotIn(1, 2).Sql() != "orders.amount NOT IN (?, ?)" {
		t.Errorf("NotIn = %q", col.NotIn(1, 2).Sql())
	}
	sub := Raw("SELECT 1")
	if col.EqSub(sub).Sql() != "orders.amount = (SELECT 1)" {
		t.Errorf("EqSub = %q", col.EqSub(sub).Sql())
	}
	if col.NotEqSub(sub).Sql() != "orders.amount != (SELECT 1)" {
		t.Errorf("NotEqSub = %q", col.NotEqSub(sub).Sql())
	}
	if col.InSub(sub).Sql() != "orders.amount IN (SELECT 1)" {
		t.Errorf("InSub = %q", col.InSub(sub).Sql())
	}
	if col.NotInSub(sub).Sql() != "orders.amount NOT IN (SELECT 1)" {
		t.Errorf("NotInSub = %q", col.NotInSub(sub).Sql())
	}
}

func TestAnyColumn_Accessors(t *testing.T) {
	tbl := "events"
	col := AnyColumn{name: "data", table: &tbl}

	if col.Args() != nil {
		t.Errorf("Args = %v, want nil", col.Args())
	}
	if col.ColumnName() != "data" {
		t.Errorf("ColumnName = %q", col.ColumnName())
	}
	if col.Alias() != nil {
		t.Errorf("Alias = %v, want nil", col.Alias())
	}

	noTbl := AnyColumn{name: "data"}
	if noTbl.TableName() != "" {
		t.Errorf("noTbl.TableName = %q", noTbl.TableName())
	}

	aliased := col.As("event_data")
	if aliased.Alias() == nil || *aliased.Alias() != "event_data" {
		t.Errorf("aliased.Alias = %v", aliased.Alias())
	}

	sub := Raw("SELECT 1")
	if col.EqSub(sub).Sql() != "events.data = (SELECT 1)" {
		t.Errorf("EqSub = %q", col.EqSub(sub).Sql())
	}
	if col.NotEqSub(sub).Sql() != "events.data != (SELECT 1)" {
		t.Errorf("NotEqSub = %q", col.NotEqSub(sub).Sql())
	}
	if col.NotIn("a").Sql() != "events.data NOT IN (?)" {
		t.Errorf("NotIn = %q", col.NotIn("a").Sql())
	}
	if col.Between(1, 10).Sql() != "events.data BETWEEN ? AND ?" {
		t.Errorf("Between = %q", col.Between(1, 10).Sql())
	}
}

// ─── ConflictInserter edge cases ────────────────────────────

func TestConflictInserter_DoUpdateNoSets(t *testing.T) {
	ci := testTable.Insert(pgDB).
		Columns(tName, tEmail).
		Values("Alice", "a@test.com").
		OnConflict(tEmail)

	// Force UPDATE action but no SetUpdate calls
	ci.conflictAction = "UPDATE"

	_, _, err := ci.ToSql()
	if err == nil {
		t.Fatal("expected error for DoUpdate with no SetUpdate")
	}
}

func TestConflictInserter_DoUpdateWithExpression(t *testing.T) {
	sql, args, err := testTable.Insert(pgDB).
		Columns(tName, tEmail).
		Values("Alice", "a@test.com").
		OnConflict(tEmail).
		SetUpdate(tName, Raw("EXCLUDED.name")).
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	if !contains(sql, "DO UPDATE SET name = EXCLUDED.name") {
		t.Errorf("sql missing expression set: %q", sql)
	}
	// Only 2 args from VALUES, none from the raw expression
	wantArgs := []any{"Alice", "a@test.com"}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestConflictInserter_Returning(t *testing.T) {
	sql, _, err := testTable.Insert(pgDB).
		Columns(tName, tEmail).
		Values("Alice", "a@test.com").
		OnConflict(tEmail).
		DoNothing().
		Returning(tID, tName).
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	if !contains(sql, "RETURNING") {
		t.Errorf("sql missing RETURNING: %q", sql)
	}
}

func TestConflictInserter_ModelsAndValues(t *testing.T) {
	u := &testUser{Name: "Alice"}
	ci := testTable.Insert(pgDB).
		Columns(tName).
		Values("Bob").
		OnConflict(tEmail).
		DoNothing()
	ci.models = append(ci.models, u)

	_, _, err := ci.ToSql()
	if err == nil {
		t.Fatal("expected error for Models + Values in ConflictInserter")
	}
}

func TestConflictInserter_ColumnsWithModels(t *testing.T) {
	u := &testUser{Name: "Alice"}
	ci := testTable.Insert(pgDB).
		Models(u).
		OnConflict(tEmail).
		DoNothing()
	ci.columns = []Column{tName}

	_, _, err := ci.ToSql()
	if err == nil {
		t.Fatal("expected error for Columns + Models in ConflictInserter")
	}
}

func TestConflictInserter_NoData(t *testing.T) {
	ci := testTable.Insert(pgDB).
		OnConflict(tEmail).
		DoNothing()

	_, _, err := ci.ToSql()
	if err == nil {
		t.Fatal("expected error for ConflictInserter with no data")
	}
}

func TestConflictInserter_ValuesNoColumns(t *testing.T) {
	ci := testTable.Insert(pgDB).
		OnConflict(tEmail).
		DoNothing()
	ci.values = append(ci.values, []any{"Alice"})

	_, _, err := ci.ToSql()
	if err == nil {
		t.Fatal("expected error for Values without Columns in ConflictInserter")
	}
}

// ─── Selector Clone branches ────────────────────────────────

func TestSelector_Clone_AllBranches(t *testing.T) {
	ordersTable := NewTable[testUser]("orders", PostgreSQLDialect{})
	oID := ordersTable.IntColumn("id")

	base := testTable.From(pgDB).
		Select(tID, tName).
		Where(tAge.Gt(18)).
		Distinct(tName).
		OrderBy(Asc(tName)).
		GroupBy(tName).
		Having(Count().Gt(1)).
		InnerJoin(ordersTable, tID, oID).
		Limit(10).
		Offset(5)

	clone := base.Clone()

	// Mutate clone — should not affect base
	clone.Select(tEmail)
	clone.Where(tName.Eq("Bob"))
	clone.OrderBy(Desc(tAge))
	clone.GroupBy(tAge)
	clone.Having(Count().Lt(100))
	clone.Limit(20)

	baseSql, _, _ := base.ToSql()
	cloneSql, _, _ := clone.ToSql()

	if baseSql == cloneSql {
		t.Errorf("clone mutation affected base: both = %q", baseSql)
	}

	// Verify base still has original values
	if !contains(baseSql, "LIMIT 10") {
		t.Errorf("base lost LIMIT 10: %q", baseSql)
	}
	if contains(baseSql, "LIMIT 20") {
		t.Errorf("base got clone's LIMIT 20: %q", baseSql)
	}
}

func TestSelector_Clone_WithCTEs(t *testing.T) {
	sub := testTable.From(pgDB).Where(tAge.Gt(18))
	base := From[testUser](pgDB, TableRef("adults")).
		With(CTE("adults", sub)).
		Where(Raw("1=1"))

	clone := base.Clone()
	clone.Where(Raw("adults.name = ?", "Alice"))

	baseSql, baseArgs, _ := base.ToSql()
	cloneSql, cloneArgs, _ := clone.ToSql()

	if baseSql == cloneSql {
		t.Errorf("clone CTE mutation affected base")
	}
	if len(baseArgs) == len(cloneArgs) {
		t.Errorf("clone args should differ from base")
	}
}

// ─── MSSQL Deleter ──────────────────────────────────────────

func TestDeleter_MSSQL(t *testing.T) {
	sql, args, err := testTable.Delete(mssqlDB).
		Where(tID.Eq(1)).
		ToSql()

	if err != nil {
		t.Fatal(err)
	}

	wantSQL := "DELETE FROM users WHERE users.id = @p1"
	wantArgs := []any{1}

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
