package gnomockd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/gnomockd"
	"github.com/orlangure/gnomock/preset/localstack"
	"github.com/stretchr/testify/require"
)

func TestLocalstack(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	bs, err := os.ReadFile("./testdata/localstack.json")
	require.NoError(t, err)

	buf := bytes.NewBuffer(bs)
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/localstack", buf)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equalf(t, http.StatusOK, res.StatusCode, string(body))

	var c *gnomock.Container

	err = json.Unmarshal(body, &c)
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

	listInput := &s3.ListObjectsV2Input{Bucket: aws.String("some-bucket")}
	files, err := svc.ListObjectsV2(context.TODO(), listInput)
	require.NoError(t, err)
	require.Len(t, files.Contents, 100)
	require.False(t, *files.IsTruncated)

	for _, f := range files.Contents {
		require.True(t, strings.HasPrefix(*f.Key, "file-"))
	}

	bs, err = json.Marshal(c)
	require.NoError(t, err)

	buf = bytes.NewBuffer(bs)
	w, r = httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", buf)
	h.ServeHTTP(w, r)

	res = w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestLocalstack_invalidService(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	bs, err := os.ReadFile("./testdata/localstack_invalid_service.json")
	require.NoError(t, err)

	buf := bytes.NewBuffer(bs)
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/localstack", buf)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equalf(t, http.StatusBadRequest, res.StatusCode, string(body))
}

func TestLocalstack_unknownService(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	bs, err := os.ReadFile("./testdata/localstack_unknown_service.json")
	require.NoError(t, err)

	buf := bytes.NewBuffer(bs)
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/localstack", buf)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equalf(t, http.StatusBadRequest, res.StatusCode, string(body))
}
