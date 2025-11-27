package main

import (
	"fmt"
	"log"

	"github.com/dr3dnought/dew"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID    int
	Name  string
	Email string
}

var DewUser = struct {
	ID    dew.IntColumn
	Name  dew.StringColumn
	Email dew.StringColumn
}{
	ID:    "id", // Имя колонки в БД
	Name:  "name",
	Email: "email",
}

func main() {
	db, err := dew.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	_, _ = db.Exec("CREATE TABLE users (id INTEGER, name TEXT, email TEXT)")
	_, _ = db.Exec("INSERT INTO users VALUES (1, 'Alice', 'alice@dew.com')")
	_, _ = db.Exec("INSERT INTO users VALUES (2, 'Bob', 'bob@dew.com')")

	type Req struct {
		id   int
		desc bool
	}

	req := Req{
		id:   10,
		desc: true,
	}

	selector := dew.From[User](db)

	if req.id > 0 {
		selector.Select(DewUser.Name, DewUser.Email).Where(dew.Or(DewUser.ID.Eq(1), DewUser.Name.Eq("Alice")))
	}

	if req.desc {
		selector.OrderBy(dew.Desc(DewUser.Name))
	}

	users, err := selector.All()

	if err != nil {
		log.Fatal(err)
	}

	for _, u := range users {
		fmt.Printf("User: %s [%s]\n", u.Name, u.Email)
	}
}
