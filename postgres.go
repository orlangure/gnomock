// Package postgres provides a Gnomock Preset for PostgreSQL database
package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq" // postgres driver
	"github.com/orlangure/gnomock"
)

// Preset creates a new Gmomock Postgres preset. This preset includes a Postgres
// specific healthcheck function, default Postgres image and port, and allows to
// optionally set up initial state
func Preset(opts ...Option) *Postgres {
	config := buildConfig(opts...)

	p := &Postgres{
		db:      config.db,
		queries: config.queries,
	}

	return p
}

// Postgres is a Gnomock Preset implementation for PostgreSQL database
type Postgres struct {
	db      string
	queries []string
}

// Image returns an image that should be pulled to create this container
func (p *Postgres) Image() string {
	return "docker.io/library/postgres"
}

// Ports returns ports that should be used to access this container
func (p *Postgres) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(defaultPort)
}

// Options returns a list of options to configure this container
func (p *Postgres) Options() []gnomock.Option {
	opts := []gnomock.Option{
		gnomock.WithHealthCheck(p.healthcheck),
		gnomock.WithEnv("POSTGRES_PASSWORD=" + defaultPassword),
		gnomock.WithInit(p.initf(p.queries)),
	}

	return opts
}

func (p *Postgres) healthcheck(c *gnomock.Container) error {
	addr := c.Address(gnomock.DefaultPort)

	db, err := connect(addr, defaultDatabase)
	if err != nil {
		return err
	}

	var one int

	row := db.QueryRow(`select 1`)

	err = row.Scan(&one)
	if err != nil {
		return err
	}

	if one != 1 {
		return fmt.Errorf("unexpected healthcheck result: 1 != %d", one)
	}

	return nil
}

func (p *Postgres) initf(queries []string) gnomock.InitFunc {
	return func(c *gnomock.Container) error {
		addr := c.Address(gnomock.DefaultPort)

		if p.db != defaultDatabase {
			db, err := connect(addr, defaultDatabase)
			if err != nil {
				return err
			}

			_, err = db.Exec("create database " + p.db)
			if err != nil {
				return err
			}
		}

		db, err := connect(addr, p.db)
		if err != nil {
			return err
		}

		for _, q := range queries {
			_, err = db.Exec(q)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func connect(addr, db string) (*sql.DB, error) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid address '%s', must be 'host:port'", addr)
	}

	host, port := parts[0], parts[1]

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, defaultUser, defaultPassword, db, defaultSSLMode,
	)

	return sql.Open("postgres", connStr)
}
