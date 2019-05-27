// +build db

package storage

import (
	"context"
	"testing"

	flipt "github.com/markphelps/flipt/rpc"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFlag(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	got, err := flagStore.GetFlag(context.TODO(), &flipt.GetFlagRequest{Key: flag.Key})

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, flag.Key, got.Key)
	assert.Equal(t, flag.Name, got.Name)
	assert.Equal(t, flag.Description, got.Description)
	assert.Equal(t, flag.Enabled, got.Enabled)
	assert.NotZero(t, flag.CreatedAt)
	assert.NotZero(t, flag.UpdatedAt)
}

func TestGetFlagNotFound(t *testing.T) {
	_, err := flagStore.GetFlag(context.TODO(), &flipt.GetFlagRequest{Key: "foo"})
	assert.EqualError(t, err, "flag \"foo\" not found")
}

func TestListFlags(t *testing.T) {
	reqs := []*flipt.CreateFlagRequest{
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
			Enabled:     true,
		},
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
		},
	}

	for _, req := range reqs {
		_, err := flagStore.CreateFlag(context.TODO(), req)
		require.NoError(t, err)
	}

	got, err := flagStore.ListFlags(context.TODO(), &flipt.ListFlagRequest{})
	require.NoError(t, err)
	assert.NotZero(t, len(got))
}

func TestFlagsPagination(t *testing.T) {
	reqs := []*flipt.CreateFlagRequest{
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
			Enabled:     true,
		},
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
		},
	}

	for _, req := range reqs {
		_, err := flagStore.CreateFlag(context.TODO(), req)
		require.NoError(t, err)
	}

	got, err := flagStore.ListFlags(context.TODO(), &flipt.ListFlagRequest{
		Limit:  1,
		Offset: 1,
	})
	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestCreateFlag(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)

	assert.Equal(t, t.Name(), flag.Key)
	assert.Equal(t, "foo", flag.Name)
	assert.Equal(t, "bar", flag.Description)
	assert.True(t, flag.Enabled)
	assert.NotZero(t, flag.CreatedAt)
	assert.Equal(t, flag.CreatedAt.Seconds, flag.UpdatedAt.Seconds)
}

func TestCreateFlag_DuplicateKey(t *testing.T) {
	_, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)

	_, err = flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	assert.EqualError(t, err, "flag \"TestCreateFlag_DuplicateKey\" is not unique")
}

func TestUpdateFlag(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)

	assert.Equal(t, t.Name(), flag.Key)
	assert.Equal(t, "foo", flag.Name)
	assert.Equal(t, "bar", flag.Description)
	assert.True(t, flag.Enabled)
	assert.NotZero(t, flag.CreatedAt)
	assert.Equal(t, flag.CreatedAt.Seconds, flag.UpdatedAt.Seconds)

	updated, err := flagStore.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{
		Key:         flag.Key,
		Name:        flag.Name,
		Description: "foobar",
		Enabled:     true,
	})

	require.NoError(t, err)

	assert.Equal(t, flag.Key, updated.Key)
	assert.Equal(t, flag.Name, updated.Name)
	assert.Equal(t, "foobar", updated.Description)
	assert.True(t, flag.Enabled)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotEqual(t, updated.CreatedAt.Seconds, updated.UpdatedAt.Seconds)
}

func TestUpdateFlag_NotFound(t *testing.T) {
	_, err := flagStore.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	assert.EqualError(t, err, "flag \"foo\" not found")
}

func TestDeleteFlag(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	err = flagStore.DeleteFlag(context.TODO(), &flipt.DeleteFlagRequest{Key: flag.Key})
	require.NoError(t, err)
}

func TestDeleteFlag_NotFound(t *testing.T) {
	err := flagStore.DeleteFlag(context.TODO(), &flipt.DeleteFlagRequest{Key: "foo"})
	require.NoError(t, err)
}

func TestCreateVariant(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	assert.NotZero(t, variant.Id)
	assert.Equal(t, flag.Key, variant.FlagKey)
	assert.Equal(t, t.Name(), variant.Key)
	assert.Equal(t, "foo", variant.Name)
	assert.Equal(t, "bar", variant.Description)
	assert.NotZero(t, variant.CreatedAt)
	assert.Equal(t, variant.CreatedAt.Seconds, variant.UpdatedAt.Seconds)

	// get the flag again
	flag, err = flagStore.GetFlag(context.TODO(), &flipt.GetFlagRequest{Key: flag.Key})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	assert.Len(t, flag.Variants, 1)
}

func TestCreateVariant_FlagNotFound(t *testing.T) {
	_, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     "foo",
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "flag \"foo\" not found")
}

func TestUpdateVariant(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	assert.NotZero(t, variant.Id)
	assert.Equal(t, flag.Key, variant.FlagKey)
	assert.Equal(t, t.Name(), variant.Key)
	assert.Equal(t, "foo", variant.Name)
	assert.Equal(t, "bar", variant.Description)
	assert.NotZero(t, variant.CreatedAt)
	assert.Equal(t, variant.CreatedAt.Seconds, variant.UpdatedAt.Seconds)

	updated, err := flagStore.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{
		Id:          variant.Id,
		FlagKey:     variant.FlagKey,
		Key:         variant.Key,
		Name:        variant.Name,
		Description: "foobar",
	})

	require.NoError(t, err)

	assert.Equal(t, variant.Id, updated.Id)
	assert.Equal(t, variant.FlagKey, updated.FlagKey)
	assert.Equal(t, variant.Key, updated.Key)
	assert.Equal(t, variant.Name, updated.Name)
	assert.Equal(t, "foobar", updated.Description)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotEqual(t, updated.CreatedAt.Seconds, updated.UpdatedAt.Seconds)

	// get the flag again
	flag, err = flagStore.GetFlag(context.TODO(), &flipt.GetFlagRequest{Key: flag.Key})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	assert.Len(t, flag.Variants, 1)
}

func TestUpdateVariant_NotFound(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	_, err = flagStore.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{
		Id:          "foo",
		FlagKey:     flag.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "variant \"foo\" not found")
}

func TestDeleteVariant(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	err = flagStore.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{FlagKey: variant.FlagKey, Id: variant.Id})
	require.NoError(t, err)

	// get the flag again
	flag, err = flagStore.GetFlag(context.TODO(), &flipt.GetFlagRequest{Key: flag.Key})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	assert.Empty(t, flag.Variants)
}

func TestDeleteVariant_ExistingRule(t *testing.T) {
	t.SkipNow()

	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := flagStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := segmentStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	rule, err := ruleStore.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	// try to delete variant with attached rule
	err = flagStore.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{
		Id:      variant.Id,
		FlagKey: flag.Key,
	})

	assert.EqualError(t, err, "atleast one rule exists that includes this variant")

	// delete the rule, then try to delete the variant again
	err = ruleStore.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{
		Id:      rule.Id,
		FlagKey: flag.Key,
	})

	require.NoError(t, err)

	err = flagStore.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{
		Id:      variant.Id,
		FlagKey: flag.Key,
	})

	require.NoError(t, err)
}

func TestDeleteVariant_NotFound(t *testing.T) {
	flag, err := flagStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	err = flagStore.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{
		Id:      "foo",
		FlagKey: flag.Key,
	})

	require.NoError(t, err)
}
