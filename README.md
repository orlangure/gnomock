# <div align="center">Gnomock – test your code without mocks</div>

🏗️ Spin up entire dependency stack

🎁 Setup initial dependency state – easily!

🏭 Test against actual, close to production software

⏳ Spend no time writing mocks

🕹️ Test actual program behavior and side effects

## <div align="center">![Build](https://github.com/orlangure/gnomock/workflows/Build/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/orlangure/gnomock)](https://goreportcard.com/report/github.com/orlangure/gnomock)</div>

Gnomock is an integration and end-to-end testing toolkit. It uses Docker to
create **temporary containers** for application dependencies, **setup** their
**initial state** and **clean them up** in the end. Gnomock allows to test the
code with **no mocks** wherever possible.

The power of Gnomock is in a variety of [Presets](#official-presets), each
implementing a specific database, service or other tools. Each preset provides
ways of setting up its initial state as easily as possible: SQL schema
creation, test data upload into S3, sending test events to Splunk, etc.

## Overview

Gnomock runs locally and exposes an API over HTTP. This API is defined using
OpenAPI 3.0
[specification](https://app.swaggerhub.com/apis/orlangure/gnomock/1.0.1). Go
programs can use extended Gnomock package directly, without HTTP layer, while
other languages require communication with a local server.

Gnomock **requires** a running and working Docker daemon running locally in the
same environment.

### Using Gnomock in Go applications

Gnomock can be used in Go programs directly, without running a local server.
See the following example to get started:

```
go get github.com/orlangure/gnomock
```

Setting up a Redis container example:

```go
import (
	redisclient "github.com/go-redis/redis/v7"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/redis"
)

p := redis.Preset()
container, _ := gnomock.Start(p)

defer func() { _ = gnomock.Stop(container) }()

addr := container.DefaultAddress()
client := redisclient.NewClient(&redisclient.Options{Addr: addr})
```

See package [reference](https://pkg.go.dev/github.com/orlangure/gnomock?tab=doc).

For Preset documentation, refer to [Presets](#official-presets) section.

### Using Gnomock server

Gnomock runs as a local server, and any program in any language can communicate
with it using OpenAPI 3.0
[specification](https://app.swaggerhub.com/apis/orlangure/gnomock/1.0.1).

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

- [Python SDK](https://github.com/orlangure/gnomock-python-sdk)
- JavaScript SDK
- PHP SDK
- Ruby SDK
- Java SDK
- [Other](https://openapi-generator.tech/docs/generators) languages

**For more details and a full specification, see
[documentation](https://app.swaggerhub.com/apis/orlangure/gnomock/1.0.1).**

Installation instruction, as well as pre-compiled binaries for MacOS and Linux,
are coming soon.

## Official presets

The power of Gnomock is in the Presets. Presets, both existing and planned, are
listed below:

| Preset | Go package | HTTP API | Go API |
|--------|------------|----------|-----------|
Localstack | https://github.com/orlangure/gnomock/tree/master/preset/localstack | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.0.1#/presets/startLocalstack) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/localstack?tab=doc)
Splunk | https://github.com/orlangure/gnomock/tree/master/preset/splunk | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.0.1#/presets/startSplunk) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/splunk?tab=doc)
Redis | https://github.com/orlangure/gnomock/tree/master/preset/redis | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.0.1#/presets/startRedis) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/redis?tab=doc)
MySQL | https://github.com/orlangure/gnomock/tree/master/preset/mysql | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.0.1#/presets/startMysql) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mysql?tab=doc)
PostgreSQL | https://github.com/orlangure/gnomock/tree/master/preset/postgres | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.0.1#/presets/startPostgres) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/postgres?tab=doc)
Microsoft SQL Server | https://github.com/orlangure/gnomock/tree/master/preset/mssql | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.0.1#/presets/startMssql) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mssql?tab=doc)
MongoDB | https://github.com/orlangure/gnomock/tree/master/preset/mongo | [Docs](https://app.swaggerhub.com/apis/orlangure/gnomock/1.0.1#/presets/startMongo) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mongo?tab=doc)
Elasticsearch | |
DynamoDB | |
Cassandra | |
MariaDB | |

It is possible to use Gnomock directly from Go code without any presets. HTTP
API only allows to setup containers using presets that exist in this
repository.

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
