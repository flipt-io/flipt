package main

import (
	"context"
	"os"

	fliptoci "go.flipt.io/flipt/internal/oci"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/content/oci"
)

func main() {
	logger, _ := zap.NewDevelopment()
	packager := fliptoci.NewPackager(logger)

	ctx := context.Background()
	src := memory.New()

	tempDir, err := os.MkdirTemp("", "oras_oci_example_*")
	if err != nil {
		panic(err) // Handle error
	}

	dst, err := oci.New(tempDir)
	if err != nil {
		panic(err)
	}

	manifestDesc, err := packager.Package(ctx, src, os.DirFS("."))
	if err != nil {
		panic(err)
	}

	logger.Info("tagging manifest", zap.String("ref", "latest"), zap.String("digest", manifestDesc.Digest.Hex()))

	if err := src.Tag(ctx, manifestDesc, "latest"); err != nil {
		panic(err)
	}

	desc, err := oras.Copy(ctx, src, "latest", dst, "latest", oras.DefaultCopyOptions)
	if err != nil {
		panic(err)
	}

	logger.Info("copied", zap.String("path", tempDir), zap.String("digest", desc.Digest.Hex()))
}
