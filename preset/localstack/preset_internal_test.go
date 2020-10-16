package localstack

import (
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/stretchr/testify/require"
)

func TestHealthCheckAddress(t *testing.T) {
	legacyPath := "http://127.0.0.1:33333/health"
	newPath := "http://127.0.0.1:44444/health"
	tests := []struct {
		version  string
		expected string
	}{
		{
			version:  "0.10.0",
			expected: legacyPath,
		},
		{
			version:  "unexpected",
			expected: newPath,
		},
		{
			version:  "latest",
			expected: newPath,
		},
		{
			version:  "0.12.9",
			expected: newPath,
		},
		{
			version:  "1.10.0",
			expected: newPath,
		},
		{
			version:  "0.12.0",
			expected: newPath,
		},
		{
			version:  "0.10.4",
			expected: legacyPath,
		},
		{
			version:  "0.11.3",
			expected: newPath,
		},
		{
			version:  "0.11.4",
			expected: newPath,
		},
		{
			version:  "foo.11.4",
			expected: newPath,
		},
		{
			version:  "0.foo.4",
			expected: newPath,
		},
		{
			version:  "0.11.foo",
			expected: newPath,
		},
	}

	c := &gnomock.Container{
		Host: "127.0.0.1",
		Ports: gnomock.NamedPorts{
			"web": gnomock.Port{
				Protocol: "tcp",
				Port:     33333,
			},
			"api": gnomock.Port{
				Protocol: "tcp",
				Port:     44444,
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
