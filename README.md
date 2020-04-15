# Gnomock MSSQL ![Build](https://github.com/orlangure/gnomock-mssql/workflows/Build/badge.svg?branch=master)

Gnomock MSSQL is a [Gnomock](https://github.com/orlangure/gnomock) preset for
running tests against a real Microsoft SQL Server database, without mocks.

```go
package mssql_test

import (
	"database/sql"
	"fmt"

	"github.com/orlangure/gnomock"
	mockmssql "github.com/orlangure/gnomock-mssql"
)

func ExamplePreset() {
	queries := `
		create table t(a int);
		insert into t (a) values (1);
		insert into t (a) values (2);
	`
	query := `insert into t (a) values (3);`
	p := mockmssql.Preset(
		mockmssql.WithLicense(true),
		mockmssql.WithAdminPassword("Passw0rd-"),
		mockmssql.WithQueries(queries, query),
		mockmssql.WithDatabase("foobar"),
	)

	container, err := gnomock.Start(p)

	defer func() { _ = gnomock.Stop(container) }()

	if err != nil {
		panic(err)
	}

	addr := container.DefaultAddress()
	connStr := fmt.Sprintf("sqlserver://sa:Passw0rd-@%s?database=foobar", addr)

	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		panic(err)
	}

	var max, avg, min, count float64

	rows := db.QueryRow("select max(a), avg(a), min(a), count(a) from t")

	err = rows.Scan(&max, &avg, &min, &count)
	if err != nil {
		panic(err)
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
