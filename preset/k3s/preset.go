// Package k3s provides a Gnomock Preset for lightweight kubernetes (k3s).
package k3s

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const defaultPort = 48443
const defaultVersion = "latest"
const defaultHostname = "localhost"

// KubeconfigPort is a port that exposes a single `/kubeconfig` endpoint. It
// can be used to retrieve a configured kubeconfig file to use to connect to
// this container using kubectl.
const KubeconfigPort = "kubeconfig"
const kubeconfigPort = 80

func init() {
	registry.Register("kubernetes", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gmomock k3s preset. This preset includes a
// k3s specific healthcheck function and default k3s image and port. Please
// note that this preset launches a privileged docker container.
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
	Port    int
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/orlangure/k3s:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	port := gnomock.TCP(p.Port)
	port.HostPort = p.Port

	return gnomock.NamedPorts{
		gnomock.DefaultPort: port,
		KubeconfigPort:      gnomock.TCP(kubeconfigPort),
	}
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(p.healthcheck),
		gnomock.WithPrivileged(),
		gnomock.WithEnv(fmt.Sprintf("K3S_API_HOST=%s", defaultHostname)),
		gnomock.WithEnv(fmt.Sprintf("K3S_API_PORT=%d", p.Port)),
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

	if p.Port == 0 {
		p.Port = defaultPort
	}
}

// ConfigBytes returns file contents of kubeconfig file that should be used to
// connect to the cluster running in the provided container.
func ConfigBytes(c *gnomock.Container) (configBytes []byte, err error) {
	res, err := http.Get(fmt.Sprintf("http://%s/kubeconfig", c.Address(KubeconfigPort)))
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

	configBytes, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read kubeconfig body: %w", err)
	}

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
