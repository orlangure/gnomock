// Package mssql provides a Gnomock Preset for Microsoft SQL Server database
package mssql

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/denisenkom/go-mssqldb" // mssql driver
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
)

const (
	masterDB        = "master"
	defaultPassword = "Gn0m!ck~"
	defaultDatabase = "mydb"
	defaultPort     = 1433
	defaultVersion  = "2019-latest"
)

func init() {
	registry.Register("mssql", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gmomock Microsoft SQL Server preset. This preset
// includes a mssql specific healthcheck function, default mssql image and
// port, and allows to optionally set up initial state.
//
// When used without any configuration, it uses `mydb` database, and `Gn0m!ck~`
// administrator password (user: `sa`). You must accept EULA to use this image
// (`WithLicense` option). By default, version `2019-latest` is used.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation of Microsoft SQL Server database.
type P struct {
	DB           string   `json:"db"`
	Password     string   `json:"password"`
	Queries      []string `json:"queries"`
	QueriesFiles []string `json:"queries_files"`
	License      bool     `json:"license"`
	Version      string   `json:"version"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("mcr.microsoft.com/mssql/server:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(defaultPort)
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(p.healthcheck),
		gnomock.WithEnv("SA_PASSWORD=" + p.Password),
		gnomock.WithInit(p.initf()),
	}

	if p.License {
		opts = append(opts, gnomock.WithEnv("ACCEPT_EULA=Y"))
	}

	return opts
}

func (p *P) healthcheck(_ context.Context, c *gnomock.Container) error {
	addr := c.Address(gnomock.DefaultPort)

	db, err := p.connect(addr, masterDB)
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
		addr := c.Address(gnomock.DefaultPort)

		db, err := p.connect(addr, masterDB)
		if err != nil {
			return err
		}

		defer func() { _ = db.Close() }()

		_, err = db.Exec("create database " + p.DB)
		if err != nil {
			return fmt.Errorf("can't create database '%s': %w", p.DB, err)
		}

		db, err = p.connect(addr, p.DB)
		if err != nil {
			return err
		}

		if len(p.QueriesFiles) > 0 {
			for _, f := range p.QueriesFiles {
				bs, err := os.ReadFile(f) // nolint:gosec
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

func (p *P) connect(addr, db string) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"sqlserver://sa:%s@%s?database=%s",
		p.Password, addr, db,
	)

	return sql.Open("sqlserver", connStr)
}

func (p *P) setDefaults() {
	if p.DB == "" {
		p.DB = defaultDatabase
	}

	if p.Password == "" {
		p.Password = defaultPassword
	}

	if p.Version == "" {
		p.Version = defaultVersion
	}
}
