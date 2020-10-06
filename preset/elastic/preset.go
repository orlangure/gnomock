// Package elastic provides a Gnomock Preset for Elasticsearch.
package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/orlangure/gnomock"
)

const defaultVersion = "7.9.2"
const defaultPort = 9200

// Preset creates a new Gmomock Elasticsearch preset. This preset includes an
// Elasticsearch specific healthcheck function and default Elasticsearch image
// and port.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation of Elasticsearch.
type P struct {
	Version string   `json:"version"`
	Inputs  []string `json:"input_files"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/library/elasticsearch:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(defaultPort)
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithEnv("discovery.type=single-node"),
		gnomock.WithHealthCheck(p.healthcheck),
	}

	if len(p.Inputs) > 0 {
		opts = append(opts, gnomock.WithInit(p.initf))
	}

	return opts
}

func (p *P) healthcheck(ctx context.Context, c *gnomock.Container) (err error) {
	cfg := elasticsearch.Config{
		Addresses:    []string{fmt.Sprintf("http://%s", c.DefaultAddress())},
		DisableRetry: true,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("can't create elasticsearch client: %w", err)
	}

	res, err := client.Info()
	if err != nil {
		return fmt.Errorf("can't get cluster info: %w", err)
	}

	defer func() {
		closeErr := res.Body.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if res.IsError() {
		return fmt.Errorf("cluster info failed: %s", res.String())
	}

	return nil
}

func (p *P) initf(ctx context.Context, c *gnomock.Container) (err error) {
	cfg := elasticsearch.Config{
		Addresses:    []string{fmt.Sprintf("http://%s", c.DefaultAddress())},
		DisableRetry: true,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("can't create elasticsearch client: %w", err)
	}

	docCount := 0

	for _, file := range p.Inputs {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
			n, err := p.ingestFile(file, client)
			if err != nil {
				return fmt.Errorf("can't ingest file '%s': %w", file, err)
			}

			docCount += n
		}
	}

	tick := time.NewTicker(time.Millisecond * 250)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-tick.C:
			total, err := p.totalDocCount(client)
			if err != nil {
				return fmt.Errorf("can't count docs: %w", err)
			}

			if total == docCount {
				return nil
			}
		}
	}
}

func (p *P) totalDocCount(client *elasticsearch.Client) (n int, err error) {
	res, err := client.Indices.Stats(client.Indices.Stats.WithFilterPath("_all.total.docs.count"))
	if err != nil {
		return 0, fmt.Errorf("failed to get index status: %w", err)
	}

	defer func() {
		closeErr := res.Body.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	if res.IsError() {
		return 0, fmt.Errorf("invalid response for index status: %s", res.String())
	}

	var out struct {
		All struct {
			Total struct {
				Docs struct {
					Count int `json:"count"`
				} `json:"docs"`
			} `json:"total"`
		} `json:"_all"`
	}

	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return 0, fmt.Errorf("invalid output for index stats: %w", err)
	}

	return out.All.Total.Docs.Count, nil
}

func (p *P) ingestFile(fName string, client *elasticsearch.Client) (docCount int, err error) {
	f, err := os.Open(fName) // nolint:gosec
	if err != nil {
		return 0, fmt.Errorf("can't open file '%s': %w", fName, err)
	}

	defer func() {
		closeErr := f.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	for decoder := json.NewDecoder(f); ; {
		var v interface{}

		if err := decoder.Decode(&v); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return docCount, fmt.Errorf("can't decode json data: %w", err)
		}

		bs, err := json.Marshal(v)
		if err != nil {
			return docCount, err
		}

		if err := p.ingestData(path.Base(fName), bs, client); err != nil {
			return 0, fmt.Errorf("failed to ingest data: %w", err)
		}

		docCount++
	}

	return docCount, nil
}

func (p *P) ingestData(index string, bs []byte, client *elasticsearch.Client) (err error) {
	res, err := client.Index(index, bytes.NewBuffer(bs))
	if err != nil {
		return fmt.Errorf("failed to index file '%s': %w", index, err)
	}

	defer func() {
		closeErr := res.Body.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	if res.IsError() {
		return fmt.Errorf("indexing of '%s' failed: %s", index, res.String())
	}

	return nil
}

func (p *P) setDefaults() {
	if p.Version == "" {
		p.Version = defaultVersion
	}
}
