package mysql

import (
	"fmt"
)

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/mysql/mysql-server:%s", p.Version)
}
