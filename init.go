package splunk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

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

func initf(password string, events []Event, timeout time.Duration) gnomock.InitFunc {
	return func(c *gnomock.Container) (err error) {
		addr := c.Address(APIPort)
		post := postWithPassword(password)
		ensureIndex := indexRegistry(post, addr)
		ingestEvent := eventForwarder(post, addr)

		for _, e := range events {
			err := ensureIndex(e.Index)
			if err != nil {
				return err
			}

			err = ingestEvent(e)
			if err != nil {
				return err
			}
		}

		timeoutChan := time.After(timeout)

		for {
			select {
			case <-timeoutChan:
				return context.Canceled
			default:
				count, err := countEvents(post, addr)
				if err == nil && count == len(events) {
					return nil
				}

				time.Sleep(time.Millisecond * 250)
			}
		}
	}
}

func indexRegistry(post postFunc, addr string) func(string) error {
	indexes := map[string]bool{"main": true}
	uri := fmt.Sprintf("https://%s/services/data/indexes", addr)

	return func(indexName string) error {
		if _, ok := indexes[indexName]; !ok {
			_, err := post(uri, bytes.NewBufferString("name="+indexName))
			if err != nil {
				return fmt.Errorf("can't create index: %w", err)
			}

			indexes[indexName] = true
		}

		return nil
	}
}

func eventForwarder(post postFunc, addr string) func(Event) error {
	return func(e Event) error {
		uri := fmt.Sprintf(
			"https://%s/services/receivers/simple?index=%s&source=%s&sourcetype=%s",
			addr, e.Index, e.Source, e.SourceType,
		)

		_, err := post(uri, bytes.NewBufferString(e.Event))
		if err != nil {
			return err
		}

		return nil
	}
}

func countEvents(post postFunc, addr string) (int, error) {
	uri := fmt.Sprintf("https://%s/services/search/jobs/export?output_mode=json", addr)
	data := url.Values{}
	data.Add("search", "search index=* | stats count")

	bs, err := post(uri, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return 0, err
	}

	var response splunkResponse

	err = json.Unmarshal(bs, &response)
	if err != nil {
		return 0, err
	}

	countStr, ok := response.Result["count"]
	if !ok {
		return 0, err
	}

	count, err := strconv.Atoi(fmt.Sprintf("%s", countStr))
	if err != nil {
		return 0, err
	}

	return count, nil
}

type postFunc func(string, *bytes.Buffer) ([]byte, error)

func postWithPassword(password string) postFunc {
	client := insecureClient()

	return func(uri string, buf *bytes.Buffer) (bs []byte, err error) {
		req, err := http.NewRequest(http.MethodPost, uri, buf)
		if err != nil {
			return nil, fmt.Errorf("can't create request: %w", err)
		}

		req.SetBasicAuth("admin", password)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}

		defer func() {
			closeErr := resp.Body.Close()
			if err == nil && closeErr != nil {
				err = fmt.Errorf("can't close response body: %w", closeErr)
			}
		}()

		if resp.StatusCode >= http.StatusBadRequest {
			return nil, errors.New(resp.Status)
		}

		bs, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("can't read response body: %w", err)
		}

		return bs, nil
	}
}

type splunkResponse struct {
	Result map[string]interface{} `json:"result"`
}
