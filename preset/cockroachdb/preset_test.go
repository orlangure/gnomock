package cockroachdb_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/cockroachdb"
	"github.com/stretchr/testify/require"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestPreset(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"v19.2.11", "v20.1.10", "v21.2.17", "v22.2.19", "v23.1.20"} {
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
		p := cockroachdb.Preset(
			cockroachdb.WithDatabase("gnomock"),
			cockroachdb.WithQueries(queries, query),
			cockroachdb.WithQueriesFile("./testdata/queries.sql"),
			cockroachdb.WithVersion(version),
		)

		container, err := gnomock.Start(p)
		require.NoError(t, err)

		defer func() { require.NoError(t, gnomock.Stop(container)) }()

		connStr := fmt.Sprintf(
			"host=%s port=%d user=root dbname=%s sslmode=disable",
			container.Host, container.DefaultPort(), "gnomock",
		)

		db, err := sql.Open("postgres", connStr)
		require.NoError(t, err)

		var max, avg, min, count float64

		rows := db.QueryRow("select max(a), avg(a), min(a), count(a) from t")
		require.NoError(t, rows.Scan(&max, &avg, &min, &count))

		require.Equal(t, float64(3), max)
		require.Equal(t, float64(2), avg)
		require.Equal(t, float64(1), min)
		require.Equal(t, float64(3), count)
		require.NoError(t, db.Close())
	}
}

func TestPreset_withDefaults(t *testing.T) {
	t.Parallel()

	container, err := gnomock.Start(cockroachdb.Preset())

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	require.NoError(t, err)

	connStr := fmt.Sprintf(
		"host=%s port=%d user=root dbname=%s sslmode=disable",
		container.Host, container.DefaultPort(), "mydb",
	)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	require.NoError(t, db.Close())
}

func TestPreset_wrongQueriesFile(t *testing.T) {
	t.Parallel()

	p := cockroachdb.Preset(
		cockroachdb.WithQueriesFile("./invalid"),
	)
	c, err := gnomock.Start(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "can't read queries file")
	require.NoError(t, gnomock.Stop(c))
}
