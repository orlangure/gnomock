package gnomockd

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomockd/errors"
	"github.com/orlangure/gnomockd/preset"
)

func startHandler(presets preset.Preseter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		p := presets.Preset(name)

		if p == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		sr := &startRequest{Preset: p}
		decoder := json.NewDecoder(r.Body)

		err := decoder.Decode(&sr)
		if err != nil {
			respondWithError(w, errors.InvalidStartRequestError(err))
			return
		}

		c, err := gnomock.Start(p, gnomock.WithOptions(&sr.Options))
		if err != nil {
			respondWithError(w, errors.StartFailedError(err, c))
			return
		}

		err = json.NewEncoder(w).Encode(c)
		if err != nil {
			respondWithError(w, errors.PrepareResponseError(err, c))
			return
		}
	}
}

type startRequest struct {
	Options gnomock.Options `json:"options"`
	Preset  gnomock.Preset  `json:"preset"`
}
