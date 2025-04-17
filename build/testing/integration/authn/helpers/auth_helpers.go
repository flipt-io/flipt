package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/v2/environments"
	sdk "go.flipt.io/flipt/sdk/go"
	sdkv2 "go.flipt.io/flipt/sdk/go/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CanReadAllIn tests that the client can read flags in the specified namespace
func CanReadAllIn(t *testing.T, ctx context.Context, client sdk.SDK, namespace string) {
	t.Run("CanReadAll", func(t *testing.T) {
		_, err := client.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
			NamespaceKey: namespace,
		})
		AssertErrIsUnauthenticated(t, err, false)
	})
}

// CannotReadAnyIn tests that the client cannot read flags in the specified namespace
func CannotReadAnyIn(t *testing.T, ctx context.Context, client sdk.SDK, namespace string) {
	t.Run("CannotReadAny", func(t *testing.T) {
		_, err := client.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
			NamespaceKey: namespace,
		})
		AssertErrIsUnauthenticated(t, err, true)
	})
}

// CanReadEnvironmentsAllIn tests that the client can read environments
func CanReadEnvironmentsAllIn(t *testing.T, ctx context.Context, client sdkv2.SDK) {
	t.Run("EnvironmentsCanReadAll", func(t *testing.T) {
		// ensure we can do resource specific operations across namespaces
		client := client.Environments()
		_, err := client.ListEnvironments(ctx, &environments.ListEnvironmentsRequest{})
		AssertErrIsUnauthenticated(t, err, false)
	})
}

// CannotReadEnvironmentsAnyIn tests that the client cannot read environments
func CannotReadEnvironmentsAnyIn(t *testing.T, ctx context.Context, client sdkv2.SDK) {
	t.Run("EnvironmentsCannotReadAny", func(t *testing.T) {
		// ensure we cannot do resource specific operations across namespaces
		client := client.Environments()
		_, err := client.ListEnvironments(ctx, &environments.ListEnvironmentsRequest{})
		AssertErrIsUnauthenticated(t, err, true)
	})
}

// AssertErrIsUnauthenticated checks if the error is an Unauthenticated error or not based on the expected value
func AssertErrIsUnauthenticated(t *testing.T, err error, unauthenticated bool) {
	if !assert.Equal(
		t,
		codes.Unauthenticated == status.Code(err),
		unauthenticated,
		"expected unauthenticated error",
	) {
		t.Logf("error: %v", err)
	}
}
