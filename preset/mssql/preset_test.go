package mssql_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/mssql"
	"github.com/stretchr/testify/require"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestPreset(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"2017-latest", "2019-latest"} {
		t.Run(version, testPreset(version))
	}
}

func testPreset(version string) func(t *testing.T) {
	return func(t *testing.T) {
		queries := `
			insert into t (a) values (1);
			insert into t (a) values (2);
		`
		query := `insert into t (a) values (3);`
		p := mssql.Preset(
			mssql.WithLicense(true),
			mssql.WithAdminPassword("Passw0rd-"),
			mssql.WithQueries(queries, query),
			mssql.WithDatabase("foobar"),
			mssql.WithVersion(version),
			mssql.WithQueriesFile("./testdata/queries.sql"),
		)

		container, err := gnomock.Start(
			p,
			gnomock.WithLogWriter(os.Stdout),
			gnomock.WithTimeout(time.Minute*10),
		)

		defer func() { _ = gnomock.Stop(container) }()

		require.NoError(t, err)

		addr := container.DefaultAddress()
		connStr := fmt.Sprintf("sqlserver://sa:Passw0rd-@%s?database=foobar", addr)

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
		require.NoError(t, db.Close())
	}
}

func TestPreset_withDefaults(t *testing.T) {
	t.Parallel()

	p := mssql.Preset(mssql.WithLicense(true))
	container, err := gnomock.Start(p, gnomock.WithTimeout(time.Minute*10))

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	require.NoError(t, err)

	addr := container.DefaultAddress()
	connStr := fmt.Sprintf("sqlserver://sa:Gn0m!ck~@%s?database=mydb", addr)

	db, err := sql.Open("sqlserver", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Close())
}

func TestPreset_wrongQueriesFile(t *testing.T) {
	t.Parallel()

	p := mssql.Preset(
		mssql.WithLicense(true),
		mssql.WithQueriesFile("./invalid"),
	)
	c, err := gnomock.Start(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "can't read queries file")
	require.NoError(t, gnomock.Stop(c))
}
