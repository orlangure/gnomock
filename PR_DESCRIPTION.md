# feat: add Jaeger preset to Gnomock

## Description

This PR adds a new preset for [Jaeger](https://www.jaegertracing.io/), a distributed tracing platform. It allows users to easily spin up a Jaeger `all-in-one` container for testing purposes using Gnomock.

Addresses issue #540.

## Implementation Details

*   **Package**: `preset/jaeger`
*   **Image**: Uses `jaegertracing/all-in-one`. The default version is set to `1.57`, but can be customized using `WithVersion()`.
*   **Ports Exposed**:
    *   `16686` (TCP): Web UI (`web` named port)
    *   `14268` (TCP): Collector HTTP (`collector` named port)
    *   `14269` (TCP): Health Check (`health` named port)
*   **Health Check**: A built-in health check verifies that the Jaeger Health Check endpoint returns HTTP 200 OK.

## Usage Example

```go
import (
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/jaeger"
)

func TestSomething(t *testing.T) {
	p := jaeger.Preset(
		jaeger.WithVersion("1.57"),
	)
	container, err := gnomock.Start(p)
	if err != nil {
		t.Fatal(err)
	}
	defer gnomock.Stop(container)

	addr := container.Address("web")
	// Use addr to access Jaeger UI or API...
}
```

## Checklist

-   [x] Implemented `Preset` interface.
-   [x] Added `WithVersion` option.
-   [x] Implemented health check using Jaeger's native health endpoint.
-   [x] Added unit tests (`preset_test.go`).
-   [x] Added documentation (`README.md`).

## Verification

I verified this locally using the included test case `TestPreset` in `preset_test.go`, which starts the container and asserts that the Web UI is reachable.
