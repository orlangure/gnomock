// Package cockroachdb includes CockroachDB implementation of Gnomock Preset interface.
// This Preset can be passed to gnomock.Start() function to create a configured
// CockroachDB container to use in tests.
//
// Containers created with this preset use `root` user without a password for
// authentication. There is currently no way to setup an initial user at create
// time.
//
// By default, a new database "mydb" is created, and all the provided queries
// are executed against it.
package cockroachdb

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"

	_ "github.com/lib/pq" // postgres driver
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
)

const (
	defaultVersion  = "v20.1.10"
	defaultPort     = 26257
	defaultDatabase = "mydb"
)

func init() {
	registry.Register("cockroachdb", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gmomock CockroachDB preset. This preset includes a CockroachDB
// specific healthcheck function and default CockroachDB image and port.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation for CockroachDB.
type P struct {
	Version      string   `json:"version"`
	DB           string   `json:"db"`
	Queries      []string `json:"queries"`
	QueriesFiles []string `json:"queries_files"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/cockroachdb/cockroach:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(defaultPort)
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(healthcheck),
		gnomock.WithCommand("start-single-node", "--insecure"),
		gnomock.WithInit(p.initf()),
	}

	return opts
}

func (p *P) setDefaults() {
	if p.Version == "" {
		p.Version = defaultVersion
	}

	if p.DB == "" {
		p.DB = defaultDatabase
	}
}

func healthcheck(ctx context.Context, c *gnomock.Container) error {
	db, err := connect(c, "")
	if err != nil {
		return err
	}

	defer func() {
		_ = db.Close()
	}()

	var one int

	return db.QueryRow(`select 1`).Scan(&one)
}

func (p *P) initf() gnomock.InitFunc {
	return func(ctx context.Context, c *gnomock.Container) error {
		db, err := connect(c, "")
		if err != nil {
			return err
		}

		_, err = db.Exec("create database " + p.DB)
		if err != nil {
			return err
		}

		_ = db.Close()

		db, err = connect(c, p.DB)
		if err != nil {
			return err
		}

		defer func() { _ = db.Close() }()

		if len(p.QueriesFiles) > 0 {
			for _, f := range p.QueriesFiles {
				bs, err := ioutil.ReadFile(f) // nolint:gosec
				if err != nil {
					return fmt.Errorf("can't read queries file '%s': %w", f, err)
				}

				p.Queries = append([]string{string(bs)}, p.Queries...)
			}
		}

		for _, q := range p.Queries {
			_, err = db.Exec(q)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func connect(c *gnomock.Container, db string) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d sslmode=disable user=root dbname=%s",
		c.Host, c.Port(gnomock.DefaultPort), db,
	)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return conn, conn.Ping()
}
