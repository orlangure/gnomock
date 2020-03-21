# Gnomock ![Build](https://github.com/orlangure/gnomock/workflows/Build/badge.svg)

Gnomock is a framework to set up temporary docker containers for integration
and end-to-end testing of other applications. It handles pulling images,
starting containers, waiting for them to become available, setting up their
initial state and cleaning up in the end.

It can be used either directly, or via already existing implementations of
various connectors built by the community ([Presets](#official-presets)).

[Reference](https://pkg.go.dev/github.com/orlangure/gnomock?tab=doc)

This project is WIP

## Preset usage

Presets allow to easily setup a service that is popular enough to have a public
Gnomock implementation. Below is a sample
[Redis](https://github.com/orlangure/gnomock-redis) setup:

```go
p := mockredis.Preset()
container, _ := gnomock.StartPreset(p)

defer func() { _ = gnomock.Stop(container) }()

addr := container.Address(gnomock.DefaultPort)
client := redis.NewClient(&redis.Options{Addr: addr})
```

## Official presets

The power of Gnomock is in the Presets developed by the community. Presets,
both existing and planned, are listed below:

- [Redis](https://github.com/orlangure/gnomock-redis)
- [Splunk](https://github.com/orlangure/gnomock-splunk)
- MySQL
- PostgreSQL
- Microsoft SQL Server
- MongoDB
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
container, err := gnomock.Start(
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
