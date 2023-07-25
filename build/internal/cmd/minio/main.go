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

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		fmt.Fprintln(os.Stderr, "Endpoint resolver", service, region)
		if service == s3.ServiceID {
			return aws.Endpoint{
				PartitionID:       "aws",
				URL:               *minioURL,
				HostnameImmutable: true,
				SigningRegion:     "",
			}, nil
		}
		fmt.Fprintln(os.Stderr, "Unknown Endpoint", service, region)
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	})

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(customResolver),
	)
	fatalOnError(err)

	s3Client := s3.NewFromConfig(cfg)

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

		name := d.Name()
		f, err := dir.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: &bucketName,
			Key:    &name,
			Body:   f,
		})

		return err
	})
	fatalOnError(err)
}
