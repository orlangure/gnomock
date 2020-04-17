// Package mongo includes mongo implementation of Gnomock Preset interface.
// This Preset can be passed to gnomock.StartPreset function to create a
// configured mongo container to use in tests
package mongo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/orlangure/gnomock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"
)

// Preset creates a new Gmomock MongoDB preset. This preset includes a MongoDB
// specific healthcheck function, default MongoDB image and port, and allows to
// optionally set up initial state
func Preset(opts ...Option) gnomock.Preset {
	config := buildConfig(opts...)

	r := &mongo{
		dataPath: config.dataPath,
		user:     config.user,
		password: config.password,
	}

	return r
}

type mongo struct {
	dataPath string
	user     string
	password string
}

// Image returns an image that should be pulled to create this container
func (m *mongo) Image() string {
	return "docker.io/library/mongo"
}

// Ports returns ports that should be used to access this container
func (m *mongo) Ports() gnomock.NamedPorts {
	return gnomock.DefaultTCP(27017)
}

// Options returns a list of options to configure this container
func (m *mongo) Options() []gnomock.Option {
	opts := []gnomock.Option{
		gnomock.WithHealthCheck(healthcheck),
	}

	if m.dataPath != "" {
		opts = append(opts, gnomock.WithInit(m.initf))
	}

	if m.user != "" && m.password != "" {
		opts = append(
			opts,
			gnomock.WithEnv("MONGO_INITDB_ROOT_USERNAME="+m.user),
			gnomock.WithEnv("MONGO_INITDB_ROOT_PASSWORD="+m.password),
		)
	}

	return opts
}

func (m *mongo) initf(c *gnomock.Container) error {
	addr := c.Address(gnomock.DefaultPort)
	uri := "mongodb://" + addr

	if m.useCustomUser() {
		uri = fmt.Sprintf("mongodb://%s:%s@%s", m.user, m.password, addr)
	}

	clientOptions := mongooptions.Client().ApplyURI(uri)

	client, err := mongodb.NewClient(clientOptions)
	if err != nil {
		return fmt.Errorf("can't create mongo client: %w", err)
	}

	err = client.Connect(context.Background())
	if err != nil {
		return fmt.Errorf("can't connect: %w", err)
	}

	topLevelDirs, err := ioutil.ReadDir(m.dataPath)
	if err != nil {
		return fmt.Errorf("can't read test data path: %w", err)
	}

	for _, topLevelDir := range topLevelDirs {
		if !topLevelDir.IsDir() {
			continue
		}

		err = m.setupDB(client, topLevelDir.Name())
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *mongo) useCustomUser() bool {
	return m.user != "" && m.password != ""
}

func (m *mongo) setupDB(client *mongodb.Client, dirName string) error {
	dataFiles, err := ioutil.ReadDir(path.Join(m.dataPath, dirName))
	if err != nil {
		return fmt.Errorf("can't read test data sub path '%s', %w", dirName, err)
	}

	for _, dataFile := range dataFiles {
		if dataFile.IsDir() {
			continue
		}

		fName := dataFile.Name()

		err = m.setupCollection(client, dirName, fName)
		if err != nil {
			return fmt.Errorf("can't setup collection from file '%s': %w", fName, err)
		}
	}

	return nil
}

func (m *mongo) setupCollection(client *mongodb.Client, dirName, dataFileName string) error {
	collectionName := strings.TrimSuffix(dataFileName, path.Ext(dataFileName))

	file, err := os.Open(path.Join(m.dataPath, dirName, dataFileName)) //nolint:gosec
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

func healthcheck(c *gnomock.Container) error {
	addr := c.Address(gnomock.DefaultPort)
	clientOptions := mongooptions.Client().ApplyURI("mongodb://" + addr)

	client, err := mongodb.NewClient(clientOptions)
	if err != nil {
		return fmt.Errorf("can't create mongo client: %w", err)
	}

	err = client.Connect(context.Background())
	if err != nil {
		return fmt.Errorf("can't connect: %w", err)
	}

	return client.Ping(context.Background(), nil)
}
