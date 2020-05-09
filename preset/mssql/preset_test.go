// +build !nopreset

package mssql_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/mssql"
	"github.com/stretchr/testify/require"
)

func ExamplePreset() {
	queries := `
		create table t(a int);
		insert into t (a) values (1);
		insert into t (a) values (2);
	`
	query := `insert into t (a) values (3);`
	p := mssql.Preset(
		mssql.WithLicense(true),
		mssql.WithAdminPassword("Passw0rd-"),
		mssql.WithQueries(queries, query),
		mssql.WithDatabase("foobar"),
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

func TestMSSQL_defaultValues(t *testing.T) {
	queries := `
		create table t(a int);
		insert into t (a) values (1);
		insert into t (a) values (2);
	`
	query := `insert into t (a) values (3);`
	p := mssql.Preset(
		mssql.WithLicense(true),
		mssql.WithQueries(queries, query),
	)

	container, err := gnomock.Start(p)

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	require.NoError(t, err)

	addr := container.DefaultAddress()
	connStr := fmt.Sprintf("sqlserver://sa:Gn0m!ck~@%s?database=mydb", addr)

	db, err := sql.Open("sqlserver", connStr)
	require.NoError(t, err)

	var max, avg, min, count float64

	rows := db.QueryRow("select max(a), avg(a), min(a), count(a) from t")

	err = rows.Scan(&max, &avg, &min, &count)
	require.NoError(t, err)

	require.Equal(t, float64(3), max)
	require.Equal(t, float64(2), avg)
	require.Equal(t, float64(1), min)
	require.Equal(t, float64(3), count)
}
