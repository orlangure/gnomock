// Package gnomockd is an HTTP wrapper around Gnomock
package gnomockd

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/orlangure/gnomock/errors"
	"github.com/orlangure/gnomock/preset"
)

// Handler returns an HTTP handler ready to serve incoming connections
func Handler() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/start/{name}", startHandler(preset.Registry())).Methods(http.MethodPost)
	router.HandleFunc("/stop", stopHandler()).Methods(http.MethodPost).Methods(http.MethodPost)

	return router
}
func respondWithError(w http.ResponseWriter, err error) {
	w.WriteHeader(errors.ErrorCode(err))

	err = json.NewEncoder(w).Encode(err)
	if err != nil {
		log.Println("can't respond with error:", err)
	}
}
