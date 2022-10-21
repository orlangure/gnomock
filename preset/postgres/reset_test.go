package postgres_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/postgres"
	"github.com/stretchr/testify/suite"
)

func TestReset(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ResetTestSuite))
}

type ResetTestSuite struct {
	suite.Suite

	container *gnomock.Container
	connStr   string
}

func (t *ResetTestSuite) TearDownSuite() {
	t.NoError(gnomock.Stop(t.container))
}

func (t *ResetTestSuite) SetupTest() {
	p := postgres.Preset(
		postgres.WithQueries(
			"create table foo(a text)",
			"insert into foo(a) values ('b')",
		),
		postgres.WithUser("gnomock", "foobar"),
	)

	c, err := gnomock.Start(
		p,
		gnomock.WithContainerName("postgres-reuse"),
		gnomock.WithContainerReuse(),
		gnomock.WithContainerReset(postgres.Reset()),
	)
	t.NoError(err)
	t.NotNil(c)

	t.container = c
	t.connStr = fmt.Sprintf(
		"host=%s port=%d user=%s password=%s  dbname=%s sslmode=disable",
		c.Host, c.DefaultPort(),
		"gnomock", "foobar", "postgres",
	)
}

func (t *ResetTestSuite) TestResetFirst() {
	db, err := sql.Open("postgres", t.connStr)
	t.NoError(err)
	t.NoError(db.Ping())
	t.T().Cleanup(func() { t.NoError(db.Close()) })

	_, err = db.Exec("insert into foo(a) values ('c'), ('d')")
	t.NoError(err)
}

func (t *ResetTestSuite) TestResetSecond() {
	db, err := sql.Open("postgres", t.connStr)
	t.NoError(err)
	t.NoError(db.Ping())
	t.T().Cleanup(func() { t.NoError(db.Close()) })

	var value string
	err = db.QueryRow("select * from foo").Scan(&value)
	t.NoError(err)
	t.Equal("b", value)
}
