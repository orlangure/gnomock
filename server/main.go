package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/orlangure/gnomock/gnomockd"
)

var version string

func main() {
	var (
		v    bool
		port int
	)

	flag.BoolVar(&v, "v", false, "display current version")
	flag.IntVar(&port, "port", 23042, "gnomockd port number")
	flag.Parse()

	if v {
		fmt.Println(version)
		os.Exit(0)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Println(http.ListenAndServe(addr, gnomockd.Handler()))
}
