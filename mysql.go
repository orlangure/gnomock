// Package mysql provides a Gnomock Preset for MySQL database
package mysql

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/go-sql-driver/mysql" // mysql driver
	"github.com/orlangure/gnomock"
)

// Preset creates a new Gmomock MySQL preset. This preset includes a MySQL
// specific healthcheck function, default MySQL image and port, and allows to
// optionally set up initial state. When used without any configuration, it
// creates a superuser "gnomock" with password "gnomick", and "mydb" database
func Preset(opts ...Option) *MySQL {
	// err is always nil for non-nil logger
	_ = mysql.SetLogger(log.New(ioutil.Discard, "", -1))

	config := buildConfig(opts...)

	p := &MySQL{
		db:       config.db,
		queries:  config.queries,
		user:     config.user,
		password: config.password,
	}

	return p
}

// MySQL is a Gnomock Preset implementation for MySQL database
type MySQL struct {
	db       string
	user     string
	password string
	queries  []string
}

// Image returns an image that should be pulled to create this container
func (p *MySQL) Image() string {
	return "docker.io/library/mysql"
}

// Ports returns ports that should be used to access this container
func (p *MySQL) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(defaultPort)
}

// Options returns a list of options to configure this container
func (p *MySQL) Options() []gnomock.Option {
	opts := []gnomock.Option{
		gnomock.WithHealthCheck(p.healthcheck),
		gnomock.WithEnv("MYSQL_USER=" + p.user),
		gnomock.WithEnv("MYSQL_PASSWORD=" + p.password),
		gnomock.WithEnv("MYSQL_DATABASE=" + p.db),
		gnomock.WithEnv("MYSQL_RANDOM_ROOT_PASSWORD=" + p.db),
		gnomock.WithInit(p.initf(p.queries)),
		gnomock.WithWaitTimeout(time.Second * 30),
	}

	return opts
}

func (p *MySQL) healthcheck(c *gnomock.Container) error {
	addr := c.Address(gnomock.DefaultPort)

	db, err := p.connect(addr)
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

func (p *MySQL) initf(queries []string) gnomock.InitFunc {
	return func(c *gnomock.Container) error {
		addr := c.Address(gnomock.DefaultPort)

		db, err := p.connect(addr)
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

func (p *MySQL) connect(addr string) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s",
		p.user, p.password, addr, p.db,
	)

	return sql.Open("mysql", connStr)
}
