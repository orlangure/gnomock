# Gnomock Azurite

Gnomock Blobstorage is a [Gnomock](https://github.com/orlangure/gnomock) preset
for running tests against Azure Blobstorage locally, powered by
[Azurite](https://github.com/Azure/Azurite) project. It allows
to setup a number of supported Azure services locally, run tests against
them, and tear them down easily.

See [Azurite](https://github.com/Azure/Azurite) documentation for
more details.

### Testing against local Azurite

```go
package azurite_test

import (
"context"
"fmt"
"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
"github.com/orlangure/gnomock"
)

func ExamplePresetBlobStorage() {
	p := azurite.Preset(
		azurite.WithVersion(azurite.DefaultVersion),
	)
	c, _ := gnomock.Start(p)

	defer func() { _ = gnomock.Stop(c) }()

	connString := fmt.Sprintf(azurite.ConnectionStringFormat, azurite.AccountName, azurite.AccountKey, c.Address(azurite.BlobServicePort), azurite.AccountName)
	ctx := context.Background()

	azblobClient, _ := azblob.NewClientFromConnectionString(connString, nil)

	containerName := "foo"
	_, _ = azblobClient.CreateContainer(ctx, containerName, nil)

	pager := azblobClient.NewListBlobsFlatPager(containerName, nil)
	pages := 0
	for pager.More() {
		resp, _ := pager.NextPage(context.TODO())
		fmt.Println("keys before:", len(resp.Segment.BlobItems))
		pages = pages + 1
	}
	fmt.Println("pages before:", pages)

	blobName := "bar"
	_, _ = azblobClient.UploadBuffer(ctx, containerName, blobName, []byte{15, 16, 17}, nil)

	pager = azblobClient.NewListBlobsFlatPager(containerName, nil)

	pages = 0
	for pager.More() {
		resp, _ := pager.NextPage(context.TODO())

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
```