package main

import (
	"fmt"
	"net/url"

	sdk "go.flipt.io/flipt/sdk/go"
	sdkgrpc "go.flipt.io/flipt/sdk/go/grpc"
	sdkhttp "go.flipt.io/flipt/sdk/go/http"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func fliptSDK(address, token string) (*sdk.SDK, error) {
	addr, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("export address is invalid %w", err)
	}

	var transport sdk.Transport
	switch addr.Scheme {
	case "http", "https":
		transport = sdkhttp.NewTransport(address)
	case "grpc":
		conn, err := grpc.Dial(addr.Host,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("failed to dial Flipt %w", err)
		}

		transport = sdkgrpc.NewTransport(conn)
	default:
		return nil, fmt.Errorf("unexpected protocol %s", addr.Scheme)
	}

	var opts []sdk.Option
	if token != "" {
		opts = append(opts, sdk.WithAuthenticationProvider(
			sdk.StaticTokenAuthenticationProvider(token),
		))
	}
	s := sdk.New(transport, opts...)
	return &s, nil
}
