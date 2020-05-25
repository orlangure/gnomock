// Package mysql provides a Gnomock Preset for MySQL database
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/orlangure/gnomock"
)

const defaultUser = "gnomock"
const defaultPassword = "gnomick"
const defaultDatabase = "mydb"
const defaultPort = 3306

// Preset creates a new Gmomock MySQL preset. This preset includes a MySQL
// specific healthcheck function, default MySQL image and port, and allows to
// optionally set up initial state. When used without any configuration, it
// creates a superuser "gnomock" with password "gnomick", and "mydb" database
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation of MySQL database
type P struct {
	DB          string   `json:"db"`
	User        string   `json:"user"`
	Password    string   `json:"password"`
	Queries     []string `json:"queries"`
	QueriesFile string   `json:"queries_file"`
}

// Image returns an image that should be pulled to create this container
func (p *P) Image() string {
	return "docker.io/library/mysql"
}

// Ports returns ports that should be used to access this container
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(defaultPort)
}

// Options returns a list of options to configure this container
func (p *P) Options() []gnomock.Option {
	// err is always nil for non-nil logger
	_ = mysqldriver.SetLogger(log.New(ioutil.Discard, "", -1))

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

func (p *P) initf() gnomock.InitFunc {
	return func(ctx context.Context, c *gnomock.Container) error {
		addr := c.Address(gnomock.DefaultPort)

		db, err := p.connect(addr)
		if err != nil {
			return err
		}

		defer func() {
			_ = db.Close()
		}()

		if p.QueriesFile != "" {
			bs, err := ioutil.ReadFile(p.QueriesFile)
			if err != nil {
				return fmt.Errorf("can't read queries file '%s': %w", p.QueriesFile, err)
			}

			p.Queries = append([]string{string(bs)}, p.Queries...)
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
}
