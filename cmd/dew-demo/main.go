package main

import (
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

	users := []*User{user1, user2, user3}

	for _, user := range users {
		err = dew.Insert[User](db, Users).
			Columns(Users.Name, Users.Profile, Users.Address).
			Values(user.Name, user.Profile, user.Address).
			Exec()
		if err != nil {
			log.Fatal("insert:", err)
		}
	}

	// -- HasKey (?) --
	fmt.Println("=== HasKey: profile ? 'role' ===")
	user, err := Users.From(db).
		Select(Users.Name, Users.Profile, Users.Address).
		Where(Users.Profile.HasKey("role")).
		One()
	if err != nil {
		log.Fatal("HasKey:", err)
	}
	fmt.Println(user.String())

	// -- Contains (@>) --
	fmt.Println("\n=== Contains: profile @> {\"age\": 30} ===")
	thirtyYearOlds, err := Users.From(db).
		Select(Users.Name, Users.Profile).
		Where(Users.Profile.Contains(&Profile{"age": float64(30)})).
		All()
	if err != nil {
		log.Fatal("Contains:", err)
	}
	for _, u := range thirtyYearOlds {
		fmt.Println(u.String())
	}

	// -- ContainedBy (<@) --
	fmt.Println("\n=== ContainedBy: profile <@ superset ===")
	superset := &Profile{"name": "Jim Beam", "age": float64(40), "role": "admin", "extra": "ignored"}
	contained, err := Users.From(db).
		Select(Users.Name, Users.Profile).
		Where(Users.Profile.ContainedBy(superset)).
		All()
	if err != nil {
		log.Fatal("ContainedBy:", err)
	}
	for _, u := range contained {
		fmt.Println(u.String())
	}

	fmt.Println("\n=== HasAnyKey: profile ?| ['role', 'missing'] ===")
	anyKey, err := Users.From(db).
		Select(Users.Name, Users.Profile).
		Where(Users.Profile.HasAnyKey("role", "missing")).
		All()
	if err != nil {
		log.Fatal("HasAnyKey:", err)
	}
	for _, u := range anyKey {
		fmt.Println(u.String())
	}

	// -- HasAllKeys (?&) --
	fmt.Println("\n=== HasAllKeys: profile ?& ['name', 'age'] ===")
	allKeys, err := Users.From(db).
		Select(Users.Name, Users.Profile).
		Where(Users.Profile.HasAllKeys("name", "age")).
		All()
	if err != nil {
		log.Fatal("HasAllKeys:", err)
	}
	for _, u := range allKeys {
		fmt.Println(u.String())
	}

	// -- PathText (->>) --
	fmt.Println("\n=== PathText: address->>'city' ===")
	sql, args := Users.From(db).
		Select(Users.Name, Users.Address.PathText("city")).
		ToSql()
	fmt.Printf("SQL:  %s\nArgs: %v\n", sql, args)

	// -- Eq (exact JSONB match) --
	fmt.Println("\n=== Eq: address = exact ===")
	exact, err := Users.From(db).
		Select(Users.Name, Users.Address).
		Where(Users.Address.Eq(&Address{City: "Chicago", Street: "789 Main St", Zip: "60601"})).
		All()
	if err != nil {
		log.Fatal("Eq:", err)
	}
	for _, u := range exact {
		fmt.Println(u.String())
	}

	// -- IsNull --
	fmt.Println("\n=== IsNull: address IS NULL ===")
	nullAddr, err := Users.From(db).
		Select(Users.Name).
		Where(Users.Address.IsNull()).
		All()
	if err != nil {
		log.Fatal("IsNull:", err)
	}
	fmt.Printf("Users with null address: %d\n", len(nullAddr))

	// -- And/Or composition --
	fmt.Println("\n=== And/Or: (profile ? 'role') AND (address->>'city' = 'Los Angeles') ===")
	composed, err := Users.From(db).
		Select(Users.Name, Users.Profile, Users.Address).
		Where(dew.And(
			Users.Profile.HasKey("role"),
			Users.Address.Contains(&Address{City: "Los Angeles"}),
		)).
		All()
	if err != nil {
		log.Fatal("And/Or:", err)
	}
	for _, u := range composed {
		fmt.Println(u.String())
	}

	fmt.Println("\nDone.")
}
