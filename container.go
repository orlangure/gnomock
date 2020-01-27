package gnomock

import (
	"fmt"
)

type Container struct {
	ID   string
	Host string
	Port string
}

func (c *Container) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
