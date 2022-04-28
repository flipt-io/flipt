package sql

import (
	"context"
	"fmt"
	"testing"
	"time"

	flipt "github.com/markphelps/flipt/rpc/flipt"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/storage/cache"
	"github.com/markphelps/flipt/storage/cache/memory"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFlag(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	got, err := store.GetFlag(context.TODO(), flag.Key)

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
	_, err := store.GetFlag(context.TODO(), "foo")
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
		_, err := store.CreateFlag(context.TODO(), req)
		require.NoError(t, err)
	}

	got, err := store.ListFlags(context.TODO())
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
		_, err := store.CreateFlag(context.TODO(), req)
		require.NoError(t, err)
	}

	got, err := store.ListFlags(context.TODO(), storage.WithLimit(1), storage.WithOffset(1))
	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestCreateFlag(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
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
	_, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)

	_, err = store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	assert.EqualError(t, err, "flag \"TestCreateFlag_DuplicateKey\" is not unique")
}

func TestUpdateFlag(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
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

	updated, err := store.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{
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
	assert.NotZero(t, updated.UpdatedAt)
}

func TestUpdateFlag_NotFound(t *testing.T) {
	_, err := store.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	assert.EqualError(t, err, "flag \"foo\" not found")
}

func TestDeleteFlag(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	err = store.DeleteFlag(context.TODO(), &flipt.DeleteFlagRequest{Key: flag.Key})
	require.NoError(t, err)
}

func TestDeleteFlag_NotFound(t *testing.T) {
	err := store.DeleteFlag(context.TODO(), &flipt.DeleteFlagRequest{Key: "foo"})
	require.NoError(t, err)
}

func TestCreateVariant(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	attachment := `{"key":"value"}`
	variant, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Attachment:  attachment,
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	assert.NotZero(t, variant.Id)
	assert.Equal(t, flag.Key, variant.FlagKey)
	assert.Equal(t, t.Name(), variant.Key)
	assert.Equal(t, "foo", variant.Name)
	assert.Equal(t, "bar", variant.Description)
	assert.Equal(t, attachment, variant.Attachment)
	assert.NotZero(t, variant.CreatedAt)
	assert.Equal(t, variant.CreatedAt.Seconds, variant.UpdatedAt.Seconds)

	// get the flag again
	flag, err = store.GetFlag(context.TODO(), flag.Key)

	require.NoError(t, err)
	assert.NotNil(t, flag)

	assert.Len(t, flag.Variants, 1)
}

func TestCreateVariant_FlagNotFound(t *testing.T) {
	_, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     "foo",
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "flag \"foo\" not found")
}

func TestCreateVariant_DuplicateName(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	// try to create another variant with the same name for this flag
	_, err = store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "variant \"foo\" is not unique")
}

func TestCreateVariant_DuplicateName_DifferentFlag(t *testing.T) {
	flag1, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         fmt.Sprintf("%s_1", t.Name()),
		Name:        "foo_1",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag1)

	variant1, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag1.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant1)

	assert.NotZero(t, variant1.Id)
	assert.Equal(t, flag1.Key, variant1.FlagKey)
	assert.Equal(t, "foo", variant1.Key)

	flag2, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         fmt.Sprintf("%s_2", t.Name()),
		Name:        "foo_2",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag2)

	variant2, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag2.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant2)

	assert.NotZero(t, variant2.Id)
	assert.Equal(t, flag2.Key, variant2.FlagKey)
	assert.Equal(t, "foo", variant2.Key)
}

func TestUpdateVariant(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	attachment1 := `{"key":"value1"}`
	variant, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
		Attachment:  attachment1,
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	assert.NotZero(t, variant.Id)
	assert.Equal(t, flag.Key, variant.FlagKey)
	assert.Equal(t, "foo", variant.Key)
	assert.Equal(t, "foo", variant.Name)
	assert.Equal(t, "bar", variant.Description)
	assert.Equal(t, attachment1, variant.Attachment)
	assert.NotZero(t, variant.CreatedAt)
	assert.Equal(t, variant.CreatedAt.Seconds, variant.UpdatedAt.Seconds)

	updated, err := store.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{
		Id:          variant.Id,
		FlagKey:     variant.FlagKey,
		Key:         variant.Key,
		Name:        variant.Name,
		Description: "foobar",
		Attachment:  `{"key":      "value2"}`,
	})

	require.NoError(t, err)

	assert.Equal(t, variant.Id, updated.Id)
	assert.Equal(t, variant.FlagKey, updated.FlagKey)
	assert.Equal(t, variant.Key, updated.Key)
	assert.Equal(t, variant.Name, updated.Name)
	assert.Equal(t, "foobar", updated.Description)
	assert.Equal(t, `{"key":"value2"}`, updated.Attachment)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotZero(t, updated.UpdatedAt)

	// get the flag again
	flag, err = store.GetFlag(context.TODO(), flag.Key)

	require.NoError(t, err)
	assert.NotNil(t, flag)

	assert.Len(t, flag.Variants, 1)
}

