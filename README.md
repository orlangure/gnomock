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
OpenAPI 3.0 [specification](./swagger). Go programs can use extended Gnomock
package directly, without HTTP layer, while other languages require
communication with a local server.

### Using Gnomock in Go applications

See package [reference](https://pkg.go.dev/github.com/orlangure/gnomock?tab=doc).

For Preset documentation, refer to [Presets](#official-presets) section.

### Using Gnomock in other languages

Gnomock runs as a local server, and any program in any language can communicate
with it using OpenAPI 3.0 [specification](./swagger). Below are some language
specific wrappers for Gnomock API:

- [Python SDK](https://github.com/orlangure/gnomock-python-sdk)
- Javascript SDK
- PHP SDK
- Ruby SDK
- Java SDK
- [Other](https://openapi-generator.tech/docs/generators) languages

## Official presets

The power of Gnomock is in the Presets developed by the community. Presets,
both existing and planned, are listed below:

| Preset | Repository | Reference |
|--------|------------|-----------|
Localstack | https://github.com/orlangure/gnomock-localstack | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock-localstack?tab=doc)
Splunk | https://github.com/orlangure/gnomock-splunk | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock-splunk?tab=doc)
Redis | https://github.com/orlangure/gnomock-redis | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock-redis?tab=doc)
MySQL | https://github.com/orlangure/gnomock-mysql | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock-mysql?tab=doc)
PostgreSQL | https://github.com/orlangure/gnomock-postgres | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock-postgres?tab=doc)
Microsoft SQL Server | https://github.com/orlangure/gnomock-mssql | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock-mssql?tab=doc)
MongoDB | https://github.com/orlangure/gnomock-mongo | [Reference](https://pkg.go.dev/github.com/orlangure/gnomock-mongo?tab=doc)
Elasticsearch | |
DynamoDB | |
Cassandra | |
MariaDB | |

## Preset usage

Presets allow to easily setup a service that is popular enough to have a public
Gnomock implementation. Below is a sample
[Redis](https://github.com/orlangure/gnomock-redis) setup:

```go
p := mockredis.Preset()
container, _ := gnomock.Start(p)

defer func() { _ = gnomock.Stop(container) }()

addr := container.DefaultAddress()
client := redis.NewClient(&redis.Options{Addr: addr})
```

With Gnomock it is easy to setup complex environments using multiple presets.
It could be done in parallel. For example, below is a dependency stack of a Go
program that uses Postgres database, Redis cache, AWS S3 storage and AWS SES:

```go
containers, err := gnomock.InParallel().
    Start(mockredis.Preset()).
    Start(mockpostgres.Preset(), mockpostgres.WithUser("user", "pass")).
    Start(
            localstack.Preset(),
            localstack.WithServices(localstack.S3, localstack.SES),
         ).
    Go()

defer func() { _ = gnomock.Stop(containers...) }()
```

## Usage without presets

Gnomock can be used directly, without Presets. It requires a bit more work
since every use case needs its own healthcheck and initialization
implementation, as well as detailed configuration targeted at that particular
use case:

```go
// assuming the container exposes 2 ports
namedPorts := gnomock.NamedPorts{
    "web80":   gnomock.TCP(80),
    "web8080": gnomock.TCP(8080),
}

// see docs for option description
container, err := gnomock.StartCustom(
    testImage, namedPorts,
    gnomock.WithHealthCheckInterval(time.Microsecond*500),
    gnomock.WithHealthCheck(healthcheck),
    gnomock.WithInit(initf),
    gnomock.WithContext(context.Background()),
    gnomock.WithStartTimeout(time.Second*30),
    gnomock.WithWaitTimeout(time.Second*1),
    gnomock.WithEnv("GNOMOCK_TEST_1=foo"),
    gnomock.WithEnv("GNOMOCK_TEST_2=bar"),
)

// stop and remove the container after tests
defer gnomock.Stop(container)

// ports bound on host are different from the ports containers expose
addr80 := container.Address("web80")
addr8080 := container.Address("web8080")
```

To keep test code clean and simple, it is better to wrap custom use cases with
Preset implementation, that can be contributed back to the community.
