// Package postgres provides a Gnomock Preset for PostgreSQL database.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"

	_ "github.com/lib/pq" // postgres driver
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
)

const defaultUser = "postgres"
const defaultPassword = "password"
const defaultDatabase = "postgres"
const defaultSSLMode = "disable"
const defaultPort = 5432
const defaultVersion = "12.5"

func init() {
	registry.Register("postgres", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gmomock Postgres preset. This preset includes a Postgres
// specific healthcheck function, default Postgres image and port, and allows to
// optionally set up initial state.
//
// By default, this preset uses `postgres` user with `password` password, with
// default database `postgres`. Default PostgresQL version is 12.5.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation of PostgreSQL database
type P struct {
	DB           string   `json:"db"`
	Queries      []string `json:"queries"`
	QueriesFiles []string `json:"queries_files"`
	User         string   `json:"user"`
	Password     string   `json:"password"`
	Version      string   `json:"version"`
}

// Image returns an image that should be pulled to create this container
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/library/postgres:%s", p.Version)
}

// Ports returns ports that should be used to access this container
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(defaultPort)
}

// Options returns a list of options to configure this container
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	if p.User != "" && p.Password != "" {
		q := fmt.Sprintf(
			`create user %s with superuser password '%s'`,
			p.User, p.Password,
		)

		p.Queries = append(p.Queries, q)
	}

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(p.healthcheck),
		gnomock.WithEnv("POSTGRES_PASSWORD=" + defaultPassword),
		gnomock.WithInit(p.initf()),
	}

	return opts
}

func (p *P) healthcheck(ctx context.Context, c *gnomock.Container) error {
	db, err := connect(c, defaultDatabase)
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
		if p.DB != defaultDatabase {
			db, err := connect(c, defaultDatabase)
			if err != nil {
				return err
			}

			_, err = db.Exec("create database " + p.DB)
			if err != nil {
				return err
			}

			_ = db.Close()
		}

		db, err := connect(c, p.DB)
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

func (p *P) setDefaults() {
	if p.DB == "" {
		p.DB = defaultDatabase
	}

	if p.Version == "" {
		p.Version = defaultVersion
	}
}

func connect(c *gnomock.Container, db string) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port(gnomock.DefaultPort),
		defaultUser, defaultPassword, db, defaultSSLMode,
	)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return conn, conn.Ping()
}
