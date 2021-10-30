<div align="center">
    <img src="https://github.com/orlangure/gnomock/raw/master/gnomock.png">
</div>

# <div align="center">Gnomock – tests without mocks</div>

🏗️ Spin up entire dependency stack

🎁 Setup initial dependency state – easily!

🏭 Test against actual, close to production software

⏳ Spend no time writing mocks

🕹️ Test actual program behavior and side effects

## <div align="center">[![PkgGoDev](https://pkg.go.dev/badge/github.com/orlangure/gnomock)](https://pkg.go.dev/github.com/orlangure/gnomock) ![Test](https://github.com/orlangure/gnomock/workflows/Test/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/orlangure/gnomock)](https://goreportcard.com/report/github.com/orlangure/gnomock) [![codecov](https://codecov.io/gh/orlangure/gnomock/branch/master/graph/badge.svg?token=F0XYPSEIMK)](https://codecov.io/gh/orlangure/gnomock)</div>

Gnomock is an integration and end-to-end testing toolkit. It uses Docker to
create **temporary containers** for application dependencies, **setup** their
**initial state** and **clean them up** in the end. Gnomock allows to test the
code with **no mocks** wherever possible.

The power of Gnomock is in a variety of [Presets](#official-presets), each
implementing a specific database, service or other tools. Each preset provides
ways of setting up its initial state as easily as possible: SQL schema
creation, test data upload into S3, sending test events to Splunk, etc.

The name "Gnomock" stands for "no mock", with a "G" for "Go" 😼. It also sounds
like "gnome", that's why the friendly garden gnome artwork (by [Michael
Zolotov](https://www.mzolotov.com/))

## Demo

See for yourself how easy and fast it is to write tests that use actual
services running in ephemeral Docker containers:

[![asciicast](https://asciinema.org/a/jSJJGk0n3q1g4Fc4bvZYYyifN.svg)](https://asciinema.org/a/jSJJGk0n3q1g4Fc4bvZYYyifN)

## Table of contents

- [Getting started](#getting-started)
  - [Using Gnomock in Go applications](#using-gnomock-in-go-applications)
  - [Using Gnomock in other languages](#using-gnomock-in-other-languages)
- [Official presets](#official-presets)
- [Similar projects](#similar-projects)
- [Troubleshooting](#troubleshooting)

## Getting started

Gnomock can be used in two different ways:

- Imported directly as a package in any **Go** project
- Accessed over HTTP running as a daemon in **any other language**

Both ways **require** an active Docker daemon running **locally** in the same
environment.

External `DOCKER_HOST` support is experimental. It cannot be reliably tested at
this moment, but it might work.

### Using Gnomock in Go applications

See the following example to get started:

```
go get github.com/orlangure/gnomock
```

Setting up a **Postgres** container with schema setup example:

```go
package db_test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq" // postgres driver
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/postgres"
)

func TestDB(t *testing.T) {
	p := postgres.Preset(
		postgres.WithUser("gnomock", "gnomick"),
		postgres.WithDatabase("mydb"),
		postgres.WithQueriesFile("/var/project/db/schema.sql"),
	)
	container, _ := gnomock.Start(p)
	t.Cleanup(func() { _ = gnomock.Stop(container) })

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s  dbname=%s sslmode=disable",
		container.Host, container.DefaultPort(),
		"gnomock", "gnomick", "mydb",
	)
	db, _ := sql.Open("postgres", connStr)
	// db has the required schema and data, and is ready to use
}
```

See package
[reference](https://pkg.go.dev/github.com/orlangure/gnomock?tab=doc). For
Preset documentation, refer to [Presets](#official-presets) section.

### Using Gnomock in other languages

If you use Go, please refer to [Using Gnomock in Go
applications](#using-gnomock-in-go-applications) section. Otherwise, refer to
[documentation](docs/server.md).

## Official presets

The power of Gnomock is in the Presets. Existing Presets with their
supported<sup>\*</sup> versions are listed below.

<small>*\* **Supported** versions are tested as part of CI pipeline. Other
versions might work as well.*</small>


| Preset | Go package | HTTP API | Go API | Supported versions |
|--------|------------|----------|--------|---------------------|
Localstack (AWS) | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/localstack) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startLocalstack) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/localstack?tab=doc) | `0.12.2`
Splunk | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/splunk) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startSplunk) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/splunk?tab=doc) | `8.0.2`
Redis | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/redis) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startRedis) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/redis?tab=doc) | `5.0.10`, `6.0.9`
Memcached | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/memcached) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startMemcached) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/memcached?tab=doc) | `1.6.9`
MySQL | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/mysql) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startMysql) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mysql?tab=doc) | `5.7.32`, `8.0.22`
MariaDB | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/mariadb) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startMariadb) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mariadb?tab=doc) | `10.5.8`
PostgreSQL | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/postgres) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startPostgres) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/postgres?tab=doc) |  `10.15`, `11.10`, `12.5`, `13.1`
Microsoft SQL Server | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/mssql) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startMssql) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mssql?tab=doc) | `2017-latest`, `2019-latest`
MongoDB | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/mongo) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startMongo) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mongo?tab=doc) | `3.6.21`, `4.4`
RabbitMQ | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/rabbitmq) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startRabbitMq) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/rabbitmq?tab=doc) | `3.8.9-alpine`, `3.8.9-management-alpine`
Kafka | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/kafka) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startKafka) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/kafka?tab=doc) | `2.5.1-L0`
Elasticsearch | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/elastic) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startElastic) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/elastic?tab=doc) | `5.6`, `6.8.13`, `7.9.3`
Kubernetes | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/k3s) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startKubernetes) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/k3s?tab=doc) | `v1.19.12`
CockroachDB | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/cockroachdb) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startCockroachDB) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/cockroachdb?tab=doc) | `v19.2.11`, `v20.1.10`
InfluxDB | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/influxdb) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startInfluxDB) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/influxdb?tab=doc) | `2.0.4-alpine`
Cassandra | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/cassandra) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.7.0#/presets/startCassandra) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/cassandra?tab=doc) | `4.0`, `3`
<!-- new presets go here -->

