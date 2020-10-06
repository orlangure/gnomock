<div align="center">
    <img src="https://github.com/orlangure/gnomock/blob/master/gnomock.png">
</div>

# <div align="center">Gnomock – tests without mocks</div>

🏗️ Spin up entire dependency stack

🎁 Setup initial dependency state – easily!

🏭 Test against actual, close to production software

⏳ Spend no time writing mocks

🕹️ Test actual program behavior and side effects

## <div align="center">[![PkgGoDev](https://pkg.go.dev/badge/github.com/orlangure/gnomock)](https://pkg.go.dev/github.com/orlangure/gnomock) ![Build](https://github.com/orlangure/gnomock/workflows/Build/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/orlangure/gnomock)](https://goreportcard.com/report/github.com/orlangure/gnomock)</div>

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
  - [Using Gnomock server](#using-gnomock-server)
- [Official presets](#official-presets)
- [Similar projects](#similar-projects)
- [Troubleshooting](#troubleshooting)

## Getting started

Gnomock can be used in two different ways:

- Imported directly as a package in any **Go** project
- Accessed over HTTP running as a daemon in any other project

⚠️ Both ways **require** an active Docker daemon running locally in
the same environment.

### Using Gnomock in Go applications

Gnomock can be used in Go programs directly, without running a local server.
See the following example to get started:

```
go get github.com/orlangure/gnomock
```

Setting up a **Postgres** container with schema setup example:

```go
import (
	"database/sql"

	_ "github.com/lib/pq" // postgres driver
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/postgres"
)

p := postgres.Preset(
    postgres.WithUser("gnomock", "gnomick"),
    postgres.WithDatabase("mydb"),
    postgres.WithQueriesFile("/var/project/db/schema.sql"),
)
container, _ := gnomock.Start(p)
defer func() { _ = gnomock.Stop(container) }()

connStr := fmt.Sprintf(
    "host=%s port=%d user=%s password=%s  dbname=%s sslmode=disable",
    container.Host, container.DefaultPort(),
    "gnomock", "gnomick", "mydb",
)
db, _ := sql.Open("postgres", connStr)
// db has the required schema and data, and is ready to use
```

See package
[reference](https://pkg.go.dev/github.com/orlangure/gnomock?tab=doc). For
Preset documentation, refer to [Presets](#official-presets) section.

### Using Gnomock server

To start a `gnomock` server, run the following on any Unix-based system:

```bash
docker run --rm \
    -p 23042:23042 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v `pwd`:`pwd` \
    orlangure/gnomock
```

`-p 23042:23042` exposes a port on the host to communicate with `gnomock`. You
can use any port you like, just make sure to configure the client properly.

`-v /var/run/docker.sock:/var/run/docker.sock` allows `gnomock` to communicate
with the docker engine running on host. Without it `gnomock` can't access
docker.

If you use any file-related `gnomock` options, like `WithQueriesFile`, you have
to make the path you use available inside the container:

```
# this makes the current folder appear inside the container under the same
# path and name:
-v `pwd`:`pwd`
```

Any program in any language can communicate with `gnomock` server using OpenAPI
3.0 [specification](https://app.swaggerhub.com/apis/orlangure/gnomock/).

Below is an example of setting up a **MySQL** container using a `POST` request:

```
$ cat mysql-preset.json
{
  "preset": {
    "db": "mydb",
    "user": "gnomock",
    "password": "p@s$w0rD",
    "queries": [
      "create table foo(bar int)",
      "insert into foo(bar) values(1)"
    ],
    "queries_file": "/home/gnomock/project/testdata/mysql/queries.sql"
  },
  "options": {}
}

$ curl --data @mysql-preset.json http://127.0.0.1:23042/start/mysql
{
  "id": "f5d08dc84421",
  "host": "string",
  "ports": {
    "default": {
      "protocol": "tcp",
      "port": 35973
    }
  }
}
```

There are auto-generated wrappers for the available API:

| Client | Sample code |
|--------|-------------|
| [Python SDK](https://github.com/orlangure/gnomock-python-sdk) | [Code](https://github.com/orlangure/gnomock/blob/master/sdktest/python/test/test_sdk.py) |
| JavaScript SDK | |
| Ruby SDK | |
| PHP SDK | |
| Java SDK | |
| [Other](https://openapi-generator.tech/docs/generators) languages | |

**For more details and a full specification, see
[documentation](https://app.swaggerhub.com/apis/orlangure/gnomock/).**

## Official presets

The power of Gnomock is in the Presets. Presets, both existing and planned, are
listed below:

| Preset | Go package | HTTP API | Go API | Latest stable image |
|--------|------------|----------|--------|---------------------|
Localstack | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/localstack) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.4.5#/presets/startLocalstack) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/localstack?tab=doc) | 0.11.0
Splunk | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/splunk) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.4.5#/presets/startSplunk) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/splunk?tab=doc) | 8.0.2
Redis | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/redis) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.4.5#/presets/startRedis) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/redis?tab=doc) | 5
Memcached | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/memcached) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.4.5#/presets/startMemcached) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/memcached?tab=doc) | 1.6.6-alpine
MySQL | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/mysql) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.4.5#/presets/startMysql) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mysql?tab=doc) | 8
MariaDB | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/mariadb) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.4.5#/presets/startMariadb) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mariadb?tab=doc) | 10
PostgreSQL | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/postgres) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.4.5#/presets/startPostgres) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/postgres?tab=doc) | 12
Microsoft SQL Server | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/mssql) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.4.5#/presets/startMssql) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mssql?tab=doc) | 2019-latest
MongoDB | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/mongo) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.4.5#/presets/startMongo) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mongo?tab=doc) | 4
RabbitMQ | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/rabbitmq) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.4.5#/presets/startRabbitMq) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/rabbitmq?tab=doc) | 3.8.5-alpine
Kafka | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/kafka) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.4.5#/presets/startKafka) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/kafka?tab=doc) | 2.5.1-L0
Elasticsearch | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/elastic) | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.4.5#/presets/startElastic) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/elastic?tab=doc) | 7.9.2
DynamoDB | |
Cassandra | |

It is possible to use Gnomock directly from Go code without any presets. HTTP
API only allows to setup containers using presets that exist in this
repository.

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
