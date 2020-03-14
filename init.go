package splunk

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/orlangure/gnomock"
)

// Event is a type used during Splunk initialization. Pass events to WithValues
// to ingest them into the container before the control over it is passed to
// the caller
type Event struct {
	// Event is the actual log entry. Can be any format
	Event string

	// Index is the name of index to ingest the log into. If the index does not
	// exist, it will be created
	Index string

	// Source will be used as "source" value of this event in Splunk
	Source string

	// SourceType will be used as "sourcetype" value of this event in Splunk
	SourceType string
}

func initf(password string, events []Event) gnomock.InitFunc {
	return func(c *gnomock.Container) (err error) {
		addr := c.Address(APIPort)
		ensureIndex := indexRegistry(addr, password)
		ingestEvent := postWithPassword(password)

		for _, e := range events {
			err := ensureIndex(e.Index)
			if err != nil {
				return err
			}

			uri := fmt.Sprintf(
				"https://%s/services/receivers/simple?index=%s&source=%s&sourcetype=%s",
				addr, e.Index, e.Source, e.SourceType,
			)

			err = ingestEvent(uri, bytes.NewBufferString(e.Event))
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func indexRegistry(addr, password string) func(string) error {
	indexes := map[string]bool{"main": true}
	post := postWithPassword(password)
	uri := fmt.Sprintf("https://%s/services/data/indexes", addr)

	return func(indexName string) error {
		if _, ok := indexes[indexName]; !ok {
			err := post(uri, bytes.NewBufferString("name="+indexName))
			if err != nil {
				return fmt.Errorf("can't create index: %w", err)
			}

			indexes[indexName] = true
		}

		return nil
	}
}

func postWithPassword(password string) func(string, *bytes.Buffer) error {
	client := insecureClient()

	return func(uri string, buf *bytes.Buffer) (err error) {
		req, err := http.NewRequest(http.MethodPost, uri, buf)
		if err != nil {
			return fmt.Errorf("can't create request: %w", err)
		}

		req.SetBasicAuth("admin", password)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}

		defer func() {
			closeErr := resp.Body.Close()
			if err == nil && closeErr != nil {
				err = fmt.Errorf("can't close response body: %w", closeErr)
			}
		}()

		if resp.StatusCode >= http.StatusBadRequest {
			return errors.New(resp.Status)
		}

		return nil
	}
}
