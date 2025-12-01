# Gnomock Pulsar

Gnomock Pulsar is a [Gnomock](https://github.com/orlangure/gnomock) preset for
running tests against a real Apache Pulsar message broker, without mocks.

```go
package pulsar_test

import (
    "context"
	"fmt"
	"net/http"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/pulsar"
	"github.com/stretchr/testify/require"
)

func TestPreset(t *testing.T) {
	t.Parallel()

	// gnomock setup
	p := pulsar.Preset(
		pulsar.WithVersion("4.1.2"),
		pulsar.WithTopics("test-topic-1", "test-topic-2"),
		pulsar.WithMessages(
			pulsar.Message{
				Topic:   "test-topic-1",
				Key:     "key1",
				Payload: "test message 1",
			},
			pulsar.Message{
				Topic:   "test-topic-2",
				Key:     "key2",
				Payload: "test message 2",
			},
		),
	)

	container, err := gnomock.Start(p)
	require.NoError(t, err)
	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	// actual test code
	brokerURL := fmt.Sprintf("pulsar://%s", container.Address(pulsar.BrokerPort))
	
	// Use the broker URL with your Pulsar client to produce/consume messages
	// Example: Check web service is accessible
	webURL := fmt.Sprintf("http://%s/metrics", container.Address(pulsar.WebServicePort))
	resp, err := http.Get(webURL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify topics were created
	for _, topic := range []string{"test-topic-1", "test-topic-2"} {
		adminURL := fmt.Sprintf("http://%s/admin/v2/persistent/public/default/%s", 
			container.Address(pulsar.WebServicePort), topic)
		
		resp, err := http.Get(adminURL)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}
}
```

## Ports

- **6650** (broker): Pulsar broker service port for client connections
- **8080** (web): Pulsar web service port for admin API and monitoring

Use `container.Address(pulsar.BrokerPort)` to get the broker URL and
`container.Address(pulsar.WebServicePort)` to get the web service URL.

## Options

- `WithVersion(version string)`: Set Pulsar version (default: "4.1.2")
- `WithTopics(topics ...string)`: Create topics on startup
- `WithMessages(messages ...Message)`: Send initial messages to topics

## Default Configuration

By default, this preset starts Pulsar in standalone mode with:
- Version 4.1.2
- No authentication required
- RocksDB as metadata store (recommended for standalone)
- Default ports: 6650 (broker), 8080 (web)

## Usage in Tests

This preset is ideal for:
- Integration testing with Pulsar clients
- Testing message production and consumption
- Validating Pulsar admin operations
- Development and testing environments

The container runs Pulsar in standalone mode, which provides a complete
Pulsar instance suitable for testing without requiring a full cluster setup.
