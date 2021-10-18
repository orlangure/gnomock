// Package mariadb provides a Gnomock Preset for MariaDB database
package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
)

const (
	defaultUser     = "gnomock"
	defaultPassword = "gnoria"
	defaultDatabase = "mydb"
	defaultPort     = 3306
	defaultVersion  = "10.5.8"
)

var setLoggerOnce sync.Once

func init() {
	registry.Register("mariadb", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gmomock MariaDB preset. This preset includes a MariaDB
// specific healthcheck function, default MariaDB image and port, and allows to
// optionally set up initial state.
//
// When used without any configuration, it creates a superuser `gnomock` with
// password `gnoria`, and `mydb` database. Default MariaDB version is 10.5.8.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation of MariaDB database.
type P struct {
	DB           string   `json:"db"`
	User         string   `json:"user"`
	Password     string   `json:"password"`
	Queries      []string `json:"queries"`
	QueriesFiles []string `json:"queries_files"`
	Version      string   `json:"version"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/library/mariadb:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(defaultPort)
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	setLoggerOnce.Do(func() {
		// err is always nil for non-nil logger
		_ = mysqldriver.SetLogger(log.New(io.Discard, "", -1))
	})

	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(p.healthcheck),
		gnomock.WithEnv("MYSQL_USER=" + p.User),
		gnomock.WithEnv("MYSQL_PASSWORD=" + p.Password),
		gnomock.WithEnv("MYSQL_DATABASE=" + p.DB),
		gnomock.WithEnv("MYSQL_RANDOM_ROOT_PASSWORD=yes"),
		gnomock.WithInit(p.initf()),
	}

	return opts
}

func (p *P) healthcheck(ctx context.Context, c *gnomock.Container) error {
	addr := c.Address(gnomock.DefaultPort)

	db, err := p.connect(addr)
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

		db, err := p.connect(addr)
		if err != nil {
			return err
		}

		defer func() { _ = db.Close() }()

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

func (p *P) connect(addr string) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?multiStatements=true",
		p.User, p.Password, addr, p.DB,
	)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}

	return db, db.Ping()
}

func (p *P) setDefaults() {
	if p.DB == "" {
		p.DB = defaultDatabase
	}

	if p.User == "" {
		p.User = defaultUser
	}

	if p.Password == "" {
		p.Password = defaultPassword
	}

	if p.Version == "" {
		p.Version = defaultVersion
	}
}
