package main

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/dr3dnought/dew"
	_ "github.com/lib/pq"
)

// -- JSONB types (implement dew.JSONB: sql.Scanner + driver.Valuer) --

type Profile map[string]any

func (p *Profile) Scan(src any) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	default:
		return fmt.Errorf("Profile.Scan: unsupported type %T", src)
	}
}

func (p Profile) Value() (driver.Value, error) {
	return json.Marshal(p)
}

type Address struct {
	City   string `json:"city"`
	Street string `json:"street"`
	Zip    string `json:"zip"`
}

func (a *Address) Scan(src any) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	case string:
		return json.Unmarshal([]byte(v), a)
	default:
		return fmt.Errorf("Address.Scan: unsupported type %T", src)
	}
}

func (a Address) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// -- Model --

type User struct {
	ID      int      `db:"id"`
	Name    string   `db:"name"`
	Profile *Profile `db:"profile"`
	Address *Address `db:"address"`
}

func (u *User) String() string {
	json, err := json.Marshal(u)
	if err != nil {
		return fmt.Sprintf("User{ID: %d, Name: %s, Profile: %v, Address: %v}", u.ID, u.Name, u.Profile, u.Address)
	}
	return string(json)
}

// -- Schema --

var Users = dew.DefineSchema("users", dew.PostgreSQLDialect{}, func(t dew.Table[User]) struct {
	dew.Table[User]
	ID      dew.IntColumn
	Name    dew.StringColumn
	Profile dew.JSONBColumn[*Profile]
	Address dew.JSONBColumn[*Address]
} {
	return struct {
		dew.Table[User]
		ID      dew.IntColumn
		Name    dew.StringColumn
		Profile dew.JSONBColumn[*Profile]
		Address dew.JSONBColumn[*Address]
	}{
		Table:   t,
		ID:      t.IntColumn("id"),
		Name:    t.StringColumn("name"),
		Profile: dew.DefineJSONBColumn[*Profile](t, "profile"),
		Address: dew.DefineJSONBColumn[*Address](t, "address"),
	}
})

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://dew:dew@localhost:5432/demo?sslmode=disable"
	}

	db, err := dew.Open("postgres", dsn, dew.PostgreSQLDialect{})
	if err != nil {
		log.Fatal(err)
	}

	// Setup
	_, err = db.Exec(`
		DROP TABLE IF EXISTS users;
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			profile JSONB,
			address JSONB
		);
	`)
	if err != nil {
		log.Fatal("setup:", err)
	}

	user1 := &User{
		Name: "John Doe",
		Profile: &Profile{
			"name": "John Doe",
			"age":  30,
		},
		Address: &Address{
			City:   "New York",
			Street: "123 Main St",
			Zip:    "10001",
		},
	}

	user2 := &User{
		Name: "Bob Smith",
		Profile: &Profile{
			"name": "Bob Smith",
			"age":  25,
			"role": "admin",
		},
		Address: &Address{
			City:   "Los Angeles",
			Street: "456 Main St",
			Zip:    "90001",
		},
	}

	user3 := &User{
		Name: "Jim Beam",
		Profile: &Profile{
			"name": "Jim Beam",
			"age":  40,
		},
		Address: &Address{
			City:   "Chicago",
			Street: "789 Main St",
			Zip:    "60601",
		},
	}

	// === Transaction example ===
	fmt.Println("=== Transaction: insert two users atomically ===")

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal("begin tx:", err)
	}

	// Both inserts go through the transaction
	err = Users.Insert(tx).
		Columns(Users.Name, Users.Profile, Users.Address).
		Values(user1.Name, user1.Profile, user1.Address).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		log.Fatal("tx insert 1:", err)
	}

	err = Users.Insert(tx).
		Columns(Users.Name, Users.Profile, Users.Address).
		Values(user2.Name, user2.Profile, user2.Address).
		Exec(ctx)
	if err != nil {
		tx.Rollback()
		log.Fatal("tx insert 2:", err)
	}

	// Query within the same transaction — sees uncommitted rows
	count, err := Users.From(tx).Count(ctx)
	if err != nil {
		tx.Rollback()
		log.Fatal("tx count:", err)
	}
	fmt.Printf("Count inside tx: %d\n", count)

	if err := tx.Commit(); err != nil {
		log.Fatal("commit:", err)
	}
	fmt.Println("Transaction committed.")

	// === Rollback example ===
	fmt.Println("\n=== Transaction rollback: insert + rollback ===")

	tx2, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal("begin tx2:", err)
	}

	err = Users.Insert(tx2).
		Columns(Users.Name, Users.Profile, Users.Address).
		Values(user3.Name, user3.Profile, user3.Address).
		Exec(ctx)
	if err != nil {
		tx2.Rollback()
		log.Fatal("tx2 insert:", err)
	}

	countInTx, err := Users.From(tx2).Count(ctx)
	if err != nil {
		tx2.Rollback()
		log.Fatal("tx2 count:", err)
	}
	fmt.Printf("Count inside tx2 (before rollback): %d\n", countInTx)

	tx2.Rollback()
	fmt.Println("Transaction rolled back.")

	// Verify rollback — count should be same as after first tx
	countAfter, err := Users.From(db).Count(ctx)
	if err != nil {
		log.Fatal("count after rollback:", err)
	}
	fmt.Printf("Count after rollback: %d\n", countAfter)

	fmt.Println("\nDone.")
}
