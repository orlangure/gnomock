package azurite

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/orlangure/gnomock"
)

func TestHealthCheckAddress(t *testing.T) {
	expectedAddress := "127.0.0.1:10000"
	tests := []struct {
		version  string
		expected string
	}{
		{
			version:  DefaultVersion,
			expected: expectedAddress,
		},
	}

	c := &gnomock.Container{
		Host: "127.0.0.1",
		Ports: gnomock.NamedPorts{
			BlobServicePort: gnomock.Port{
				Protocol: "tcp",
				Port:     10000,
			},
			QueueServicePort: gnomock.Port{
				Protocol: "tcp",
				Port:     10001,
			},
			TableServicePort: gnomock.Port{
				Protocol: "tcp",
				Port:     10002,
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.version, func(t *testing.T) {
			p := P{Version: test.version}
			actual := p.healthCheckAddress(c)
			require.Equal(t, test.expected, actual)
		})
	}
}
