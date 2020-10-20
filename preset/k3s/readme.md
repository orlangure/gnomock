# Gnomock Lightweight Kubernetes (k3s)

Gnomock Lightweight Kubernetes (k3s) is a
[Gnomock](https://github.com/orlangure/gnomock) preset for running tests
against a real lightweight kubernetes cluster (k3s), without mocks.

```go
package k3s_test

import (
	"context"
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

	p := k3s.Preset()
	c, err := gnomock.Start(
		p,
		gnomock.WithContainerName("k3s"),
		gnomock.WithDebugMode(),
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
```

