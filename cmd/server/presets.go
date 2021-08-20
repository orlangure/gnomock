package main

// all known presets should go right here so that they are available when
// requested over HTTP:
import (
	_ "github.com/orlangure/gnomock/preset/cassandra"
	_ "github.com/orlangure/gnomock/preset/cockroachdb"
	_ "github.com/orlangure/gnomock/preset/elastic"
	_ "github.com/orlangure/gnomock/preset/influxdb"
	_ "github.com/orlangure/gnomock/preset/k3s"
	_ "github.com/orlangure/gnomock/preset/kafka"
	_ "github.com/orlangure/gnomock/preset/localstack"
	_ "github.com/orlangure/gnomock/preset/mariadb"
	_ "github.com/orlangure/gnomock/preset/memcached"
	_ "github.com/orlangure/gnomock/preset/mongo"
	_ "github.com/orlangure/gnomock/preset/mssql"
	_ "github.com/orlangure/gnomock/preset/mysql"
	_ "github.com/orlangure/gnomock/preset/postgres"
	_ "github.com/orlangure/gnomock/preset/rabbitmq"
	_ "github.com/orlangure/gnomock/preset/redis"
	_ "github.com/orlangure/gnomock/preset/splunk"
	// new presets go here
)
