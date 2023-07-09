package gnomockd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/orlangure/gnomock/preset/azurite"
	"github.com/stretchr/testify/assert"

	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/internal/gnomockd"
	"github.com/stretchr/testify/require"
)

func TestAzurite(t *testing.T) {
	t.Parallel()

	h := gnomockd.Handler()
	bs, err := os.ReadFile("./testdata/azurite.json")
	require.NoError(t, err)

	buf := bytes.NewBuffer(bs)
	w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/start/azurite", buf)
	h.ServeHTTP(w, r)

	res := w.Result()

	defer func() { require.NoError(t, res.Body.Close()) }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equalf(t, http.StatusOK, res.StatusCode, string(body))

	var c *gnomock.Container

	err = json.Unmarshal(body, &c)
	require.NoError(t, err)

	connString := fmt.Sprintf(
		azurite.ConnectionStringFormat,
		azurite.AccountName,
		azurite.AccountKey,
		c.Address(azurite.BlobServicePort),
		azurite.AccountName)

	ctx := context.Background()
	azblobClient, connErr := azblob.NewClientFromConnectionString(connString, nil)
	require.NoError(t, connErr)

	var maxResults int32 = 200
	options := azblob.ListBlobsFlatOptions{
		Include:    container.ListBlobsInclude{},
		Marker:     nil,
		MaxResults: &maxResults,
		Prefix:     nil,
	}
	pager := azblobClient.NewListBlobsFlatPager("some-bucket", &options)
	assert.Equal(t, pager.More(), true)

	pagesScanned := 0

	for pager.More() {
		pagesScanned++
		resp, err := pager.NextPage(ctx)

		assert.NoError(t, err)
		assert.Equal(t, 100, len(resp.Segment.BlobItems))

		for _, v := range resp.Segment.BlobItems {
			assert.True(t, strings.HasPrefix(*v.Name, "/file-"))
		}
	}

	assert.Equal(t, 1, pagesScanned)

	bs, err = json.Marshal(c)
	require.NoError(t, err)

	buf = bytes.NewBuffer(bs)
	w, r = httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/stop", buf)
	h.ServeHTTP(w, r)

	res = w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
}
