package azurite_test

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/azurite"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestWithBlobstorageFiles(t *testing.T) {
	// testdata/blobstorage includes 10 files in some-container/dir folder
	p := azurite.Preset(
		azurite.WithBlobstorageFiles("./testdata/blobstorage"),
		azurite.WithVersion(azurite.DefaultVersion),
	)
	c, err := gnomock.Start(p)

	defer func() { require.NoError(t, gnomock.Stop(c)) }()

	require.NoError(t, err)

	connString := fmt.Sprintf(azurite.ConnectionStringFormat, azurite.AccountName, azurite.AccountKey, c.Address(azurite.BlobServicePort), azurite.AccountName)

	azblobClient, err := azblob.NewClientFromConnectionString(connString, nil)
	require.NoError(t, err)

	// some-container is automatically created, and now includes 10 files
	containerName := "some-container"

	nextMarker := listAndCheckFiles(t, azblobClient, containerName, 4, 4, nil)
	nextMarker = listAndCheckFiles(t, azblobClient, containerName, 5, 5, nextMarker)
	_ = listAndCheckFiles(t, azblobClient, containerName, 10, 1, nextMarker)

}

func listAndCheckFiles(t *testing.T, azblobClient *azblob.Client, containerName string, maxResults int32, maxResultsExpected int, marker *string) (nextMarker *string) {
	pager := azblobClient.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
		MaxResults: &maxResults,
		Marker:     marker,
	})
	require.Equal(t, pager.More(), true)
	resp, err := pager.NextPage(context.Background())
	require.NoError(t, err)
	require.Equal(t, maxResultsExpected, len(resp.Segment.BlobItems))
	nextMarker = resp.NextMarker
	checkFiles(t, resp.Segment.BlobItems)
	return
}

func checkFiles(t *testing.T, blobItems []*container.BlobItem) {
	for _, f := range blobItems {
		require.True(t, strings.HasPrefix(*f.Name, "/dir/f-") || strings.HasPrefix(*f.Name, "dir/f-"))
	}
}
