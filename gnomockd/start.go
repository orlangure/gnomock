package gnomockd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/orlangure/gnomock/errors"
	"github.com/orlangure/gnomock/gnomock"
	"github.com/orlangure/gnomock/preset"
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

		started := make(chan bool)
		logWriter, allLogs := setupLogWriter(started)

		c, err := gnomock.Start(p, gnomock.WithOptions(&sr.Options), gnomock.WithLogWriter(logWriter))

		close(started)

		if err != nil {
			containerLogs := <-allLogs
			err = fmt.Errorf("%s: %w", strings.Join(containerLogs, ";"), err)
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

func setupLogWriter(done chan bool) (io.Writer, chan []string) {
	logReader, logWriter := io.Pipe()
	receivedLogLines, allLogs := make(chan string), make(chan []string, 1)

	go func() {
		defer close(receivedLogLines)

		scanner := bufio.NewScanner(logReader)
		for scanner.Scan() {
			receivedLogLines <- scanner.Text()
		}
	}()

	go func() {
		defer func() {
			close(allLogs)

			_, _ = logReader.Close(), logWriter.Close()
		}()

		logs := make([]string, 0)

		for {
			select {
			case <-done:
				allLogs <- logs

				return
			case l := <-receivedLogLines:
				logs = append(logs, l)
			}
		}
	}()

	return logWriter, allLogs
}

type startRequest struct {
	Options gnomock.Options `json:"options"`
	Preset  gnomock.Preset  `json:"preset"`
}
