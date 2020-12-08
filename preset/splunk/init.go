package splunk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/orlangure/gnomock"
)

var errConflict = fmt.Errorf("409: conflict")

// Event is a type used during Splunk initialization. Pass events to WithValues
// to ingest them into the container before the control over it is passed to
// the caller
type Event struct {
	// Event is the actual log entry. Can be any format
	Event string `json:"event"`

	// Index is the name of index to ingest the log into. If the index does not
	// exist, it will be created
	Index string `json:"index"`

	// Source will be used as "source" value of this event in Splunk
	Source string `json:"source"`

	// SourceType will be used as "sourcetype" value of this event in Splunk
	SourceType string `json:"sourcetype"`

	// Time represents event timestamp in seconds, milliseconds or nanoseconds
	// (and maybe even in microseconds, whatever splunk recognizes)
	Time int64 `json:"time"`
}

func (p *P) initf() gnomock.InitFunc {
	return func(ctx context.Context, c *gnomock.Container) (err error) {
		if p.ValuesFile != "" {
			f, err := os.Open(p.ValuesFile)
			if err != nil {
				return fmt.Errorf("can't open values file '%s': %w", p.ValuesFile, err)
			}

			defer func() {
				closeErr := f.Close()
				if err == nil && closeErr != nil {
					err = closeErr
				}
			}()

			events := make([]Event, 0)
			decoder := json.NewDecoder(f)

			for {
				var e Event

				err = decoder.Decode(&e)
				if errors.Is(err, io.EOF) {
					break
				}

				if err != nil {
					return fmt.Errorf("can't read initial event: %w", err)
				}

				events = append(events, e)
			}

			p.Values = append(events, p.Values...)
		}

		err = Ingest(ctx, c, p.AdminPassword, p.Values...)
		if err != nil {
			return fmt.Errorf("can't ingest events: %w", err)
		}

		return nil
	}
}

// Ingest adds the provided events to splunk container. Use the same password
// you provided in WithPassword. Send as many events as you like, this function
// only returns when all the events were indexed, or when the context is timed
// out
func Ingest(ctx context.Context, c *gnomock.Container, password string, events ...Event) error {
	postForm := requestWithPassword(http.MethodPost, password, false)
	apiAddr := c.Address(APIPort)

	token, err := issueToken(postForm, apiAddr)
	if err != nil {
		return fmt.Errorf("can't issue new HEC token: %w", err)
	}

	ensureIndex := indexRegistry(postForm, apiAddr)
	ingestEvent := eventForwarder(
		requestWithPassword(http.MethodPost, token, true),
		c.Address(CollectorPort),
	)

	initialCount, err := countEvents(postForm, apiAddr)
	if err != nil {
		return fmt.Errorf("can't get initial event count: %w", err)
	}

	for _, e := range events {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
			err := ensureIndex(e.Index)
			if err != nil {
				return err
			}

			err = ingestEvent(e)
			if err != nil {
				return err
			}
		}
	}

	var (
		lastErr       error
		lastCount     int
		expectedCount int = initialCount + len(events)
	)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("event count didn't match: want %d, got %d; last error: %v: %w",
				len(events), lastCount, lastErr, context.Canceled)
		default:
			lastCount, lastErr = countEvents(postForm, apiAddr)
			if lastErr == nil && lastCount == expectedCount {
				return nil
			}

			time.Sleep(time.Millisecond * 250)
		}
	}
}

func issueToken(post postFunc, addr string) (string, error) {
	newTokenURL := fmt.Sprintf("https://%s/services/data/inputs/http?output_mode=json", addr)
	tokenName := fmt.Sprintf("gnomock-%d", time.Now().UnixNano())

	data := url.Values{}
	data.Set("name", tokenName)
	buf := bytes.NewBufferString(data.Encode())

	bs, err := post(newTokenURL, buf)
	if err != nil {
		return "", fmt.Errorf("can't create new HEC token: %w", err)
	}

	r := splunkTokenResponse{}

	err = json.Unmarshal(bs, &r)
	if err != nil {
		return "", fmt.Errorf("can't unmarshal HEC token: %w", err)
	}

	return r.Entry[0].Content.Token, nil
}

func indexRegistry(post postFunc, addr string) func(string) error {
	indexes := map[string]bool{"main": true}
	uri := fmt.Sprintf("https://%s/services/data/indexes?output_mode=json", addr)

	return func(indexName string) error {
		if _, ok := indexes[indexName]; !ok {
			_, err := post(uri, bytes.NewBufferString("name="+indexName))
			if err != nil && !errors.Is(err, errConflict) {
				return fmt.Errorf("can't create index: %w", err)
			}

			indexes[indexName] = true
		}

		return nil
	}
}

func eventForwarder(post postFunc, addr string) func(Event) error {
	return func(e Event) error {
		uri := fmt.Sprintf("https://%s/services/collector?output_mode=json", addr)

		eventBytes, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("can't marshal event to json: %w", err)
		}

		_, err = post(uri, bytes.NewBuffer(eventBytes))
		if err != nil {
			return err
		}

		return nil
	}
}

func countEvents(post postFunc, addr string) (int, error) {
	uri := fmt.Sprintf("https://%s/services/search/jobs/export", addr)
	data := url.Values{}
	data.Add("search", "search index=* | stats count")
	data.Add("output_mode", "json")

	bs, err := post(uri, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return 0, err
	}

	var response splunkSearchcResponse

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

func requestWithPassword(method, password string, isJSON bool) postFunc {
	client := insecureClient()

	return func(uri string, buf *bytes.Buffer) (bs []byte, err error) {
		req, err := http.NewRequest(method, uri, buf)
		if err != nil {
			return nil, fmt.Errorf("can't create request: %w", err)
		}

		req.SetBasicAuth("admin", password)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		if isJSON {
			req.Header.Set("Content-Type", "application/json")
		}

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

		bs, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("can't read response body: %w", err)
		}

		if resp.StatusCode == http.StatusConflict {
			return nil, errConflict
		}

		if resp.StatusCode >= http.StatusBadRequest {
			return nil, errors.New(resp.Status + ": " + string(bs))
		}

		return bs, nil
	}
}

type splunkTokenResponse struct {
	Entry []struct {
		Content struct {
			Token string `json:"token"`
		} `json:"content"`
	} `json:"entry"`
}

type splunkSearchcResponse struct {
	Result map[string]interface{} `json:"result"`
}
