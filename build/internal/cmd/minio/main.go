package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	var minioURL = flag.String("minio-url", "", "Address for target minio service")
	var testdataDir = flag.String("testdata-dir", "", "Directory path to testdata")
	flag.Parse()

	fatalOnError := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	if *minioURL == "" {
		log.Fatal("Must supply non-empty --minio-url flag value.")
	}

	minioRegion := "minio"

	fmt.Fprintln(os.Stderr, "Syncing data to minio at", *minioURL, minioRegion)

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	fatalOnError(err)

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = minioURL
		o.UsePathStyle = true
		o.Region = minioRegion
	})

	output, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	fatalOnError(err)
	for i := range output.Buckets {
		fmt.Fprintln(os.Stderr, "S3 buckets", *output.Buckets[i].Name)
	}

	// cd /opt/bin && gzip -d mc.gz && chmod 755 mc && mc alias set mycloud http://localhost:9000&& mc admin trace -a mycloud
	bucketName := "testdata"
	fmt.Fprintln(os.Stderr, "S3 bucket name", bucketName)
	_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	fatalOnError(err)

	dir := os.DirFS(*testdataDir)
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
		// use path to get full path to file
		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &bucketName,
			Key:    &path,
			Body:   f,
		})

		return err
	})
	fatalOnError(err)
}
