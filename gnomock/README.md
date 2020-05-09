# Preset usage

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

# Usage without presets

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
