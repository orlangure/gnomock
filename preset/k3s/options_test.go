package k3s_test

import (
	"testing"

	"github.com/orlangure/gnomock/preset/k3s"
	"github.com/stretchr/testify/assert"
)

func TestWithVersion(t *testing.T) {
	tests := []struct {
		inVersion          string
		expectedSetVersion string
	}{
		{"v1.17.11", "v1.17.11-k3s1"},
		{"v1.17.12", "v1.17.12-k3s1"},
		{"v1.17.13", "v1.17.13-k3s1"},
		{"v1.17.2", "v1.17.2-k3s1"},
		{"v1.17.3", "v1.17.3-k3s1"},
		{"v1.17.4", "v1.17.4-k3s1"},
		{"v1.17.5", "v1.17.5-k3s1"},
		{"v1.17.6", "v1.17.6-k3s1"},
		{"v1.17.7", "v1.17.7-k3s1"},
		{"v1.17.9", "v1.17.9-k3s1"},
		{"v1.18.10", "v1.18.10-k3s1"},
		{"v1.18.2", "v1.18.2-k3s1"},
		{"v1.18.3", "v1.18.3-k3s1"},
		{"v1.18.4", "v1.18.4-k3s1"},
		{"v1.18.6", "v1.18.6-k3s1"},
		{"v1.18.8", "v1.18.8-k3s1"},
		{"v1.18.9", "v1.18.9-k3s1"},
		{"v1.19.1", "v1.19.1-k3s1"},
		{"v1.19.10", "v1.19.10-k3s1"},
		{"v1.19.11", "v1.19.11-k3s1"},
		{"v1.19.12", "v1.19.12-k3s1"},
		{"v1.19.2", "v1.19.2-k3s1"},
		{"v1.19.3", "v1.19.3-k3s1"},
		{"v1.19.4", "v1.19.4-k3s1"},
		{"v1.19.5", "v1.19.5-k3s1"},
		{"v1.19.7", "v1.19.7-k3s1"},
		{"v1.19.8", "v1.19.8-k3s1"},
		{"v1.19.9", "v1.19.9-k3s1"},
		{"v1.20.2", "v1.20.2-k3s1"},
		{"v1.20.4", "v1.20.4-k3s1"},
		{"v1.20.5", "v1.20.5-k3s1"},
		{"v1.20.6", "v1.20.6-k3s1"},
		{"v1.20.7", "v1.20.7-k3s1"},
		{"v1.20.8", "v1.20.8-k3s1"},
		{"v1.21.0", "v1.21.0-k3s1"},
		{"v1.21.1", "v1.21.1-k3s1"},
		{"v1.21.2", "v1.21.2-k3s1"},
		{"latest", "latest"},
		{"v1.25.8-k3s1-amd64", "v1.25.8-k3s1-amd64"},
		{"v1.26.3-k3s1-arm64", "v1.26.3-k3s1-arm64"},
	}

	for _, tt := range tests {
		t.Run(tt.inVersion, func(t *testing.T) {
			p := k3s.Preset(k3s.WithVersion(tt.inVersion))

			assert.IsType(t, &k3s.P{}, p)
			if k3sp, ok := p.(*k3s.P); ok {
				assert.Equal(t, tt.expectedSetVersion, k3sp.Version)
			}
		})
	}
}
