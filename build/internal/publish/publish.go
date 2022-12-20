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
)

const (
	localScheme  = "docker-local://"
	remoteScheme = "docker://"
)

func Publish(ctx context.Context, path string, container *dagger.Container) (string, error) {
	switch {
	case strings.HasPrefix(path, localScheme):
		return local(ctx, path[len(localScheme):], container)
	case strings.HasPrefix(path, remoteScheme):
		return remote(ctx, path[len(remoteScheme):], container)
	}

	return "", errors.New("unexpected publish scheme")
}

func local(ctx context.Context, path string, container *dagger.Container) (string, error) {
	tar, err := tempFile()
	if err != nil {
		return "", err
	}
	defer os.Remove(tar)

	if _, err := container.Export(ctx, tar); err != nil {
		return "", err
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}

	fi, err := os.Open(tar)
	if err != nil {
		return "", err
	}
	defer fi.Close()

	resp, err := cli.ImageLoad(ctx, fi, false)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.JSON {
		var data map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return "", err
		}

		stream, ok := data["stream"]
		if !ok {
			fmt.Printf("Load Response: %q\n", data)
			return "", nil
		}

		id := strings.TrimSpace(stream[strings.Index(stream, "sha256:"):])
		if err := cli.ImageTag(ctx, id, path); err != nil {
			return "", err
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

func remote(ctx context.Context, path string, container *dagger.Container) (string, error) {
	return "", nil
}

func tempFile() (string, error) {
	fi, err := os.CreateTemp("", "flipt-image-*.tar")
	if err != nil {
		return "", err
	}
	defer fi.Close()

	return fi.Name(), nil
}
