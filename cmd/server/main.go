package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/orlangure/gnomock/internal/gnomockd"
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

	if pStr, ok := os.LookupEnv("GNOMOCKD_PORT"); ok {
		if p, err := strconv.Atoi(pStr); err == nil {
			port = p
		}
	}

	addr := fmt.Sprintf(":%d", port)
	log.Println(http.ListenAndServe(addr, gnomockd.Handler())) // nolint: gosec
}
