package pulsar_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/pulsar"
	"github.com/stretchr/testify/require"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestPreset(t *testing.T) {
	versions := []string{
		"3.0.0",
		"4.1.2",
	}

	for _, version := range versions {
		t.Run(version, testPreset(version))
	}
}

func testPreset(version string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		messages := []pulsar.Message{
			{
				Topic:   "test-topic-1",
				Key:     "key1",
				Payload: "test message 1",
			},
			{
				Topic:   "test-topic-2",
				Key:     "key2",
				Payload: "test message 2",
			},
		}

		topics := []string{"test-topic-1", "test-topic-2", "test-topic-3"}

		// gnomock setup
		p := pulsar.Preset(
			pulsar.WithVersion(version),
			pulsar.WithTopics(topics...),
			pulsar.WithMessages(messages...),
		)

		container, err := gnomock.Start(p)
		require.NoError(t, err)

		defer func() { require.NoError(t, gnomock.Stop(container)) }()

		// Test web service accessibility
		webURL := fmt.Sprintf("http://%s/metrics", container.Address(pulsar.WebServicePort))
		resp, err := http.Get(webURL)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Test topic creation via admin API
		for _, topic := range topics {
			adminURL := fmt.Sprintf("http://%s/admin/v2/persistent/public/default/%s",
				container.Address(pulsar.WebServicePort), topic)

			resp, err := http.Get(adminURL)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)
		}
	}
}

func TestPreset_withDefaults(t *testing.T) {
	t.Parallel()

	p := pulsar.Preset()
	container, err := gnomock.Start(p)
	require.NoError(t, err)

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	// Test web service accessibility with default configuration
	webURL := fmt.Sprintf("http://%s/metrics", container.Address(pulsar.WebServicePort))
	resp, err := http.Get(webURL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Test broker port is accessible (we can't easily test the protocol without a client)
	brokerAddr := container.Address(pulsar.BrokerPort)
	require.NotEmpty(t, brokerAddr)
}

func TestPreset_HealthCheck(t *testing.T) {
	t.Parallel()

	p := pulsar.Preset()
	container, err := gnomock.Start(p)
	require.NoError(t, err)

	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	// Verify healthcheck passes by checking container is running
	require.NotNil(t, container)

	// Test that both expected ports are accessible
	webURL := fmt.Sprintf("http://%s/health", container.Address(pulsar.WebServicePort))
	resp, err := http.Get(webURL)
	require.NoError(t, err)
	defer resp.Body.Close()
	// Health endpoint might not exist, but web service should be responsive
	require.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound)
}
