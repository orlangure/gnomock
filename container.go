package gnomock

import (
	"fmt"
	"os"
	"strings"
)

// Container represents a docker container created for testing. Host and Ports
// fields should be used to configure the connection to this container. ID
// matches the original docker container ID.
type Container struct {
	// A unique identifier of this container. The format of this ID may change
	// in the future.
	ID string `json:"id,omitempty"`

	// Host name of bound ports
	//
	// Default: localhost
	Host string `json:"host,omitempty"`

	// A collections of ports exposed by this container. Each port has an alias
	// and an actual port number as exposed on the host
	Ports NamedPorts `json:"ports,omitempty"`

	gateway string
	onStop  func() error
}

// Address is a convenience function that returns host:port that can be used to
// connect to this container. If a container was created with DefaultTCP call,
// use DefaultPort as the name. Otherwise, use the name of one of the ports
// used during setup.
func (c *Container) Address(name string) string {
	p := c.Port(name)
	if p == 0 {
		return ""
	}

	return fmt.Sprintf("%s:%d", c.Host, p)
}

// DefaultAddress return Address() with DefaultPort.
func (c *Container) DefaultAddress() string {
	return c.Address(DefaultPort)
}

// Port is a convenience function that returns port number with the provided
// name.
func (c *Container) Port(name string) int {
	return c.Ports.Get(name).Port
}

// DefaultPort returns Port() with DefaultPort.
func (c *Container) DefaultPort() int {
	return c.Port(DefaultPort)
}

// DockerID returns the ID of this container as known to Docker.
func (c *Container) DockerID() string {
	id, _ := parseID(c.ID)
	return id
}

func isInDocker() bool {
	env := os.Getenv("GNOMOCK_ENV")
	return env == "gnomockd"
}

func customDockerHostAddr() string {
	return os.Getenv("GNOMOCK_HOST_ADDR")
}

func generateID(id, sidecar string) string {
	if len(id) > 10 {
		id = id[:10]
	}

	if len(sidecar) > 10 {
		sidecar = sidecar[:10]
	}

	return fmt.Sprintf("%s-%s", id, sidecar)
}

func parseID(input string) (id, sidecar string) {
	parts := strings.Split(input, "-")
	if len(parts) != 2 {
		return input, ""
	}

	return parts[0], parts[1]
}
