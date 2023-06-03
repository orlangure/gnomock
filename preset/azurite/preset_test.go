package azurite_test

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/orlangure/gnomock/preset/azurite"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/orlangure/gnomock"
)

func TestPreset_Blobstorage(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"latest"} {
		t.Run(version, testBlobStorage(version))
	}
}

func testBlobStorage(version string) func(*testing.T) {
	return func(t *testing.T) {
		p := azurite.Preset(
			azurite.WithVersion(version),
		)
		c, err := gnomock.Start(p)

		defer func() { require.NoError(t, gnomock.Stop(c)) }()

		require.NoError(t, err)

		connString := fmt.Sprintf(azurite.ConnectionStringFormat, azurite.AccountName, azurite.AccountKey, c.Address(azurite.BlobServicePort), azurite.AccountName)

		ctx := context.Background()

		azblobClient, err := azblob.NewClientFromConnectionString(connString, nil)
		require.NoError(t, err)

		containerName := "foo"
		_, err = azblobClient.CreateContainer(ctx, containerName, nil)
		require.NoError(t, err)

		pager := azblobClient.NewListBlobsFlatPager(containerName, nil)
		pages := 0
		for pager.More() {
			resp, err := pager.NextPage(context.Background())
			require.NoError(t, err)
			require.Equal(t, 0, len(resp.Segment.BlobItems))
			pages = pages + 1
		}
		require.Equal(t, 1, pages)

		blobName := "bar"
		_, err = azblobClient.UploadBuffer(ctx, containerName, blobName, []byte{15, 16, 17}, nil)
		require.NoError(t, err)

		pager = azblobClient.NewListBlobsFlatPager(containerName, nil)
		require.Equal(t, pager.More(), true)

		for pager.More() {
			resp, err := pager.NextPage(context.Background())
			require.NoError(t, err)
			require.Equal(t, 1, len(resp.Segment.BlobItems))
			for _, v := range resp.Segment.BlobItems {
				require.Equal(t, blobName, *v.Name)
			}
		}
	}
}

func TestPreset_wrongBlobstoragePath(t *testing.T) {
	t.Parallel()

	p := azurite.Preset(
		azurite.WithBlobstorageFiles("./invalid"),
	)
	c, err := gnomock.Start(p)
	require.Error(t, err)
	require.Contains(t, err.Error(), "can't read blobstorage initial files")
	require.NoError(t, gnomock.Stop(c))
}

func ExamplePresetBlobStorage() {
	p := azurite.Preset(
		azurite.WithVersion(azurite.DefaultVersion),
	)
	c, startErr := gnomock.Start(p)
	if startErr != nil {
		fmt.Println("Starting azurite gnomock failed ", startErr)
		return
	}

	defer func() { _ = gnomock.Stop(c) }()

	connString := fmt.Sprintf(azurite.ConnectionStringFormat, azurite.AccountName, azurite.AccountKey, c.Address(azurite.BlobServicePort), azurite.AccountName)
	ctx := context.Background()

	azblobClient, connectError := azblob.NewClientFromConnectionString(connString, nil)
	if connectError != nil {
		fmt.Println("Connecting to azurite failed ", connectError)
		return
	}

	containerName := "foo"
	_, createContainerError := azblobClient.CreateContainer(ctx, containerName, nil)
	if createContainerError != nil {
		fmt.Println("Creating azure container failed ", createContainerError)
		return
	}

	pager := azblobClient.NewListBlobsFlatPager(containerName, nil)
	pages := 0
	for pager.More() {
		resp, _ := pager.NextPage(context.Background())
		fmt.Println("keys before:", len(resp.Segment.BlobItems))
		pages = pages + 1
	}
	fmt.Println("pages before:", pages)

	blobName := "bar"
	_, _ = azblobClient.UploadBuffer(ctx, containerName, blobName, []byte{15, 16, 17}, nil)

	pager = azblobClient.NewListBlobsFlatPager(containerName, nil)

	pages = 0
	for pager.More() {
		resp, _ := pager.NextPage(context.Background())

		fmt.Println("keys after:", len(resp.Segment.BlobItems))
		for _, v := range resp.Segment.BlobItems {
			fmt.Println("filename:", *v.Name)
		}
		pages = pages + 1
	}
	fmt.Println("pages after:", 1)

	//Output:
	//keys before: 0
	//pages before: 1
	//keys after: 1
	//filename: bar
	//pages after: 1
}
