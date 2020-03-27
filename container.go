package gnomock

import (
	"fmt"
)

// Container represents a docker container created for testing. Host and Ports
// fields should be used to configure the connection to this container. ID
// matches the original docker container ID
type Container struct {
	ID    string
	Host  string
	Ports NamedPorts

	onStop func() error
}

// Address is a convenience function that returns host:port that can be used to
// connect to this container. If a container was created with DefaultTCP call,
// use DefaultPort as the name. Otherwise, use the name of one of the ports
// used during setup
func (c *Container) Address(name string) string {
	p := c.Ports.Get(name)
	return fmt.Sprintf("%s:%d", c.Host, p.Port)
}
