# <div align="center">Gnomock – test your Go code without mocks</div>

🏗️ Spin up entire dependency stack

🎁 Setup initial dependency state – easily!

🏭 Test against actual, close to production software

⏳ Spend no time writing mocks

🕹️ Test actual program behavior and side effects

## <div align="center">![Build](https://github.com/orlangure/gnomock/workflows/Build/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/orlangure/gnomock)](https://goreportcard.com/report/github.com/orlangure/gnomock)</div>

Gnomock is an integration and end-to-end testing framework for Go applications.
It uses Docker to create **temporary containers** for application dependencies,
**setup** their **initial state** and **clean them up** in the end. Gnomock
allows to test the code with **no mocks** wherever possible.

The power of Gnomock is in a variety of [Presets](#official-presets), each
implementing a specific database, service or other tools. Each preset provides
ways of setting up its initial state as easily as possible: SQL schema
creation, test data upload into S3, sending test events to Splunk, etc.

## 

[Reference](https://pkg.go.dev/github.com/orlangure/gnomock?tab=doc) |
[Roadmap](#roadmap)

## 

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
It could be done in parallel:

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

## Official presets

The power of Gnomock is in the Presets developed by the community. Presets,
both existing and planned, are listed below:

- [Localstack](https://github.com/orlangure/gnomock-localstack)
- [Splunk](https://github.com/orlangure/gnomock-splunk)
- [Redis](https://github.com/orlangure/gnomock-redis)
- [MySQL](https://github.com/orlangure/gnomock-mysql)
- [PostgreSQL](https://github.com/orlangure/gnomock-postgres)
- [Microsoft SQL Server](https://github.com/orlangure/gnomock-mssql)
- [MongoDB](https://github.com/orlangure/gnomock-mongo)
- Elasticsearch
- DynamoDB
- Cassandra
- MariaDB

## Usage without presets

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

## Roadmap

Gnomock is being developed at this moment. There are many tasks that need to be
completed before it can be considered publicly available:

- [ ] Implement [presets](#official-presets). This list is compiled from the
  most popular images in Docker Hub. It is a starting point.
- [ ] Improve README of Gnomock and all the Presets. These files show usage
  examples, but can be improved without turning into full blown guides. Users
  should be able to get started using only Gnomock and Preset README.
- [ ] Prepare contribution guides for Gnomock and Preset repositories. Gnomock
  should get its power from community-implemented Presets, so it should be very
  clear how to implement new presets, or contribute to Gnomock itself.
- [ ] Start spreading the word.
