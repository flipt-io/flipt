package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"

	gstorage "cloud.google.com/go/storage"
)

func main() {
	var testdataDir, bucket string
	flag.StringVar(&testdataDir, "testdata-dir", "", "Directory path to testdata")
	flag.StringVar(&bucket, "bucket", "testdata", "Google Cloud Storage bucket")
	flag.Parse()

	fatalOnError := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	blobURL := os.Getenv("STORAGE_EMULATOR_HOST")

	if blobURL == "" {
		log.Fatal("Must supply non-empty env STORAGE_EMULATOR_HOST value.")
	}

	fmt.Fprintln(os.Stderr, "Syncing data to gcs blob at", blobURL)

	ctx := context.Background()
	client, err := gstorage.NewClient(ctx)
	fatalOnError(err)
	defer client.Close()
	fmt.Fprintln(os.Stderr, "Using GCS bucket", bucket)
	bkt := client.Bucket(bucket)
	err = bkt.Create(ctx, "", nil)
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

		w := bkt.Object(path).NewWriter(ctx)
		_, err = io.Copy(w, f)
		if err != nil {
			return err
		}

		return w.Close()
	})
	fatalOnError(err)
}
