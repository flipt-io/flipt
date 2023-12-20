package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func main() {
	var blobURL, testdataDir, container string
	flag.StringVar(&blobURL, "url", "", "Address for target azurite blob service")
	flag.StringVar(&testdataDir, "testdata-dir", "", "Directory path to testdata")
	flag.StringVar(&container, "container", "testdata", "Azurite blob container")
	flag.Parse()

	fatalOnError := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	if blobURL == "" {
		log.Fatal("Must supply non-empty -url flag value.")
	}

	fmt.Fprintln(os.Stderr, "Syncing data to azurite blob at", blobURL)

	ctx := context.Background()

	credentials, err := azblob.NewSharedKeyCredential(
		os.Getenv("AZURE_STORAGE_ACCOUNT"),
		os.Getenv("AZURE_STORAGE_KEY"),
	)
	fatalOnError(err)
	client, err := azblob.NewClientWithSharedKeyCredential(blobURL, credentials, nil)
	fatalOnError(err)

	fmt.Fprintln(os.Stderr, "Using azurite blob container", container)
	_, err = client.CreateContainer(ctx, container, nil)
	fatalOnError(err)

	dir := os.DirFS(testdataDir)
	fatalOnError(err)

	// copy testdata into target s3 bucket
	err = fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fmt.Fprintln(os.Stderr, "Copying", path)

		f, err := dir.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = client.UploadStream(ctx, container, path, f, nil)

		return err
	})
	fatalOnError(err)
}
