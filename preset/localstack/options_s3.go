package localstack

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/orlangure/gnomock"
)

// WithS3Files sets up S3 service running in localstack with the contents of
// `path` directory. The first level children of `path` must be directories,
// their names will be used to create buckets. Below them, all the files in any
// other directories, these files will be uploaded as-is.
//
// For example, if you put your test files in testdata/my-bucket/dir/, Gnomock
// will create "my-bucket" for you, and pull "dir" with all its contents into
// this bucket.
//
// This function does nothing if you don't provide localstack.S3 as one of the
// services in WithServices.
func WithS3Files(path string) Option {
	return func(p *P) {
		p.S3Path = path
	}
}

func (p *P) initS3(c *gnomock.Container) error {
	if p.S3Path == "" {
		return nil
	}

	s3Endpoint := fmt.Sprintf("http://%s/", c.Address(APIPort))

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(_ context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "a",
				SecretAccessKey: "b",
				SessionToken:    "c",
			}, nil
		},
		)))
	if err != nil {
		return fmt.Errorf("can't create s3 config: %w", err)
	}

	svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = &s3Endpoint
	})

	buckets, err := p.createBuckets(svc)
	if err != nil {
		return fmt.Errorf("can't create buckets: %w", err)
	}

	err = p.uploadFiles(svc, buckets)
	if err != nil {
		return err
	}

	return nil
}

func (p *P) createBuckets(svc *s3.Client) ([]string, error) {
	files, err := os.ReadDir(p.S3Path)
	if err != nil {
		return nil, fmt.Errorf("can't read s3 initial files: %w", err)
	}

	buckets := []string{}

	// create buckets from top-level folders under `path`
	for _, f := range files {
		if !f.IsDir() {
			continue
		}

		bucket := f.Name()

		err := p.createBucket(svc, bucket)
		if err != nil {
			return nil, fmt.Errorf("can't create bucket '%s': %w", bucket, err)
		}

		buckets = append(buckets, bucket)
	}

	return buckets, nil
}

func (p *P) createBucket(svc *s3.Client, bucket string) error {
	input := &s3.CreateBucketInput{Bucket: aws.String(bucket)}

	if _, err := svc.CreateBucket(context.TODO(), input); err != nil {
		return fmt.Errorf("can't create bucket '%s': %w", bucket, err)
	}

	return nil
}

func (p *P) uploadFiles(svc *s3.Client, buckets []string) error {
	for _, bucket := range buckets {
		err := filepath.Walk(
			path.Join(p.S3Path, bucket),
			func(fPath string, file os.FileInfo, err error) error {
				if err != nil {
					return fmt.Errorf("error reading input file '%s': %w", fPath, err)
				}

				if file.IsDir() {
					return nil
				}

				err = p.uploadFile(svc, bucket, fPath)
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

func (p *P) uploadFile(svc *s3.Client, bucket, file string) (err error) {
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

	localPath := path.Join(p.S3Path, bucket)
	key := file[len(localPath):]

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   inputFile,
	}

	_, err = svc.PutObject(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("can't upload file '%s' to bucket '%s': %w", file, bucket, err)
	}

	return nil
}
