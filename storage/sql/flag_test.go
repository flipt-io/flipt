package sql

import (
	"context"
	"encoding/json"
	"fmt"

	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/storage"
	"go.flipt.io/flipt/storage/sql/common"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *DBTestSuite) TestGetFlag() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	got, err := s.store.GetFlag(context.TODO(), flag.Key)

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, flag.Key, got.Key)
	assert.Equal(t, flag.Name, got.Name)
	assert.Equal(t, flag.Description, got.Description)
	assert.Equal(t, flag.Enabled, got.Enabled)
	assert.NotZero(t, flag.CreatedAt)
	assert.NotZero(t, flag.UpdatedAt)
}

func (s *DBTestSuite) TestGetFlagNotFound() {
	t := s.T()

	_, err := s.store.GetFlag(context.TODO(), "foo")
	assert.EqualError(t, err, "flag \"foo\" not found")
}

func (s *DBTestSuite) TestListFlags() {
	t := s.T()

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
		_, err := s.store.CreateFlag(context.TODO(), req)
		require.NoError(t, err)
	}

	res, err := s.store.ListFlags(context.TODO())
	require.NoError(t, err)
	got := res.Results
	assert.NotZero(t, len(got))
}

func (s *DBTestSuite) TestListFlagsPagination_LimitOffset() {
	t := s.T()

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
		_, err := s.store.CreateFlag(context.TODO(), req)
		require.NoError(t, err)
	}

	res, err := s.store.ListFlags(context.TODO(), storage.WithOrder(storage.OrderDesc), storage.WithLimit(1), storage.WithOffset(1))
	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)
}

func (s *DBTestSuite) TestListFlagsPagination_LimitWithNextPage() {
	t := s.T()

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
		_, err := s.store.CreateFlag(context.TODO(), req)
		require.NoError(t, err)
	}

	// TODO: the ordering (DESC) is required because the default ordering is ASC and we are not clearing the DB between tests
	opts := []storage.QueryOption{storage.WithOrder(storage.OrderDesc), storage.WithLimit(1)}

	res, err := s.store.ListFlags(context.TODO(), opts...)
	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, reqs[1].Key, got[0].Key)
	assert.NotEmpty(t, res.NextPageToken)

	pageToken := &common.PageToken{}
	err = json.Unmarshal([]byte(res.NextPageToken), pageToken)
	require.NoError(t, err)
	assert.Equal(t, reqs[0].Key, pageToken.Key)
	assert.NotZero(t, pageToken.Offset)

	opts = append(opts, storage.WithPageToken(res.NextPageToken))

	res, err = s.store.ListFlags(context.TODO(), opts...)
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, reqs[0].Key, got[0].Key)
}

func (s *DBTestSuite) TestCreateFlag() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
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

func (s *DBTestSuite) TestCreateFlag_DuplicateKey() {
	t := s.T()

	_, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)

	_, err = s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	assert.EqualError(t, err, "flag \"TestDBTestSuite/TestCreateFlag_DuplicateKey\" is not unique")
}

func (s *DBTestSuite) TestUpdateFlag() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
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

	updated, err := s.store.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{
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

func (s *DBTestSuite) TestUpdateFlag_NotFound() {
	t := s.T()

	_, err := s.store.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	assert.EqualError(t, err, "flag \"foo\" not found")
}

func (s *DBTestSuite) TestDeleteFlag() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	err = s.store.DeleteFlag(context.TODO(), &flipt.DeleteFlagRequest{Key: flag.Key})
	require.NoError(t, err)
}

func (s *DBTestSuite) TestDeleteFlag_NotFound() {
	t := s.T()

	err := s.store.DeleteFlag(context.TODO(), &flipt.DeleteFlagRequest{Key: "foo"})
	require.NoError(t, err)
}

func (s *DBTestSuite) TestCreateVariant() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	attachment := `{"key":"value"}`
	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
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
	flag, err = s.store.GetFlag(context.TODO(), flag.Key)

	require.NoError(t, err)
	assert.NotNil(t, flag)

	assert.Len(t, flag.Variants, 1)
}

func (s *DBTestSuite) TestCreateVariant_FlagNotFound() {
	t := s.T()

	_, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     "foo",
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "flag \"foo\" not found")
}

func (s *DBTestSuite) TestCreateVariant_DuplicateName() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	// try to create another variant with the same name for this flag
	_, err = s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "variant \"foo\" is not unique")
}

func (s *DBTestSuite) TestCreateVariant_DuplicateName_DifferentFlag() {
	t := s.T()

	flag1, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         fmt.Sprintf("%s_1", t.Name()),
		Name:        "foo_1",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag1)

	variant1, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
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

	flag2, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         fmt.Sprintf("%s_2", t.Name()),
		Name:        "foo_2",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag2)

	variant2, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
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

func (s *DBTestSuite) TestUpdateVariant() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	attachment1 := `{"key":"value1"}`
	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
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

	updated, err := s.store.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{
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
	flag, err = s.store.GetFlag(context.TODO(), flag.Key)

	require.NoError(t, err)
	assert.NotNil(t, flag)

	assert.Len(t, flag.Variants, 1)
}

func (s *DBTestSuite) TestUpdateVariant_NotFound() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	_, err = s.store.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{
		Id:          "foo",
		FlagKey:     flag.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "variant \"foo\" not found")
}

func (s *DBTestSuite) TestUpdateVariant_DuplicateName() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant1, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant1)

	variant2, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         "bar",
		Name:        "bar",
		Description: "baz",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant2)

	_, err = s.store.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{
		Id:          variant2.Id,
		FlagKey:     variant2.FlagKey,
		Key:         variant1.Key,
		Name:        variant2.Name,
		Description: "foobar",
	})

	assert.EqualError(t, err, "variant \"foo\" is not unique")
}

func (s *DBTestSuite) TestDeleteVariant() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	err = s.store.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{FlagKey: variant.FlagKey, Id: variant.Id})
	require.NoError(t, err)

	// get the flag again
	flag, err = s.store.GetFlag(context.TODO(), flag.Key)

	require.NoError(t, err)
	assert.NotNil(t, flag)

	assert.Empty(t, flag.Variants)
}

func (s *DBTestSuite) TestDeleteVariant_ExistingRule() {
	t := s.T()

	// TODO
	t.SkipNow()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:     flag.Key,
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, segment)

	rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	require.NoError(t, err)
	assert.NotNil(t, rule)

	// try to delete variant with attached rule
	err = s.store.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{
		Id:      variant.Id,
		FlagKey: flag.Key,
	})

	assert.EqualError(t, err, "atleast one rule exists that includes this variant")

	// delete the rule, then try to delete the variant again
	err = s.store.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{
		Id:      rule.Id,
		FlagKey: flag.Key,
	})

	require.NoError(t, err)

	err = s.store.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{
		Id:      variant.Id,
		FlagKey: flag.Key,
	})

	require.NoError(t, err)
}

func (s *DBTestSuite) TestDeleteVariant_NotFound() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	err = s.store.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{
		Id:      "foo",
		FlagKey: flag.Key,
	})

	require.NoError(t, err)
}
