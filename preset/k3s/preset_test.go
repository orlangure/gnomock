package k3s_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/k3s"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestPreset(t *testing.T) {
	t.Parallel()

	p := k3s.Preset(
		k3s.WithPort(48448),
		k3s.WithVersion("v1.26.3-k3s1"),
	)
	c, err := gnomock.Start(
		p,
		gnomock.WithContainerName("k3s"),
	)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, gnomock.Stop(c))
	}()

	kubeconfig, err := k3s.Config(c)
	require.NoError(t, err)

	client, err := kubernetes.NewForConfig(kubeconfig)
	require.NoError(t, err)

	ctx := context.Background()

	pods, err := client.CoreV1().Pods(metav1.NamespaceDefault).List(ctx, metav1.ListOptions{})
	require.NoError(t, err)
	require.Empty(t, pods.Items)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gnomock",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "gnomock",
					Image: "docker.io/orlangure/gnomock-test-image",
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}

	_, err = client.CoreV1().Pods(metav1.NamespaceDefault).Create(ctx, pod, metav1.CreateOptions{})
	require.NoError(t, err)

	pods, err = client.CoreV1().Pods(metav1.NamespaceDefault).List(ctx, metav1.ListOptions{})
	require.NoError(t, err)
	require.Len(t, pods.Items, 1)
	require.Equal(t, "gnomock", pods.Items[0].Name)
}

func TestPreset_withDefaults(t *testing.T) {
	t.Parallel()

	p := k3s.Preset()
	c, err := gnomock.Start(p, gnomock.WithContainerName("k3s-default"))
	require.NoError(t, err)

	defer func() {
		require.NoError(t, gnomock.Stop(c))
	}()

	kubeconfig, err := k3s.Config(c)
	require.NoError(t, err)

	client, err := kubernetes.NewForConfig(kubeconfig)
	require.NoError(t, err)

	ctx := context.Background()

	pods, err := client.CoreV1().Pods(metav1.NamespaceDefault).List(ctx, metav1.ListOptions{})
	require.NoError(t, err)
	require.Empty(t, pods.Items)
}

func TestConfigBytes(t *testing.T) {
	t.Parallel()

	t.Run("fails on wrong url", func(t *testing.T) {
		p := k3s.Preset()
		c := &gnomock.Container{
			Host:  "1%%2",
			Ports: p.Ports(),
		}

		bs, err := k3s.ConfigBytes(c)
		require.Nil(t, bs)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid URL")
	})

	t.Run("fails on wrong port", func(t *testing.T) {
		ports := gnomock.NamedPorts{
			k3s.KubeConfigPortName: gnomock.TCP(1),
		}

		c := &gnomock.Container{
			Host:  "127.0.0.1",
			Ports: ports,
		}

		bs, err := k3s.ConfigBytes(c)
		require.Nil(t, bs)
		require.Error(t, err)
		require.Contains(t, err.Error(), "connection refused")
	})
}

func TestPreset_WithDynamicPort(t *testing.T) {
	t.Parallel()

	p := k3s.Preset(
		k3s.WithDynamicPort(),
	)
	c, err := gnomock.Start(p, gnomock.WithContainerName("k3s-dynamic"))
	require.NoError(t, err)

	defer func() {
		require.NoError(t, gnomock.Stop(c))
	}()

	kubeconfig, err := k3s.Config(c)
	require.NoError(t, err)

	client, err := kubernetes.NewForConfig(kubeconfig)
	require.NoError(t, err)

	ctx := context.Background()

	pods, err := client.CoreV1().Pods(metav1.NamespaceDefault).List(ctx, metav1.ListOptions{})
	require.NoError(t, err)
	require.Empty(t, pods.Items)
}

func TestPreset_Versions(t *testing.T) {
	tests := []struct {
		inVersion string
	}{
		{"v1.24.12-k3s1"},
		{"v1.25.8-k3s1"},
		{"v1.26.3-k3s1"},
		// k3s v1.27.0 not available yet: (https://github.com/k3s-io/k3s/pull/7271)
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.inVersion, func(t *testing.T) {
			t.Parallel()

			p := k3s.Preset(
				k3s.WithDynamicPort(),
				k3s.WithVersion(tt.inVersion),
			)
			c, err := gnomock.Start(p, gnomock.WithContainerName(fmt.Sprintf("k3s-%s", tt.inVersion)))
			require.NoError(t, err)

			defer func() {
				require.NoError(t, gnomock.Stop(c))
			}()

			kubeconfig, err := k3s.Config(c)
			require.NoError(t, err)

			client, err := kubernetes.NewForConfig(kubeconfig)
			require.NoError(t, err)

			ctx := context.Background()

			pods, err := client.CoreV1().Pods(metav1.NamespaceDefault).List(ctx, metav1.ListOptions{})
			require.NoError(t, err)
			require.Empty(t, pods.Items)
		})
	}
}
