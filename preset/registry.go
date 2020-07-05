// Package preset provides access to existing presets
package preset

import (
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/localstack"
	"github.com/orlangure/gnomock/preset/mongo"
	"github.com/orlangure/gnomock/preset/mssql"
	"github.com/orlangure/gnomock/preset/mysql"
	"github.com/orlangure/gnomock/preset/postgres"
	"github.com/orlangure/gnomock/preset/rabbitmq"
	"github.com/orlangure/gnomock/preset/redis"
	"github.com/orlangure/gnomock/preset/splunk"
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
	case "rabbitmq":
		return &rabbitmq.P{}
	}

	return nil
}

// Registry returns a default, inline Preseter. Its entire implementation is
// hard-coded: every new Preset must be added to the implementation and
// compiled again
func Registry() Preseter {
	return registry{}
}
