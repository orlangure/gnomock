package mysql_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/mysql"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestPreset(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"8.0.22", "5.7.32"} {
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
		p := mysql.Preset(
			mysql.WithUser("Sherlock", "Holmes"),
			mysql.WithDatabase("books"),
			mysql.WithQueries(queries, query),
			mysql.WithQueriesFile("./testdata/queries.sql"),
			mysql.WithVersion(version),
		)

		container, err := gnomock.Start(p)

		defer func() { _ = gnomock.Stop(container) }()

		require.NoError(t, err)

		addr := container.DefaultAddress()
		connStr := fmt.Sprintf(
			"%s:%s@tcp(%s)/%s",
			"Sherlock", "Holmes", addr, "books",
		)

		db, err := sql.Open("mysql", connStr)
		require.NoError(t, err)

		var maximum, avg, minimum, count float64

		rows := db.QueryRow("select max(a), avg(a), min(a), count(a) from t")

		err = rows.Scan(&maximum, &avg, &minimum, &count)
		require.NoError(t, err)

		require.Equal(t, float64(3), maximum)
		require.Equal(t, float64(2), avg)
		require.Equal(t, float64(1), minimum)
		require.Equal(t, float64(3), count)

		require.NoError(t, db.Close())
	}
}

func TestPreset_withDefaults(t *testing.T) {
	t.Parallel()

	p := mysql.Preset()
	container, err := gnomock.Start(p)

	defer func() { _ = gnomock.Stop(container) }()

	require.NoError(t, err)

	addr := container.DefaultAddress()
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s",
		"gnomock", "gnomock", addr, "mydb",
	)

	db, err := sql.Open("mysql", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Close())
}

func TestPreset_wrongQueriesFile(t *testing.T) {
	t.Parallel()

	p := mysql.Preset(mysql.WithQueriesFile("./invalid"))
	c, err := gnomock.Start(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "can't read queries file")
	require.NoError(t, gnomock.Stop(c))
}
