package postgres_test

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/orlangure/gnomock"
	mockpostgres "github.com/orlangure/gnomock-postgres"
)

func ExamplePostgres() {
	queries := []string{
		"create table t(a int)",
		"insert into t (a) values (1)",
		"insert into t (a) values (2)",
		"insert into t (a) values (3)",
	}
	p := mockpostgres.Preset(
		mockpostgres.WithUser("gnomock", "gnomick"),
		mockpostgres.WithDatabase("mydb"),
		mockpostgres.WithQueries(queries),
	)

	container, err := gnomock.StartPreset(p)
	if err != nil {
		panic(err)
	}

	defer func() { _ = gnomock.Stop(container) }()

	addr := container.Address(gnomock.DefaultPort)
	addrParts := strings.Split(addr, ":")
	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s port=%s dbname=%s sslmode=disable",
		addrParts[0], "gnomock", "gnomick", addrParts[1], "mydb",
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	var max, avg, min, count float64

	rows := db.QueryRow("select max(a), avg(a), min(a), count(a) from t")

	err = rows.Scan(&max, &avg, &min, &count)
	if err != nil {
		panic("can't query the database: " + err.Error())
	}

	fmt.Println("max", 3)
	fmt.Println("avg", 2)
	fmt.Println("min", 1)
	fmt.Println("count", 3)

	// Output:
	// max 3
	// avg 2
	// min 1
	// count 3
}
