// Package mssql provides a Gnomock Preset for Microsoft SQL Server database
package mssql

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/denisenkom/go-mssqldb" // mssql driver
	"github.com/orlangure/gnomock"
)

const masterDB = "master"
const defaultPassword = "Gn0m!ck~"
const defaultDatabase = "mydb"
const defaultPort = 1433

// Preset creates a new Gmomock Microsoft SQL Server preset. This preset
// includes a mssql specific healthcheck function, default mssql image and
// port, and allows to optionally set up initial state. When used without any
// configuration, it uses "mydb" database, and "Gn0m!ck~" administrator
// password (user: sa). You must accept EULA to use this image (WithLicense
// option)
func Preset(opts ...Option) gnomock.Preset {
	p := &P{
		DB:       defaultDatabase,
		Password: defaultPassword,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation of Microsoft SQL Server database
type P struct {
	DB       string   `json:"db"`
	Password string   `json:"password"`
	Queries  []string `json:"queries"`
	License  bool     `json:"license"`
}

// Image returns an image that should be pulled to create this container
func (p *P) Image() string {
	return "mcr.microsoft.com/mssql/server"
}

// Ports returns ports that should be used to access this container
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(defaultPort)
}

// Options returns a list of options to configure this container
func (p *P) Options() []gnomock.Option {
	opts := []gnomock.Option{
		gnomock.WithHealthCheck(p.healthcheck),
		gnomock.WithEnv("SA_PASSWORD=" + p.Password),
		gnomock.WithInit(p.initf(p.Queries)),
		gnomock.WithWaitTimeout(time.Second * 30),
	}

	if p.License {
		opts = append(opts, gnomock.WithEnv("ACCEPT_EULA=Y"))
	}

	return opts
}

func (p *P) healthcheck(c *gnomock.Container) error {
	addr := c.Address(gnomock.DefaultPort)

	db, err := p.connect(addr, masterDB)
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

func (p *P) initf(queries []string) gnomock.InitFunc {
	return func(c *gnomock.Container) error {
		addr := c.Address(gnomock.DefaultPort)

		db, err := p.connect(addr, masterDB)
		if err != nil {
			return err
		}

		_, err = db.Exec("create database " + p.DB)
		if err != nil {
			return fmt.Errorf("can't create database '%s': %w", p.DB, err)
		}

		db, err = p.connect(addr, p.DB)
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

func (p *P) connect(addr, db string) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"sqlserver://sa:%s@%s?database=%s",
		p.Password, addr, db,
	)

	return sql.Open("sqlserver", connStr)
}
