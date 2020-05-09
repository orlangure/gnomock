// +build !nopreset

package postgres_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/postgres"
	"github.com/stretchr/testify/require"
)

func TestPreset(t *testing.T) {
	t.Parallel()

	queries := `
		create table t(a int);
		insert into t (a) values (1);
		insert into t (a) values (2);
	`
	query := `insert into t (a) values (3);`
	p := postgres.Preset(
		postgres.WithUser("gnomock", "gnomick"),
		postgres.WithDatabase("mydb"),
		postgres.WithQueries(queries, query),
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
}
