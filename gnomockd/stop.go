package gnomockd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/errors"
)

func stopHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sr stopRequest

		err := json.NewDecoder(r.Body).Decode(&sr)
		if err != nil {
			respondWithError(w, errors.InvalidStopRequestError(err))
			return
		}

		if sr.ID == "" {
			respondWithError(w, errors.InvalidStopRequestError(fmt.Errorf("missing container id")))
			return
		}

		c := &gnomock.Container{ID: sr.ID}

		err = gnomock.Stop(c)
		if err != nil {
			respondWithError(w, errors.StopFailedError(err, c))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type stopRequest struct {
	ID string `json:"id"`
}
