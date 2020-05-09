package localstack_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	localstack "github.com/orlangure/gnomock-localstack"
	"github.com/orlangure/gnomock/gnomock"
	"github.com/stretchr/testify/require"
)

//nolint:funlen
func TestWithS3Files(t *testing.T) {
	// testdata/s3 includes 2500 files in my-bucket/dir folder
	p := localstack.Preset(
		localstack.WithServices(localstack.S3),
		localstack.WithS3Files("testdata/s3"),
	)
	c, err := gnomock.Start(p)

	defer func() { require.NoError(t, gnomock.Stop(c)) }()

	require.NoError(t, err)

	s3Endpoint := fmt.Sprintf("http://%s/", c.Address(localstack.APIPort))
	config := &aws.Config{
		Region:           aws.String("us-east-1"),
		Endpoint:         aws.String(s3Endpoint),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials("a", "b", "c"),
	}

	sess, err := session.NewSession(config)
	require.NoError(t, err)

	svc := s3.New(sess)

	// my-bucket is automatically created, and now includes 2500 files
	// s3 pagination allows to pull 1000 files at a time
	listInput := &s3.ListObjectsV2Input{Bucket: aws.String("my-bucket")}
	files, err := svc.ListObjectsV2(listInput)
	require.NoError(t, err)
	require.Len(t, files.Contents, 1000)
	require.True(t, *files.IsTruncated)

	for _, f := range files.Contents {
		require.True(t, strings.HasPrefix(*f.Key, "dir/f-"))
	}

	// list next 1000 files
	listInput = &s3.ListObjectsV2Input{
		Bucket:            aws.String("my-bucket"),
		ContinuationToken: files.NextContinuationToken,
	}
	files, err = svc.ListObjectsV2(listInput)
	require.NoError(t, err)
	require.Len(t, files.Contents, 1000)
	require.True(t, *files.IsTruncated)

	for _, f := range files.Contents {
		require.True(t, strings.HasPrefix(*f.Key, "dir/f-"))
	}

	// list last batch of files, only 500 left
	listInput = &s3.ListObjectsV2Input{
		Bucket:            aws.String("my-bucket"),
		ContinuationToken: files.NextContinuationToken,
	}
	files, err = svc.ListObjectsV2(listInput)
	require.NoError(t, err)
	require.Len(t, files.Contents, 500)
	require.False(t, *files.IsTruncated)

	for _, f := range files.Contents {
		require.True(t, strings.HasPrefix(*f.Key, "dir/f-"))
	}
}