func TestUpdateVariant_NotFound(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	_, err = store.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{
		Id:          "foo",
		FlagKey:     flag.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "variant \"foo\" not found")
}

func TestUpdateVariant_DuplicateName(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant1, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant1)

	variant2, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         "bar",
		Name:        "bar",
		Description: "baz",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant2)

	_, err = store.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{
		Id:          variant2.Id,
		FlagKey:     variant2.FlagKey,
		Key:         variant1.Key,
		Name:        variant2.Name,
		Description: "foobar",
	})

	assert.EqualError(t, err, "variant \"foo\" is not unique")
}

func TestDeleteVariant(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	err = store.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{FlagKey: variant.FlagKey, Id: variant.Id})
	require.NoError(t, err)

	// get the flag again
	flag, err = store.GetFlag(context.TODO(), flag.Key)

	require.NoError(t, err)
	assert.NotNil(t, flag)

	assert.Empty(t, flag.Variants)
}

func TestDeleteVariant_ExistingRule(t *testing.T) {
	// TODO
	t.SkipNow()

	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	rule, err := store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	// try to delete variant with attached rule
	err = store.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{
		Id:      variant.Id,
		FlagKey: flag.Key,
	})

	assert.EqualError(t, err, "atleast one rule exists that includes this variant")

	// delete the rule, then try to delete the variant again
	err = store.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{
		Id:      rule.Id,
		FlagKey: flag.Key,
	})

	require.NoError(t, err)

	err = store.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{
		Id:      variant.Id,
		FlagKey: flag.Key,
	})

	require.NoError(t, err)
}

func TestDeleteVariant_NotFound(t *testing.T) {
	flag, err := store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	err = store.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{
		Id:      "foo",
		FlagKey: flag.Key,
	})

	require.NoError(t, err)
}

var benchFlag *flipt.Flag

func BenchmarkGetFlag(b *testing.B) {
	var (
		ctx       = context.Background()
		flag, err = store.CreateFlag(ctx, &flipt.CreateFlagRequest{
			Key:         b.Name(),
			Name:        "foo",
			Description: "bar",
			Enabled:     true,
		})
	)

	if err != nil {
		b.Fatal(err)
	}

	_, err = store.CreateVariant(ctx, &flipt.CreateVariantRequest{
		FlagKey: flag.Key,
		Key:     "baz",
	})

	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	b.Run("get-flag", func(b *testing.B) {
		var f *flipt.Flag

		for i := 0; i < b.N; i++ {
			f, _ = store.GetFlag(context.TODO(), flag.Key)
		}

		benchFlag = f
	})
}

func BenchmarkGetFlag_CacheMemory(b *testing.B) {
	var (
		l, _       = test.NewNullLogger()
		logger     = logrus.NewEntry(l)
		cacher     = memory.NewCache(5*time.Minute, 10*time.Minute, logger)
		storeCache = cache.NewStore(logger, cacher, store)

		ctx = context.Background()

		flag, err = storeCache.CreateFlag(ctx, &flipt.CreateFlagRequest{
			Key:         b.Name(),
			Name:        "foo",
			Description: "bar",
			Enabled:     true,
		})
	)

	if err != nil {
		b.Fatal(err)
	}

	_, err = storeCache.CreateVariant(ctx, &flipt.CreateVariantRequest{
		FlagKey: flag.Key,
		Key:     "baz",
	})

	if err != nil {
		b.Fatal(err)
	}

	var f *flipt.Flag

	// warm the cache
	f, _ = storeCache.GetFlag(context.TODO(), flag.Key)

	b.ResetTimer()

	b.Run("get-flag-cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			f, _ = storeCache.GetFlag(context.TODO(), flag.Key)
		}

		benchFlag = f
	})
}
