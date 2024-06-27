package localstack_test

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/localstack"
	"github.com/stretchr/testify/require"
)

func TestWithS3Files(t *testing.T) {
	// testdata/s3 includes 100 files in my-bucket/dir folder
	p := localstack.Preset(
		localstack.WithServices(localstack.S3),
		localstack.WithS3Files("testdata/s3"),
		localstack.WithVersion("0.11.0"),
	)
	c, err := gnomock.Start(p)

	defer func() { require.NoError(t, gnomock.Stop(c)) }()

	require.NoError(t, err)

	s3Endpoint := fmt.Sprintf("http://%s/", c.Address(localstack.APIPort))

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "a",
				SecretAccessKey: "b",
				SessionToken:    "c",
			}, nil
		},
		)))

	require.NoError(t, err)

	svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = &s3Endpoint
	})

	// my-bucket is automatically created, and now includes 100 files
	listInput := &s3.ListObjectsV2Input{
		Bucket:  aws.String("my-bucket"),
		MaxKeys: aws.Int32(40),
	}

	files, err := svc.ListObjectsV2(context.TODO(), listInput)
	require.NoError(t, err)
	require.Len(t, files.Contents, 40)
	require.True(t, *files.IsTruncated)

	for _, f := range files.Contents {
		require.True(t, strings.HasPrefix(*f.Key, "/dir/f-")) //TODO: Confirm prefix is correct after aws-sdk-go-v2 migration added a leading slash.
	}

	listInput = &s3.ListObjectsV2Input{
		Bucket:            aws.String("my-bucket"),
		ContinuationToken: files.NextContinuationToken,
		MaxKeys:           aws.Int32(50),
	}
	files, err = svc.ListObjectsV2(context.TODO(), listInput)
	require.NoError(t, err)
	require.Len(t, files.Contents, 50)
	require.True(t, *files.IsTruncated)

	for _, f := range files.Contents {
		require.True(t, strings.HasPrefix(*f.Key, "/dir/f-")) //TODO: Confirm prefix is correct after aws-sdk-go-v2 migration added a leading slash.
	}

	// list last batch of files, only 10 left
	listInput = &s3.ListObjectsV2Input{
		Bucket:            aws.String("my-bucket"),
		ContinuationToken: files.NextContinuationToken,
		MaxKeys:           aws.Int32(100),
	}
	files, err = svc.ListObjectsV2(context.TODO(), listInput)
	require.NoError(t, err)
	require.Len(t, files.Contents, 10)
	require.False(t, *files.IsTruncated)

	for _, f := range files.Contents {
		require.True(t, strings.HasPrefix(*f.Key, "/dir/f-")) //TODO: Confirm prefix is correct after aws-sdk-go-v2 migration added a leading slash.
	}
}
