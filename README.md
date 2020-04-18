# Gnomock MySQL ![Build](https://github.com/orlangure/gnomock-mysql/workflows/Build/badge.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/orlangure/gnomock-mysql)](https://goreportcard.com/report/github.com/orlangure/gnomock-mysql)

Gnomock MySQL is a [Gnomock](https://github.com/orlangure/gnomock) preset for
running tests against a real MySQL database, without mocks.

```go
package mysql_test

import (
	"database/sql"
	"fmt"

	"github.com/orlangure/gnomock"
	mockmysql "github.com/orlangure/gnomock-mysql"
)

func ExamplePreset() {
	queries := `
		create table t(a int);
		insert into t (a) values (1);
		insert into t (a) values (2);
	`
	query := `insert into t (a) values (3);`
	p := mockmysql.Preset(
		mockmysql.WithUser("Sherlock", "Holmes"),
		mockmysql.WithDatabase("books"),
		mockmysql.WithQueries(queries, query),
	)

	container, err := gnomock.Start(p)

	defer func() { _ = gnomock.Stop(container) }()

	if err != nil {
		panic(err)
	}

	addr := container.DefaultAddress()
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s",
		"Sherlock", "Holmes", addr, "books",
	)

	db, err := sql.Open("mysql", connStr)
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
```
