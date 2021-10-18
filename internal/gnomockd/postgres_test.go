package gnomockd_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/gnomockd"
	_ "github.com/orlangure/gnomock/preset/postgres"
	"github.com/stretchr/testify/require"
)

func TestPostgres(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	bs, err := os.ReadFile("./testdata/postgres.json")
	require.NoError(t, err)

	buf := bytes.NewBuffer(bs)
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/postgres", buf)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equalf(t, http.StatusOK, res.StatusCode, string(body))

	var c *gnomock.Container

	err = json.Unmarshal(body, &c)
	require.NoError(t, err)

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s  dbname=%s sslmode=disable",
		c.Host, c.DefaultPort(),
		"gnomock", "foobar", "gnomockd_db",
	)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	row := db.QueryRow(`select count(distinct ip_address) from customers`)
	count := 0
	require.NoError(t, row.Scan(&count))
	require.Equal(t, 1000, count)

	row = db.QueryRow(`select a from tbl`)
	value := 0
	require.NoError(t, row.Scan(&value))
	require.Equal(t, 42, value)

	row = db.QueryRow(`select max(foo) from things limit 1`)
	value = 0
	require.NoError(t, row.Scan(&value))
	require.Equal(t, 3, value)

	bs, err = json.Marshal(c)
	require.NoError(t, err)

	buf = bytes.NewBuffer(bs)
	w, r = httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", buf)
	h.ServeHTTP(w, r)

	res = w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
}
