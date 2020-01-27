package gnomock

import "fmt"

var ErrImageNotSet = fmt.Errorf("image not set")
var ErrPortNotSet = fmt.Errorf("port not set")
var ErrEnvClient = fmt.Errorf("can't connect to docker host")
