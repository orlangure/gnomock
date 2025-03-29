// This program is intended to run as a sidecar of other, temporary containers.
// After startup, it begins to listen to incoming HTTP connections on port
// :8008. Requests to `/sync/:id` endpoint hang until canceled, and then
// trigger container `:id` to be terminated.
//
// If this program doesn't get any input for 10 seconds, it halts. It also
// terminates after an attempt to stop the requested container.
//
// Inspired by testcontainers/moby-ryuk project.
package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/orlangure/gnomock"
)

func main() {
	var (
		once      sync.Once
		connected = make(chan bool)
	)

	go func() {
		log.Println("waiting for input")

		select {
		case <-connected:
		case <-time.After(time.Second * 10):
			log.Fatalln("no incoming connections, stopping")
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("/sync/", func(_ http.ResponseWriter, r *http.Request) {
		once.Do(func() { connected <- true })

		id := strings.TrimPrefix(r.URL.Path, "/sync/")
		log.Println("got request to kill", id)
		<-r.Context().Done()

		if err := gnomock.Stop(&gnomock.Container{
			ID: id,
		}); err != nil {
			log.Fatalf("can't stop container %s: %s\n", id, err.Error())
		}

		log.Println(id, "killed, exiting")
		os.Exit(0)
	})
	log.Fatalln(http.ListenAndServe(":8008", nil))
}
