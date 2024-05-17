package postgres_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/postgres"
	"github.com/stretchr/testify/require"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestPreset(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"10.15", "11.10", "12.5", "13.1", "14.11", "15.6", "16.2"} {
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
		p := postgres.Preset(
			postgres.WithUser("gnomock", "gnomick"),
			postgres.WithDatabase("mydb"),
			postgres.WithQueries(queries, query),
			postgres.WithQueriesFile("./testdata/queries.sql"),
			postgres.WithVersion(version),
			postgres.WithTimezone("Europe/Paris"),
		)

		container, err := gnomock.Start(p)
		require.NoError(t, err)

		defer func() { require.NoError(t, gnomock.Stop(container)) }()

		connStr := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s  dbname=%s sslmode=disable",
			container.Host, container.DefaultPort(),
			"gnomock", "gnomick", "mydb",
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

		var timezone string

		timezoneRow := db.QueryRow("show timezone")
		require.NoError(t, timezoneRow.Scan(&timezone))
		require.Equal(t, "Europe/Paris", timezone)

		require.NoError(t, db.Close())
	}
}

func TestPreset_withDefaults(t *testing.T) {
	t.Parallel()

	p := postgres.Preset()
	container, err := gnomock.Start(p)
	require.NoError(t, err)

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s  dbname=%s sslmode=disable",
		container.Host, container.DefaultPort(),
		"postgres", "password", "postgres",
	)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, db.Ping())
	require.NoError(t, err)

	var timezone string

	timezoneRow := db.QueryRow("show timezone")
	require.NoError(t, timezoneRow.Scan(&timezone))
	require.Equal(t, "Etc/UTC", timezone)

	require.NoError(t, db.Close())
}

func TestPreset_wrongQueriesFile(t *testing.T) {
	t.Parallel()

	p := postgres.Preset(
		postgres.WithQueriesFile("./invalid"),
	)
	c, err := gnomock.Start(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "can't read queries file")
	require.NoError(t, gnomock.Stop(c))
}

func TestPreset_withContainerReuseAndDatabase(t *testing.T) {
	t.Parallel()

	p := postgres.Preset(postgres.WithDatabase("reused"))

	c1, err := gnomock.Start(
		p,
		gnomock.WithContainerReuse(),
		gnomock.WithContainerName("reusable-postgres"),
	)
	require.NoError(t, err)
	require.NotNil(t, c1)

	c2, err := gnomock.Start(
		p,
		gnomock.WithContainerReuse(),
		gnomock.WithContainerName("reusable-postgres"),
	)
	require.NoError(t, err)
	require.NotNil(t, c2)
	require.Equal(t, c1.ID, c2.ID)

	t.Cleanup(func() { require.NoError(t, gnomock.Stop(c1, c2)) })
}
