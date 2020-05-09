// Package preset provides access to existing presets
package preset

import (
	localstack "github.com/orlangure/gnomock-localstack"
	mongo "github.com/orlangure/gnomock-mongo"
	mssql "github.com/orlangure/gnomock-mssql"
	mysql "github.com/orlangure/gnomock-mysql"
	postgres "github.com/orlangure/gnomock-postgres"
	redis "github.com/orlangure/gnomock-redis"
	splunk "github.com/orlangure/gnomock-splunk"
	"github.com/orlangure/gnomock/gnomock"
)

// Preseter is a type that creates Gnomock Presets given their name
type Preseter interface {
	Preset(string) gnomock.Preset
}

type registry struct{}

func (r registry) Preset(name string) gnomock.Preset {
	switch name {
	case "localstack":
		return &localstack.P{}
	case "mongo":
		return &mongo.P{}
	case "mssql":
		return &mssql.P{}
	case "mysql":
		return &mysql.P{}
	case "postgres":
		return &postgres.P{}
	case "redis":
		return &redis.P{}
	case "splunk":
		return &splunk.P{}
	}

	return nil
}

// Registry returns a default, inline Preseter. Its entire implementation is
// hard-coded: every new Preset must be added to the implementation and
// compiled again
func Registry() Preseter {
	return registry{}
}
