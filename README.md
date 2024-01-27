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

Gnomock is an integration and end-to-end testing toolkit. It uses Docker to create **temporary containers** for application dependencies, **setup** their **initial state** and **clean them up** in the end. Gnomock allows to test the code with **no mocks** wherever possible.

The power of Gnomock is in a variety of [Presets](#official-presets), each implementing a specific database, service or other tools. Each preset provides ways of setting up its initial state as easily as possible: SQL schema creation, test data upload into S3, sending test events to Splunk, etc.

The name "Gnomock" stands for "no mock", with a "G" for "Go" 😼. It also sounds like "gnome", that's why the friendly garden gnome artwork (by [Michael Zolotov](https://www.mzolotov.com/))

## Demo

See for yourself how easy and fast it is to write tests that use actual services running in ephemeral Docker containers:

[![asciicast](https://asciinema.org/a/jSJJGk0n3q1g4Fc4bvZYYyifN.svg)](https://asciinema.org/a/jSJJGk0n3q1g4Fc4bvZYYyifN)

## Table of contents

- [Getting started](#getting-started)
  - [Using Gnomock in Go applications](#using-gnomock-in-go-applications)
  - [Using Gnomock in other languages](#using-gnomock-in-other-languages)
- [Official presets](#official-presets)
- [Similar projects](#similar-projects)
- [Troubleshooting](#troubleshooting)
- [Giving back](#giving-back) 💰

## Getting started

Gnomock can be used in two different ways:

- Imported directly as a package in any **Go** project
- Accessed over HTTP running as a daemon in **any other language**

Both ways **require** an active Docker daemon running **locally** in the same environment.

External `DOCKER_HOST` support is experimental. It cannot be reliably tested at this moment, but it might work.

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

See package [reference](https://pkg.go.dev/github.com/orlangure/gnomock?tab=doc). For Preset documentation, refer to [Presets](#official-presets) section.

### Using Gnomock in other languages

If you use Go, please refer to [Using Gnomock in Go applications](#using-gnomock-in-go-applications) section. Otherwise, refer to [documentation](docs/server.md).

## Official presets

The power of Gnomock is in the Presets. Existing Presets with their supported<sup>\*</sup> versions are listed below.

<small>*\* **Supported** versions are tested as part of CI pipeline. Other versions might work as well.*</small>

| Preset | Go package | Go API | Supported versions                        | arm64 |
|--------|------------|--------|-------------------------------------------|-------|
[Localstack](https://github.com/localstack/localstack) (AWS) | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/localstack) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/localstack?tab=doc) | `0.12.2`, `0.13.1`, `0.14.0`, `2.3.0`, `3.1.0`              | ✅
Splunk | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/splunk) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/splunk?tab=doc) | `8.0.2`                                   | ❌
Redis | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/redis) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/redis?tab=doc) | `5.0.10`, `6.0.9`, `7.2.4`                         | ✅
Memcached | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/memcached) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/memcached?tab=doc) | `1.6.9`, `1.6.23`                                   | ✅
MySQL | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/mysql) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mysql?tab=doc) | `5.7.32`, `8.0.22`                        | ✅
MariaDB | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/mariadb) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mariadb?tab=doc) | `10.5.8`, `11.2.2`                                  | ✅
PostgreSQL | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/postgres) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/postgres?tab=doc) | `10.15`, `11.10`, `12.5`, `13.1`, `14.11`, `15.6`, `16.2`          | ✅
Microsoft SQL Server | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/mssql) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mssql?tab=doc) | `2017-latest`, `2019-latest`              | ❌
MongoDB | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/mongo) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/mongo?tab=doc) | `3.6.21`, `4.4`, `5.0`                           | ✅
RabbitMQ | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/rabbitmq) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/rabbitmq?tab=doc) | `3.8.9-alpine`, `3.8.9-management-alpine`, `3.13-alpine` | ✅
Kafka | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/kafka) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/kafka?tab=doc) | `3.3.1-L0`,  `3.6.1-L0`                                       | ✅
Elasticsearch | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/elastic) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/elastic?tab=doc) | `8.13.0`, `7.17.21`                         | ✅
Kubernetes | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/k3s) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/k3s?tab=doc) | `v1.26.3-k3s1`                            | ✅
CockroachDB | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/cockroachdb) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/cockroachdb?tab=doc) | `v19.2.11`, `v20.1.10`, `v21.2.17`, `v22.2.19`, `v23.1.20`                    | ✅
InfluxDB | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/influxdb) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/influxdb?tab=doc) | `2.7.6-alpine`                            | ✅
Cassandra | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/cassandra) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/cassandra?tab=doc) | `4.0`, `3`                                | ✅
Vault | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/vault) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/vault?tab=doc) | `1.10`, `1.11`, `1.12`, `1.13`     | ✅
Azurite | [Go package](https://github.com/orlangure/gnomock/tree/master/preset/azurite) | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock/preset/azurite?tab=doc) | `3.30.0`                                  | ✅
<!-- new presets go here -->

It is possible to use Gnomock [directly from Go](https://pkg.go.dev/github.com/orlangure/gnomock#StartCustom) code without any presets. HTTP API only allows to setup containers using presets that exist in this repository.

## Similar projects

Gnomock is not the only project that aims to simplify integration and end-to-end testing by using ephemeral docker containers:

- `testcontainers/testcontainers-go`
- `ory/dockertest`

These projects are amazing, and they give plenty of flexibility and power to their users. There are many things that are possible with them, but are impossible with Gnomock. Still, below is a short list of things that sometimes give Gnomock an advantage:

- **Gnomock tries to provide a batteries-included solution**. Gnomock has a growing number of Presets, each one implementing an integration with a popular external service. For every Preset, there already is a number of "invisible" utilities that transparently relieve you from implementing them yourself:
  - __Built-in health check function__ that you don't even need to know it exists. It makes sure you only get control over a container when it is ready to use.
  - __Wrappers for some of the configuration__ exposed by the container, such as default username/password. You can easily provide your own credentials to connect to the container.
  - __Seed data ingestion__ for your convenience. Sometimes you just need to make sure your queries work given some data. Gnomock puts your data in there with a single line of code. Sometimes you only test a program that consumes messages from Kafka, and Gnomock produces the messages for you with another line of code.
- **Simple API** that does not expose anything that happens "under the hood" most of the time. Yet Gnomock allows some additional configuration and custom Preset implementation whenever necessary.
- Gnomock's vision includes **being useful not only in Go** projects, but in any projects via HTTP. It already supports almost all its features over HTTP layer and has a clear OpenAPI spec.
- Gnomock has a friendly **garden gnome mascot**😻

## Troubleshooting

### Tests with Gnomock take too long and time-out eventually

It happens a lot locally if your internet isn't fast enough to pull docker images used in tests. In CI, such as in Github Actions, the images are downloaded very quickly. To work around this issue locally, pull the image manually before running the tests. You only need to do it once, the images stay in local cache until deleted. For example, to pull Postgres 11 image, run:

```
docker pull postgres:11
```

### Tests time-out even when the image exists locally

It can happen if the containers can't become ready to use before they time out. By default, Gnomock uses fairly high timeouts for new containers (for starting and for setting them up). If you choose to change default timeout using `WithTimeout` (`timeout` in HTTP), it is possible that the values you choose are too short.

### Tests pass when run one-by-one, and fail when run in parallel

It happens when you try to start up **a lot** of containers at the same time. The system, especially in CI environments such as Github Actions, cannot handle the load, and containers fail to become healthy before they time-out. That's the reason Gnomock has a few separate build jobs, each running only a small subset of tests, one package at a time.

### Containers fail to setup with a "File not found" error

If you run `gnomock` as a server, you need to make sure the files you use in your setup are available inside `gnomock` container. Use `-v $(pwd):$(pwd)` argument to `docker run` to mount the current working directory under the same path inside the `gnomock` container. If you prefer to keep a permanent `gnomock` container running, you can mount your entire `$HOME` directory (or any other directory where you keep the code).

## Giving back

This is a free and open source project that hopefully helps its users, at least a little. Even though I don't need donations to support it, I understand that there are people that wish to give back anyway. If you are one of them, I encourage you to [plant some trees with Tree Nation](https://tree-nation.com/plant/offer) 🌲 🌳 🌴

If you want me to know about your contribution, make sure to use `orlangure+gnomock@gmail.com` as the recipient email.

Thank you!
