package cassandra_test

import (
	"testing"

	"github.com/gocql/gocql"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/cassandra"
	"github.com/stretchr/testify/require"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestPreset(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"4.0", "3"} {
		t.Run(version, testPreset(version))
	}
}

func testPreset(version string) func(t *testing.T) {
	return func(t *testing.T) {
		p := cassandra.Preset(
			cassandra.WithVersion(version),
		)
		container, err := gnomock.Start(p)

		defer func() { require.NoError(t, gnomock.Stop(container)) }()

		require.NoError(t, err)

		addr := container.DefaultAddress()
		require.NotEmpty(t, addr)

		cluster := gocql.NewCluster(addr)
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cassandra.DefaultUser,
			Password: cassandra.DefaultPassword,
		}

		session, err := cluster.CreateSession()
		require.NoError(t, err)

		defer session.Close()

		err = session.Query(`
			create keyspace gnomock
			with replication = {'class':'SimpleStrategy', 'replication_factor' : 1};
		`).Exec()
		require.NoError(t, err)

		err = session.Query("CREATE TABLE gnomock.test (id UUID, PRIMARY KEY (id));").Exec()
		require.NoError(t, err)
	}
}
