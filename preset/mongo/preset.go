// Package mongo includes mongo implementation of Gnomock Preset interface.
// This Preset can be passed to gnomock.StartPreset function to create a
// configured mongo container to use in tests
package mongo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/registry"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"
)

const defaultVersion = "4.4"

func init() {
	registry.Register("mongo", func() gnomock.Preset { return &P{} })
}

// Preset creates a new Gmomock MongoDB preset. This preset includes a MongoDB
// specific healthcheck function, default MongoDB image and port, and allows to
// optionally set up initial state.
//
// By default, this preset uses MongoDB 4.4.
func Preset(opts ...Option) gnomock.Preset {
	p := &P{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// P is a Gnomock Preset implementation of MongoDB.
type P struct {
	DataPath string `json:"data_path"`
	User     string `json:"user"`
	Password string `json:"password"`
	Version  string `json:"version"`
}

// Image returns an image that should be pulled to create this container.
func (p *P) Image() string {
	return fmt.Sprintf("docker.io/library/mongo:%s", p.Version)
}

// Ports returns ports that should be used to access this container.
func (p *P) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(27017)
}

// Options returns a list of options to configure this container.
func (p *P) Options() []gnomock.Option {
	p.setDefaults()

	opts := []gnomock.Option{
		gnomock.WithHealthCheck(healthcheck),
	}

	if p.DataPath != "" {
		opts = append(opts, gnomock.WithInit(p.initf))
	}

	if p.User != "" && p.Password != "" {
		opts = append(
			opts,
			gnomock.WithEnv("MONGO_INITDB_ROOT_USERNAME="+p.User),
			gnomock.WithEnv("MONGO_INITDB_ROOT_PASSWORD="+p.Password),
		)
	}

	return opts
}

func (p *P) setDefaults() {
	if p.Version == "" {
		p.Version = defaultVersion
	}
}

func (p *P) initf(ctx context.Context, c *gnomock.Container) error {
	addr := c.Address(gnomock.DefaultPort)
	uri := "mongodb://" + addr

	if p.useCustomUser() {
		uri = fmt.Sprintf("mongodb://%s:%s@%s", p.User, p.Password, addr)
	}

	clientOptions := mongooptions.Client().ApplyURI(uri)

	client, err := mongodb.NewClient(clientOptions)
	if err != nil {
		return fmt.Errorf("can't create mongo client: %w", err)
	}

	err = client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("can't connect: %w", err)
	}

	topLevelDirs, err := os.ReadDir(p.DataPath)
	if err != nil {
		return fmt.Errorf("can't read test data path: %w", err)
	}

	for _, topLevelDir := range topLevelDirs {
		if !topLevelDir.IsDir() {
			continue
		}

		err = p.setupDB(client, topLevelDir.Name())
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *P) useCustomUser() bool {
	return p.User != "" && p.Password != ""
}

func (p *P) setupDB(client *mongodb.Client, dirName string) error {
	dataFiles, err := os.ReadDir(path.Join(p.DataPath, dirName))
	if err != nil {
		return fmt.Errorf("can't read test data sub path '%s', %w", dirName, err)
	}

	for _, dataFile := range dataFiles {
		if dataFile.IsDir() {
			continue
		}

		fName := dataFile.Name()

		err = p.setupCollection(client, dirName, fName)
		if err != nil {
			return fmt.Errorf("can't setup collection from file '%s': %w", fName, err)
		}
	}

	return nil
}

func (p *P) setupCollection(client *mongodb.Client, dirName, dataFileName string) error {
	collectionName := strings.TrimSuffix(dataFileName, path.Ext(dataFileName))

	file, err := os.Open(path.Join(p.DataPath, dirName, dataFileName)) //nolint:gosec
	if err != nil {
		return fmt.Errorf("can't open file '%s': %w", dataFileName, err)
	}

	vr, err := bsonrw.NewExtJSONValueReader(file, false)
	if err != nil {
		return fmt.Errorf("can't read file '%s': %w", dataFileName, err)
	}

	dec, err := bson.NewDecoder(vr)
	if err != nil {
		return fmt.Errorf("can't create BSON decoder for '%s': %w", dataFileName, err)
	}

	ctx := context.Background()

	for {
		var val interface{}

		err = dec.Decode(&val)
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("can't decode file '%s': %w", dataFileName, err)
		}

		_, err = client.Database(dirName).Collection(collectionName).InsertOne(ctx, val)
		if err != nil {
			return fmt.Errorf("can't insert value from '%s' (%v): %w", dataFileName, val, err)
		}
	}
}

func healthcheck(ctx context.Context, c *gnomock.Container) error {
	addr := c.Address(gnomock.DefaultPort)
	clientOptions := mongooptions.Client().ApplyURI("mongodb://" + addr)

	client, err := mongodb.NewClient(clientOptions)
	if err != nil {
		return fmt.Errorf("can't create mongo client: %w", err)
	}

	err = client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("can't connect: %w", err)
	}

	return client.Ping(ctx, nil)
}