It is possible to use Gnomock [directly from
Go](https://pkg.go.dev/github.com/orlangure/gnomock#StartCustom) code without
any presets. HTTP API only allows to setup containers using presets that exist
in this repository.

## Similar projects

Gnomock is not the only project that aims to simplify integration and
end-to-end testing by using ephemeral docker containers:

- `testcontainers/testcontainers-go`
- `ory/dockertest`

These projects are amazing, and they give plenty of flexibility and power to
their users. There are many things that are possible with them, but are
impossible with Gnomock. Still, below is a short list of things that sometimes
give Gnomock an advantage:

- **Gnomock tries to provide a batteries-included solution**. Gnomock has a
  growing number of Presets, each one implementing an integration with a
  popular external service. For every Preset, there already is a number of
  "invisible" utilities that transparently relieve you from implementing them
  yourself:
  - __Built-in health check function__ that you don't even need to know it
    exists. It makes sure you only get control over a container when it is
    ready to use.
  - __Wrappers for some of the configuration__ exposed by the container, such
    as default username/password. You can easily provide your own credentials
    to connect to the container.
  - __Seed data ingestion__ for your convenience. Sometimes you just need to
    make sure your queries work given some data. Gnomock puts your data in
    there with a single line of code. Sometimes you only test a program that
    consumes messages from Kafka, and Gnomock produces the messages for you
    with another line of code.
- **Simple API** that does not expose anything that happens "under the hood"
  most of the time. Yet Gnomock allows some additional configuration and custom
  Preset implementation whenever necessary.
- Gnomock's vision includes **being useful not only in Go** projects, but in
  any projects via HTTP. It already supports almost all its features over HTTP
  layer, has a clear OpenAPI spec, and even a proof of concept wrapper in
  Python.
- Gnomock has a friendly **garden gnome mascot**😻

## Troubleshooting

### Tests with Gnomock take too long and time-out eventually

It happens a lot locally if your internet isn't fast enough to pull docker
images used in tests. In CI, such as in Github Actions, the images are
downloaded very quickly. To work around this issue locally, pull the image
manually before running the tests. You only need to do it once, the images stay
in local cache until deleted. For example, to pull Postgres 11 image, run:

```
docker pull postgres:11
```

### Tests time-out even when the image exists locally

It can happen if the containers can't become ready to use before they time out.
By default, Gnomock uses fairly high timeouts for new containers (for starting
and for setting them up). If you choose to change default timeout using
`WithTimeout` (`timeout` in HTTP), it is possible that the values you choose
are too short.

### Tests pass when run one-by-one, and fail when run in parallel

It happens when you try to start up **a lot** of containers at the same time.
The system, especially in CI environments such as Github Actions, cannot handle
the load, and containers fail to become healthy before they time-out. That's
the reason Gnomock has a few separate build jobs, each running only a small
subset of tests, one package at a time.

### Containers fail to setup with a "File not found" error

If you run `gnomock` as a server, you need to make sure the files you use in
your setup are available inside `gnomock` container. Use `-v $(pwd):$(pwd)`
argument to `docker run` to mount the current working directory under the same
path inside the `gnomock` container. If you prefer to keep a permanent
`gnomock` container running, you can mount your entire `$HOME` directory (or
any other directory where you keep the code).
