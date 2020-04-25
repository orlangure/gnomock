package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/orlangure/gnomockd/gnomockd"
)

func main() {
	var port int

	flag.IntVar(&port, "port", 23042, "gnomockd port number")
	flag.Parse()

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Println(http.ListenAndServe(addr, gnomockd.Handler()))
}
