package azurite

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/orlangure/gnomock"
)

// WithBlobstorageFiles sets up Blobstorage service running in azurite with the contents of
// `path` directory. The first level children of `path` must be directories,
// their names will be used to create containers. Below them, all the files in any
// other directories, these files will be uploaded as-is.
//
// For example, if you put your test files in testdata/my-container/dir/, Gnomock
// will create "my-container" for you, and pull "dir" with all its contents into
// this container.
func WithBlobstorageFiles(path string) Option {
	return func(p *P) {
		p.BlobstorePath = path
	}
}

func (p *P) initBlobstorage(c *gnomock.Container) error {
	if p.BlobstorePath == "" {
		return nil
	}

	ctx := context.Background()

	connString := fmt.Sprintf(ConnectionStringFormat, AccountName, AccountKey, c.Address(BlobServicePort), AccountName)

	azblobClient, err := azblob.NewClientFromConnectionString(connString, nil)
	if err != nil {
		return err
	}

	containerNames, err := p.createContainer(ctx, azblobClient)
	if err != nil {
		return fmt.Errorf("can't create containerNames: %w", err)
	}

	err = p.uploadFiles(ctx, azblobClient, containerNames)
	if err != nil {
		return err
	}

	return nil
}

func (p *P) createContainer(ctx context.Context, azblobClient *azblob.Client) ([]string, error) {
	files, err := os.ReadDir(p.BlobstorePath)
	if err != nil {
		return nil, fmt.Errorf("can't read blobstorage initial files: %w", err)
	}

	containers := []string{}

	// create containers from top-level folders under `path`
	for _, f := range files {
		if !f.IsDir() {
			continue
		}

		container := f.Name()

		err := p.createContainers(ctx, azblobClient, container)
		if err != nil {
			return nil, fmt.Errorf("can't create container '%s': %w", container, err)
		}

		containers = append(containers, container)
	}

	return containers, nil
}

func (p *P) createContainers(ctx context.Context, azblobClient *azblob.Client, containerName string) error {
	if _, err := azblobClient.CreateContainer(ctx, containerName, nil); err != nil {
		return fmt.Errorf("can't create containerName '%s': %w", containerName, err)
	}

	return nil
}

func (p *P) uploadFiles(ctx context.Context, azblobClient *azblob.Client, containerNames []string) error {
	for _, containerName := range containerNames {
		containerName := containerName

		err := filepath.Walk(
			path.Join(p.BlobstorePath, containerName),
			func(fPath string, file os.FileInfo, err error) error {
				if err != nil {
					return fmt.Errorf("error reading input file '%s': %w", fPath, err)
				}

				if file.IsDir() {
					return nil
				}

				err = p.uploadFile(ctx, azblobClient, containerName, fPath)
				if err != nil {
					return err
				}

				return nil
			},
		)
		if err != nil {
			return fmt.Errorf("error uploading input dir: %w", err)
		}
	}

	return nil
}

func (p *P) uploadFile(ctx context.Context, azblobClient *azblob.Client, containerName, file string) (err error) {
	inputFile, err := os.Open(file) //nolint:gosec
	if err != nil {
		return fmt.Errorf("can't open file '%s': %w", file, err)
	}

	defer func() {
		closeErr := inputFile.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	localPath := path.Join(p.BlobstorePath, containerName)
	key := file[len(localPath):]

	_, err = azblobClient.UploadFile(ctx, containerName, key, inputFile, nil)
	if err != nil {
		return fmt.Errorf("can't upload file '%s' to containerName '%s': %w", file, containerName, err)
	}

	return nil
}
