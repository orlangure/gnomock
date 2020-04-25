package gnomockd_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomockd/gnomockd"
	"github.com/stretchr/testify/require"
)

//nolint:bodyclose
func TestPostgres(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	bs, err := ioutil.ReadFile("./testdata/postgres.json")
	require.NoError(t, err)

	buf := bytes.NewBuffer(bs)
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/postgres", buf)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	require.Equal(t, http.StatusOK, res.StatusCode)

	var c *gnomock.Container

	err = json.NewDecoder(res.Body).Decode(&c)
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

	bs, err = json.Marshal(c)
	require.NoError(t, err)

	buf = bytes.NewBuffer(bs)
	w, r = httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", buf)
	h.ServeHTTP(w, r)

	res = w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
}
