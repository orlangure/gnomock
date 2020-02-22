package gnomock

import (
	"fmt"
)

// Container represents a docker container created for testing. Host and Port
// fields should be used to configure the connection to this container. ID
// matches the original docker container ID
type Container struct {
	ID   string
	Host string
	Port string
}

// Address is a convenience function that returns host:port that can be used to
// connect to this container
func (c *Container) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
