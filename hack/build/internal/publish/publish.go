package publish

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"dagger.io/dagger"
	"github.com/docker/docker/client"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/nodeutils"
	"sigs.k8s.io/kind/pkg/cmd"
)

type TargetType string

const (
	LocalTargetType  TargetType = "local"
	RemoteTargetType TargetType = "remote"
	KindTargetType   TargetType = "kind"
)

type PublishSpec struct {
	TargetType  TargetType `json:"type"`
	KindCluster string     `json:"kind_cluster"`
	Target      string     `json:"target"`
}

type Variants map[dagger.Platform]*dagger.Container

func (v Variants) ToSlice() (dst []*dagger.Container) {
	dst = make([]*dagger.Container, 0, len(v))
	for _, vn := range v {
		dst = append(dst, vn)
	}
	return
}

func Publish(ctx context.Context, spec PublishSpec, daggerClient *dagger.Client, variants Variants) (string, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return "", fmt.Errorf("publish: %w", err)
	}

	if spec.TargetType == RemoteTargetType {
		return remote(ctx, spec.Target, daggerClient, variants)
	}

	// in local environment only variant matters (the local platform)
	// also, docker load doesn't support multi-platform OCI images
	// see: https://stackoverflow.com/questions/72945407/how-do-i-import-and-run-a-multi-platform-oci-image-in-docker-for-macos
	platform, err := daggerClient.DefaultPlatform(ctx)
	if err != nil {
		return "", err
	}

	container, ok := variants[platform]
	if !ok {
		return "", fmt.Errorf("platform not found in variants %q", platform)
	}

	ref, err := local(ctx, cli, spec.Target, container)
	if err != nil {
		return "", err
	}

	switch spec.TargetType {
	case KindTargetType:
		return spec.Target, kind(ctx, cli, spec.KindCluster, ref)
	case LocalTargetType:
		return spec.Target, nil
	default:
		return "", errors.New("unexpected target type")
	}
}

func kind(ctx context.Context, cli *client.Client, clusterName, ref string) error {
	provider := cluster.NewProvider(
		cluster.ProviderWithDocker(),
		cluster.ProviderWithLogger(cmd.NewLogger()),
	)
	nodes, err := provider.ListNodes(clusterName)
	if err != nil {
		return fmt.Errorf("kind: listing nodes: %w", err)
	}

	if len(nodes) < 1 {
		return errors.New("node not found")
	}

	rd, err := cli.ImageSave(ctx, []string{ref})
	if err != nil {
		return fmt.Errorf("kind: saving image (reference %q): %w", ref, err)
	}

	fi, err := os.CreateTemp("", "kind-image-*.tar")
	if err != nil {
		return err
	}
	defer func() {
		fi.Close()
		os.Remove(fi.Name())
	}()

	if err := nodeutils.LoadImageArchive(nodes[0], rd); err != nil {
		return fmt.Errorf("kind: loading image archive: %w", err)
	}

	//return nodeutils.ReTagImage(nodes[0], ref, path)
	return nil
}

func exportImage(ctx context.Context, container *dagger.Container) (string, error) {
	tar, err := tempFile()
	if err != nil {
		return "", err
	}

	if _, err := container.Export(ctx, tar); err != nil {
		return "", err
	}

	return tar, nil
}

func local(ctx context.Context, cli *client.Client, path string, container *dagger.Container) (string, error) {
	tar, err := tempFile()
	if err != nil {
		return "", err
	}
	defer os.Remove(tar)

	if _, err := container.Export(ctx, tar); err != nil {
		return "", fmt.Errorf("local: export image: %w", err)
	}

	fi, err := os.Open(tar)
	if err != nil {
		return "", fmt.Errorf("local: open tar: %w", err)
	}
	defer fi.Close()

	resp, err := cli.ImageLoad(ctx, fi, false)
	if err != nil {
		return "", fmt.Errorf("local: load image: %w", err)
	}
	defer resp.Body.Close()

	if resp.JSON {
		decoder := json.NewDecoder(resp.Body)
		var last map[string]string
		for {
			err := decoder.Decode(&last)
			if errors.Is(err, io.EOF) {
				break
			}
		}

		stream, ok := last["stream"]
		if !ok {
			return "", errors.New("local: parsing response: stream not found")
		}

		id := strings.TrimSpace(stream[strings.Index(stream, "sha256:"):])
		if err := cli.ImageTag(ctx, id, path); err != nil {
			return "", fmt.Errorf("local: tag image: %w", err)
		}

		return path, nil
	} else {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		fmt.Println("Load Response:", string(data))
	}

	return "", nil
}

func remote(ctx context.Context, path string, client *dagger.Client, variants Variants) (string, error) {
	container := client.Container()
	opts := dagger.ContainerPublishOpts{
		PlatformVariants: variants.ToSlice(),
	}

	// if we only have a single variant then skip doing
	// multi platform variant build
	if len(variants) == 1 {
		for _, variant := range variants {
			container = variant
			opts.PlatformVariants = nil
		}
	}

	return container.Publish(ctx, path, opts)
}

func tempFile() (string, error) {
	fi, err := os.CreateTemp("", "build-image-*.tar")
	if err != nil {
		return "", err
	}
	defer fi.Close()

	return fi.Name(), nil
}
