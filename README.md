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

## Documentation

Gnomock runs locally and exposes an API over HTTP. This API is defined using
OpenAPI 3.0
[specification](https://app.swaggerhub.com/apis/orlangure/gnomock/1.0.1). Go
programs can use extended Gnomock package directly, without HTTP layer, while
other languages require communication with a local server.

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

### Using Gnomock in Go applications

Setting up a Redis container example:

```go
p := mockredis.Preset()
container, _ := gnomock.Start(p)

defer func() { _ = gnomock.Stop(container) }()

addr := container.DefaultAddress()
client := redis.NewClient(&redis.Options{Addr: addr})
```

See package [reference](https://pkg.go.dev/github.com/orlangure/gnomock?tab=doc).

For Preset documentation, refer to [Presets](#official-presets) section.

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
