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
	"github.com/orlangure/gnomock/internal/gnomockd"
	_ "github.com/orlangure/gnomock/preset/mariadb"
	"github.com/stretchr/testify/require"
)

func TestMariaDB(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	bs, err := ioutil.ReadFile("./testdata/mariadb.json")
	require.NoError(t, err)

	buf := bytes.NewBuffer(bs)
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/mariadb", buf)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equalf(t, http.StatusOK, res.StatusCode, string(body))

	var c *gnomock.Container

	err = json.Unmarshal(body, &c)
	require.NoError(t, err)

	addr := c.DefaultAddress()
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s",
		"gnomock", "foobar", addr, "gnomockd_db",
	)

	db, err := sql.Open("mysql", connStr)
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
