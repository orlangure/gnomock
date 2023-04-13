// Package k3s provides a Gnomock Preset for lightweight kubernetes (k3s). This
// preset by no means should be used in any kind of deployment, and no other
// presets are supposed to be deployed in it. The goal of this preset is to
// allow easier testing of Kubernetes automation tools.
//
// This preset uses the `docker.io/rancher/k3s` image on Docker Hub as described
// by the [K3s documentation](https://docs.k3s.io/advanced#running-k3s-in-docker.)
//
// > ```bash
// > $ docker run \
// > --privileged \
// > --name k3s-server-1 \
// > --hostname k3s-server-1 \
// > -p 6443:6443 \
// > -d rancher/k3s:v1.24.10-k3s1 \
// > server
// > ```
//
// Please make sure to pick a version here:
// https://hub.docker.com/r/rancher/k3s/tags.
//
// Keep in mind that k3s runs in a single docker container, meaning it might be
// limited in memory, CPU and storage. Also remember that this cluster always
// runs on a single node.
//
// To connect to this cluster, use `Config` function that can be used together
// with Kubernetes client for Go, or `ConfigBytes` that can be saved as
// `kubeconfig` file and used by `kubectl`.
package k3s

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// apiPort is the port that the K3s HTTPS Kubernetes API gets served over.
	apiPort = 6443
	// defaultVersion is the default k3s version to run.
	defaultVersion = "v1.19.3-k3s3"
)

const (
	// KubeconfigPort is a port that exposes a single `/kubeconfig.yaml`
	// endpoint. It can be used to retrieve a configured kubeconfig file to use
	// to connect to this container using kubectl.
	kubeconfigPort = 48480

	// KubeConfigPortName is the name of the kubeconfig port that serves the
	// `kubeconfig.yaml`.
	KubeConfigPortName = "kubeconfig"

	// k3sManifestsDir is the directory with the K3s container where manifests
	// get automatically applied from.
	k3sManifestsDir = "/var/lib/rancher/k3s/server/manifests/"
)

// kubeconfigHttpd is a representation of the httpd manifest for k3s to
// automatically apply that will serve the k3s admin kubeconfig at
// `/kubeconfig.yaml`.
var kubeconfigHttpd = map[string]interface{}{
	"apiVersion": "v1",
	"kind":       "Pod",
	"metadata": map[string]interface{}{
		"name":      "kubeconfig-httpd",
		"namespace": "kube-system",
	},
	"spec": map[string]interface{}{
		"hostNetwork": true,
		"containers": []map[string]interface{}{
			{
				"name":  "web",
				"image": "docker.io/library/busybox:latest",
				"command": []string{
					"httpd", "-f", "-v",
					"-p", strconv.Itoa(kubeconfigPort),
				},
				"workingDir": "/var/gnomock/",
				"ports": []map[string]interface{}{
					{
						"name":          "http",
						"containerPort": kubeconfigPort,
						"protocol":      "TCP",
					},
				},
				"volumeMounts": []map[string]interface{}{
					{
						"name":      "kubeconfig-dir",
						"mountPath": "/var/gnomock/",
					},
				},
			},
		},
		"volumes": []map[string]interface{}{
			{
				"name": "kubeconfig-dir",
				"hostPath": map[string]interface{}{
					"path": "/var/gnomock/",
					"type": "Directory",
				},
			},
		},
	},
}

func init() {
	registry.Register("kubernetes", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gmomock k3s preset. This preset includes a
// k3s specific healthcheck function and default k3s image and port. Please
// note that this preset launches a privileged docker container.
//
// By default, this preset sets up k3s v1.19.3.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation of lightweight kubernetes (k3s).
type P struct {
	Version string `json:"version"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/rancher/k3s:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	port := gnomock.TCP(apiPort)

	return gnomock.NamedPorts{
		gnomock.DefaultPort: port,
		KubeConfigPortName:  gnomock.TCP(kubeconfigPort),
	}
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	kubeconfigHttpdJSON, err := json.Marshal(kubeconfigHttpd)
	if err != nil {
		log.Fatalf("could not marshal kubeconfig-httpd pod: %s", err)
	}

	httpdManifestB64 := base64.StdEncoding.EncodeToString(kubeconfigHttpdJSON)
	httpdManifestPath := filepath.Join(k3sManifestsDir, "kubeconfig-httpd.json")
	writeHttpdManifestCmd := fmt.Sprintf(
		`mkdir -p %s && echo "%s" | base64 -d > "%s"`,
		filepath.Dir(httpdManifestPath),
		httpdManifestB64,
		httpdManifestPath,
	)

	k3sServerCmd := fmt.Sprintf(
		`/bin/k3s server --disable=traefik --https-listen-port %d`,
		apiPort,
	)

	opts := []gnomock.Option{
		gnomock.WithDebugMode(),
		gnomock.WithHealthCheck(p.healthcheck),
		gnomock.WithPrivileged(),
		gnomock.WithEnv("K3S_KUBECONFIG_OUTPUT=/var/gnomock/kubeconfig.yaml"),
		gnomock.WithEnv("K3S_KUBECONFIG_MODE=644"),
		gnomock.WithEntrypoint(
			"/bin/sh", "-c",
			fmt.Sprintf(`%s && %s`, writeHttpdManifestCmd, k3sServerCmd),
		),
	}

	return opts
}

func (p *P) healthcheck(ctx context.Context, c *gnomock.Container) (err error) {
	kubeconfig, err := Config(c)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	// this is valid only for health checks, and solves a problem where
	// gnomockd performs these calls from within its own container by accessing
	// the cluster at 172.0.0.1, which is not one of the addresses in the
	// certificate
	kubeconfig.Host = c.DefaultAddress()

	client, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client from kubeconfig: %w", err)
	}

	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list cluster nodes: %w", err)
	}

	if len(nodes.Items) == 0 {
		return fmt.Errorf("no nodes found in cluster")
	}

	sas, err := client.CoreV1().ServiceAccounts(metav1.NamespaceDefault).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list service accounts: %w", err)
	}

	if len(sas.Items) == 0 {
		return fmt.Errorf("no service accounts found in cluster")
	}

	return nil
}

func (p *P) setDefaults() {
	if p.Version == "" {
		p.Version = defaultVersion
	}
}

// ConfigBytes returns file contents of kubeconfig file that should be used to
// connect to the cluster running in the provided container.
func ConfigBytes(c *gnomock.Container) (configBytes []byte, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	url := fmt.Sprintf("http://%s/kubeconfig.yaml", c.Address(KubeConfigPortName))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("kubeconfig unavailable: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kubeconfig unavailable: %w", err)
	}

	defer func() {
		closeErr := res.Body.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid kubeconfig response code '%d'", res.StatusCode)
	}

	configBytes, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read kubeconfig body: %w", err)
	}

	configBytes = bytes.ReplaceAll(configBytes,
		[]byte(fmt.Sprintf("https://127.0.0.1:%d", apiPort)),
		[]byte("https://"+c.DefaultAddress()),
	)

	return configBytes, nil
}

// Config returns `*rest.Config` instance of Kubernetes client-go package. This
// config can be used to create a new client that will work against k3s cluster
// running in the provided container.
func Config(c *gnomock.Container) (*rest.Config, error) {
	configBytes, err := ConfigBytes(c)
	if err != nil {
		return nil, fmt.Errorf("can't get kubeconfig bytes: %w", err)
	}

	kubeconfig, err := clientcmd.RESTConfigFromKubeConfig(configBytes)
	if err != nil {
		return nil, fmt.Errorf("can't create kubeconfig from bytes: %w", err)
	}

	return kubeconfig, nil
}
