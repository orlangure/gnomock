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
			respondWithError(w, errors.NewPresetNotFoundError(name))
			return
		}

		sr := &startRequest{Preset: p}
		decoder := json.NewDecoder(r.Body)

		err := decoder.Decode(&sr)
		if err != nil {
			respondWithError(w, errors.NewInvalidStartRequestError(err))
			return
		}

		c, err := gnomock.Start(p, gnomock.WithOptions(&sr.Options))
		if err != nil {
			respondWithError(w, errors.NewStartFailedError(err, c))
			return
		}

		err = json.NewEncoder(w).Encode(c)
		if err != nil {
			respondWithError(w, errors.NewStartFailedError(err, c))
			return
		}
	}
}

type startRequest struct {
	Options gnomock.Options `json:"options"`
	Preset  gnomock.Preset  `json:"preset"`
}
