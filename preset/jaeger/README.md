# Jaeger

Jaeger preset for Gnomock.

```go
import "github.com/orlangure/gnomock/preset/jaeger"

// ...

p := jaeger.Preset(jaeger.WithVersion("1.57"))
container, err := gnomock.Start(p)
```
