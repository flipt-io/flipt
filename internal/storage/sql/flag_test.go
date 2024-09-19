package sql_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	"go.flipt.io/flipt/internal/storage/sql/common"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *DBTestSuite) TestGetFlag() {
	t := s.T()

	metadataMap := map[string]any{
		"key": "value",
	}

	metadata, err := structpb.NewStruct(metadataMap)
	require.NoError(t, err, "Failed to create metadata struct")

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
		Metadata:    metadata,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	got, err := s.store.GetFlag(context.TODO(), storage.NewResource(storage.DefaultNamespace, flag.Key))

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, storage.DefaultNamespace, got.NamespaceKey)
	assert.Equal(t, flag.Key, got.Key)
	assert.Equal(t, flag.Name, got.Name)
	assert.Equal(t, flag.Description, got.Description)
	assert.Equal(t, flag.Enabled, got.Enabled)
	assert.Equal(t, flag.Metadata.String(), got.Metadata.String())

	assert.NotZero(t, flag.CreatedAt)
	assert.NotZero(t, flag.UpdatedAt)
}

func (s *DBTestSuite) TestGetFlagNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	got, err := s.store.GetFlag(context.TODO(), storage.NewResource(s.namespace, flag.Key))

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, s.namespace, got.NamespaceKey)
	assert.Equal(t, flag.Key, got.Key)
	assert.Equal(t, flag.Name, got.Name)
	assert.Equal(t, flag.Description, got.Description)
	assert.Equal(t, flag.Enabled, got.Enabled)
	assert.NotZero(t, flag.CreatedAt)
	assert.NotZero(t, flag.UpdatedAt)
}

func (s *DBTestSuite) TestGetFlag_NotFound() {
	t := s.T()

	_, err := s.store.GetFlag(context.TODO(), storage.NewResource(storage.DefaultNamespace, "foo"))
	assert.EqualError(t, err, "flag \"default/foo\" not found")
}

func (s *DBTestSuite) TestGetFlagNamespace_NotFound() {
	t := s.T()

	_, err := s.store.GetFlag(context.TODO(), storage.NewResource(s.namespace, "foo"))
	assert.EqualError(t, err, fmt.Sprintf("flag \"%s/foo\" not found", s.namespace))
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

	_, err := s.store.ListFlags(context.TODO(), storage.ListWithOptions(
		storage.NewNamespace(storage.DefaultNamespace),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](storage.WithPageToken("Hello World"))),
	)
	require.EqualError(t, err, "pageToken is not valid: \"Hello World\"")

	res, err := s.store.ListFlags(context.TODO(), storage.ListWithOptions(storage.NewNamespace(storage.DefaultNamespace)))
	require.NoError(t, err)

	got := res.Results
	assert.NotEmpty(t, got)

	for _, flag := range got {
		assert.Equal(t, storage.DefaultNamespace, flag.NamespaceKey)
		assert.NotZero(t, flag.CreatedAt)
		assert.NotZero(t, flag.UpdatedAt)
	}
}

func (s *DBTestSuite) TestListFlagsNamespace() {
	t := s.T()

	reqs := []*flipt.CreateFlagRequest{
		{
			NamespaceKey: s.namespace,
			Key:          uuid.Must(uuid.NewV4()).String(),
			Name:         "foo",
			Description:  "bar",
			Enabled:      true,
		},
		{
			NamespaceKey: s.namespace,
			Key:          uuid.Must(uuid.NewV4()).String(),
			Name:         "foo",
			Description:  "bar",
		},
	}

	for _, req := range reqs {
		_, err := s.store.CreateFlag(context.TODO(), req)
		require.NoError(t, err)
	}

	res, err := s.store.ListFlags(context.TODO(), storage.ListWithOptions(storage.NewNamespace(s.namespace)))
	require.NoError(t, err)

	got := res.Results
	assert.NotEmpty(t, got)

	for _, flag := range got {
		assert.Equal(t, s.namespace, flag.NamespaceKey)
		assert.NotZero(t, flag.CreatedAt)
		assert.NotZero(t, flag.UpdatedAt)
	}
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
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
			Enabled:     true,
		},
	}

	for _, req := range reqs {
		if s.db.Driver == fliptsql.MySQL {
			// required for MySQL since it only s.stores timestamps to the second and not millisecond granularity
			time.Sleep(time.Second)
		}
		_, err := s.store.CreateFlag(context.TODO(), req)
		require.NoError(t, err)
	}

	oldest, middle, newest := reqs[0], reqs[1], reqs[2]

	// TODO: the ordering (DESC) is required because the default ordering is ASC and we are not clearing the DB between tests
	// get middle flag
	req := storage.ListWithOptions(
		storage.NewNamespace(storage.DefaultNamespace),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](
			storage.WithOrder(storage.OrderDesc), storage.WithLimit(1), storage.WithOffset(1),
		),
	)
	res, err := s.store.ListFlags(context.TODO(), req)
	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)

	assert.Equal(t, middle.Key, got[0].Key)

	// get first (newest) flag
	req = storage.ListWithOptions(
		storage.NewNamespace(storage.DefaultNamespace),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](
			storage.WithOrder(storage.OrderDesc), storage.WithLimit(1),
		),
	)
	res, err = s.store.ListFlags(context.TODO(), req)
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)

	assert.Equal(t, newest.Key, got[0].Key)

	// get last (oldest) flag
	req = storage.ListWithOptions(
		storage.NewNamespace(storage.DefaultNamespace),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](
			storage.WithOrder(storage.OrderDesc), storage.WithLimit(1), storage.WithOffset(2),
		),
	)
	res, err = s.store.ListFlags(context.TODO(), req)
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)

	assert.Equal(t, oldest.Key, got[0].Key)

	// get all flags
	req = storage.ListWithOptions(
		storage.NewNamespace(storage.DefaultNamespace),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](
			storage.WithOrder(storage.OrderDesc),
		),
	)
	res, err = s.store.ListFlags(context.TODO(), req)
	require.NoError(t, err)

	got = res.Results

	assert.Equal(t, newest.Key, got[0].Key)
	assert.Equal(t, middle.Key, got[1].Key)
	assert.Equal(t, oldest.Key, got[2].Key)
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
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
			Enabled:     true,
		},
	}

	for _, req := range reqs {
		if s.db.Driver == fliptsql.MySQL {
			// required for MySQL since it only s.stores timestamps to the second and not millisecond granularity
			time.Sleep(time.Second)
		}
		_, err := s.store.CreateFlag(context.TODO(), req)
		require.NoError(t, err)
	}

	oldest, middle, newest := reqs[0], reqs[1], reqs[2]

	// TODO: the ordering (DESC) is required because the default ordering is ASC and we are not clearing the DB between tests
	// get newest flag
	req := storage.ListWithOptions(
		storage.NewNamespace(storage.DefaultNamespace),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](storage.WithOrder(storage.OrderDesc), storage.WithLimit(1)),
	)
	res, err := s.store.ListFlags(context.TODO(), req)
	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, newest.Key, got[0].Key)
	assert.NotEmpty(t, res.NextPageToken)

	pageToken := &common.PageToken{}
	pTokenB, err := base64.StdEncoding.DecodeString(res.NextPageToken)
	require.NoError(t, err)

	err = json.Unmarshal(pTokenB, pageToken)
	require.NoError(t, err)
	// next page should be the middle flag
	assert.Equal(t, middle.Key, pageToken.Key)
	assert.NotZero(t, pageToken.Offset)

	req.QueryParams.PageToken = res.NextPageToken

	// get middle flag
	res, err = s.store.ListFlags(context.TODO(), req)
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, middle.Key, got[0].Key)

	pTokenB, err = base64.StdEncoding.DecodeString(res.NextPageToken)
	require.NoError(t, err)

	err = json.Unmarshal(pTokenB, pageToken)

	require.NoError(t, err)
	// next page should be the oldest flag
	assert.Equal(t, oldest.Key, pageToken.Key)
	assert.NotZero(t, pageToken.Offset)

	req.QueryParams.Limit = 1
	req.QueryParams.Order = storage.OrderDesc
	req.QueryParams.PageToken = res.NextPageToken

	// get oldest flag
	res, err = s.store.ListFlags(context.TODO(), req)
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, oldest.Key, got[0].Key)

	req = storage.ListWithOptions(
		storage.NewNamespace(storage.DefaultNamespace),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](
			storage.WithOrder(storage.OrderDesc), storage.WithLimit(3),
		),
	)
	// get all flags
	res, err = s.store.ListFlags(context.TODO(), req)
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 3)
	assert.Equal(t, newest.Key, got[0].Key)
	assert.Equal(t, middle.Key, got[1].Key)
	assert.Equal(t, oldest.Key, got[2].Key)
}

func (s *DBTestSuite) TestListFlagsPagination_FullWalk() {
	t := s.T()

	namespace := uuid.Must(uuid.NewV4()).String()

	ctx := context.Background()
	_, err := s.store.CreateNamespace(ctx, &flipt.CreateNamespaceRequest{
		Key: namespace,
	})
	require.NoError(t, err)

	var (
		totalFlags = 9
		pageSize   = uint64(3)
	)

	for i := 0; i < totalFlags; i++ {
		req := flipt.CreateFlagRequest{
			NamespaceKey: namespace,
			Key:          fmt.Sprintf("flag_%03d", i),
			Name:         "foo",
			Description:  "bar",
		}

		_, err := s.store.CreateFlag(ctx, &req)
		require.NoError(t, err)

		for i := 0; i < 2; i++ {
			if i > 0 && s.db.Driver == fliptsql.MySQL {
				// required for MySQL since it only s.stores timestamps to the second and not millisecond granularity
				time.Sleep(time.Second)
			}

			_, err := s.store.CreateVariant(ctx, &flipt.CreateVariantRequest{
				NamespaceKey: namespace,
				FlagKey:      req.Key,
				Key:          fmt.Sprintf("variant_%d", i),
			})
			require.NoError(t, err)
		}
	}

	req := storage.ListWithOptions(
		storage.NewNamespace(namespace),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](
			storage.WithLimit(pageSize),
		),
	)
	resp, err := s.store.ListFlags(ctx, req)
	require.NoError(t, err)

	found := resp.Results
	for token := resp.NextPageToken; token != ""; token = resp.NextPageToken {
		req.QueryParams.PageToken = token
		resp, err = s.store.ListFlags(ctx, req)
		require.NoError(t, err)

		found = append(found, resp.Results...)
	}

	require.Len(t, found, totalFlags)

	for i := 0; i < totalFlags; i++ {
		assert.Equal(t, namespace, found[i].NamespaceKey)

		expectedFlag := fmt.Sprintf("flag_%03d", i)
		assert.Equal(t, expectedFlag, found[i].Key)
		assert.Equal(t, "foo", found[i].Name)
		assert.Equal(t, "bar", found[i].Description)

		require.Len(t, found[i].Variants, 2)
		assert.Equal(t, namespace, found[i].Variants[0].NamespaceKey)
		assert.Equal(t, expectedFlag, found[i].Variants[0].FlagKey)
		assert.Equal(t, "variant_0", found[i].Variants[0].Key)

		assert.Equal(t, namespace, found[i].Variants[1].NamespaceKey)
		assert.Equal(t, expectedFlag, found[i].Variants[1].FlagKey)
		assert.Equal(t, "variant_1", found[i].Variants[1].Key)
	}
}

func (s *DBTestSuite) TestCreateFlag() {
	t := s.T()

	metadataMap := map[string]any{
		"key": "value",
	}

	metadata, _ := structpb.NewStruct(metadataMap)

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
		Metadata:    metadata,
	})

	require.NoError(t, err)

	assert.Equal(t, storage.DefaultNamespace, flag.NamespaceKey)
	assert.Equal(t, t.Name(), flag.Key)
	assert.Equal(t, "foo", flag.Name)
	assert.Equal(t, "bar", flag.Description)
	assert.True(t, flag.Enabled)
	assert.Equal(t, metadata.String(), flag.Metadata.String())
	assert.NotZero(t, flag.CreatedAt)
	assert.Equal(t, flag.CreatedAt.Seconds, flag.UpdatedAt.Seconds)
}

func (s *DBTestSuite) TestCreateFlagNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)

	assert.Equal(t, s.namespace, flag.NamespaceKey)
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

	assert.EqualError(t, err, "flag \"default/TestDBTestSuite/TestCreateFlag_DuplicateKey\" is not unique")
}

func (s *DBTestSuite) TestCreateFlagNamespace_DuplicateKey() {
	t := s.T()

	_, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)

	_, err = s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	assert.EqualError(t, err, fmt.Sprintf("flag \"%s/%s\" is not unique", s.namespace, t.Name()))
}

func (s *DBTestSuite) TestUpdateFlag() {
	t := s.T()

	metadataMap := map[string]any{
		"key": "value",
	}

	metadata, _ := structpb.NewStruct(metadataMap)
	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
		Metadata:    metadata,
	})

	require.NoError(t, err)

	assert.Equal(t, storage.DefaultNamespace, flag.NamespaceKey)
	assert.Equal(t, t.Name(), flag.Key)
	assert.Equal(t, "foo", flag.Name)
	assert.Equal(t, "bar", flag.Description)
	assert.True(t, flag.Enabled)
	assert.Equal(t, metadata.String(), flag.Metadata.String())
	assert.NotZero(t, flag.CreatedAt)
	assert.Equal(t, flag.CreatedAt.Seconds, flag.UpdatedAt.Seconds)

	updatedMetadataMap := map[string]any{
		"key": "value",
		"foo": "bar",
	}

	updatedMetadata, _ := structpb.NewStruct(updatedMetadataMap)
	updated, err := s.store.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{
		Key:         flag.Key,
		Name:        flag.Name,
		Description: "foobar",
		Enabled:     true,
		Metadata:    updatedMetadata,
	})

	require.NoError(t, err)

	assert.Equal(t, storage.DefaultNamespace, updated.NamespaceKey)
	assert.Equal(t, flag.Key, updated.Key)
	assert.Equal(t, flag.Name, updated.Name)
	assert.Equal(t, "foobar", updated.Description)
	assert.True(t, flag.Enabled)
	assert.Equal(t, updatedMetadata.String(), updated.Metadata.String())
	assert.NotZero(t, updated.CreatedAt)
	assert.NotZero(t, updated.UpdatedAt)
}

func (s *DBTestSuite) TestUpdateFlagNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)

	assert.Equal(t, s.namespace, flag.NamespaceKey)
	assert.Equal(t, t.Name(), flag.Key)
	assert.Equal(t, "foo", flag.Name)
	assert.Equal(t, "bar", flag.Description)
	assert.True(t, flag.Enabled)
	assert.NotZero(t, flag.CreatedAt)
	assert.Equal(t, flag.CreatedAt.Seconds, flag.UpdatedAt.Seconds)

	updated, err := s.store.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          flag.Key,
		Name:         flag.Name,
		Description:  "foobar",
		Enabled:      true,
	})

	require.NoError(t, err)

	assert.Equal(t, s.namespace, updated.NamespaceKey)
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

	assert.EqualError(t, err, "flag \"default/foo\" not found")
}

func (s *DBTestSuite) TestUpdateFlagNamespace_NotFound() {
	t := s.T()

	_, err := s.store.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          "foo",
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	assert.EqualError(t, err, fmt.Sprintf("flag \"%s/foo\" not found", s.namespace))
}

func (s *DBTestSuite) TestUpdateFlag_DefaultVariant() {
	t := s.T()

	t.Run("update flag with default variant", func(t *testing.T) {
		flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
			Key:         t.Name(),
			Name:        "foo",
			Description: "bar",
			Enabled:     true,
		})

		require.NoError(t, err)

		variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
			FlagKey:     flag.Key,
			Key:         t.Name(),
			Name:        "foo",
			Description: "bar",
			Attachment:  `{"key":"value"}`,
		})

		require.NoError(t, err)
		assert.NotNil(t, variant)

		_, err = s.store.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{
			Key:              flag.Key,
			Name:             flag.Name,
			Description:      "foobar",
			Enabled:          true,
			DefaultVariantId: variant.Id,
		})

		require.NoError(t, err)

		// get the flag again
		flag, err = s.store.GetFlag(context.TODO(), storage.NewResource(storage.DefaultNamespace, flag.Key))

		require.NoError(t, err)
		assert.NotNil(t, flag)
		assert.Equal(t, variant.Id, flag.DefaultVariant.Id)
		assert.Equal(t, variant.Key, flag.DefaultVariant.Key)
		assert.Equal(t, variant.Name, flag.DefaultVariant.Name)
		assert.Equal(t, variant.Description, flag.DefaultVariant.Description)
		assert.Equal(t, variant.Attachment, flag.DefaultVariant.Attachment)
	})

	t.Run("update flag with default variant not found", func(t *testing.T) {
		flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
			Key:         t.Name(),
			Name:        "foo",
			Description: "bar",
			Enabled:     true,
		})

		require.NoError(t, err)

		_, err = s.store.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{
			Key:              flag.Key,
			Name:             flag.Name,
			Description:      "foobar",
			Enabled:          true,
			DefaultVariantId: "non-existent",
		})

		assert.EqualError(t, err, "variant \"non-existent\" not found for flag \"default/TestDBTestSuite/TestUpdateFlag_DefaultVariant/update_flag_with_default_variant_not_found\"")
	})

	t.Run("update flag with variant from different flag", func(t *testing.T) {
		flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
			Key:         t.Name(),
			Name:        "foo",
			Description: "bar",
			Enabled:     true,
		})

		require.NoError(t, err)

		flag2, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
			Key:         fmt.Sprintf("%s_two", t.Name()),
			Name:        "foo",
			Description: "bar",
			Enabled:     true,
		})

		require.NoError(t, err)

		variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
			FlagKey:     flag2.Key,
			Key:         t.Name(),
			Name:        "foo",
			Description: "bar",
			Attachment:  `{"key":"value"}`,
		})

		require.NoError(t, err)

		_, err = s.store.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{
			Key:              flag.Key,
			Name:             flag.Name,
			Description:      "foobar",
			Enabled:          true,
			DefaultVariantId: variant.Id,
		})

		assert.EqualError(t, err, fmt.Sprintf("variant \"%s\" not found for flag \"%s/%s\"", variant.Id, "default", flag.Key))
	})

	t.Run("update flag with default variant in different namespace", func(t *testing.T) {

		flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
			NamespaceKey: s.namespace,
			Key:          t.Name(),
			Name:         "foo",
			Description:  "bar",
			Enabled:      true,
		})

		require.NoError(t, err)

		variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
			NamespaceKey: s.namespace,
			FlagKey:      flag.Key,
			Key:          t.Name(),
			Name:         "foo",
			Description:  "bar",
			Attachment:   `{"key":"value"}`,
		})

		require.NoError(t, err)

		// flag in default namespace
		flag2, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
			Key:         t.Name(),
			Name:        "foo",
			Description: "bar",
			Enabled:     true,
		})

		require.NoError(t, err)

		// try to update flag in non-default namespace with default variant from default namespace
		_, err = s.store.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{
			Key:              flag2.Key,
			Name:             flag2.Name,
			Description:      flag2.Description,
			Enabled:          true,
			DefaultVariantId: variant.Id,
		})

		assert.EqualError(t, err, fmt.Sprintf("variant \"%s\" not found for flag \"%s/%s\"", variant.Id, "default", flag2.Key))
	})
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

func (s *DBTestSuite) TestDeleteFlagNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	err = s.store.DeleteFlag(context.TODO(), &flipt.DeleteFlagRequest{
		NamespaceKey: s.namespace,
		Key:          flag.Key,
	})

	require.NoError(t, err)
}

func (s *DBTestSuite) TestDeleteFlag_NotFound() {
	t := s.T()

	err := s.store.DeleteFlag(context.TODO(), &flipt.DeleteFlagRequest{Key: "foo"})
	require.NoError(t, err)
}

func (s *DBTestSuite) TestDeleteFlagNamespace_NotFound() {
	t := s.T()

	err := s.store.DeleteFlag(context.TODO(), &flipt.DeleteFlagRequest{
		NamespaceKey: s.namespace,
		Key:          "foo",
	})

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
	assert.Equal(t, storage.DefaultNamespace, variant.NamespaceKey)
	assert.Equal(t, flag.Key, variant.FlagKey)
	assert.Equal(t, t.Name(), variant.Key)
	assert.Equal(t, "foo", variant.Name)
	assert.Equal(t, "bar", variant.Description)
	assert.Equal(t, attachment, variant.Attachment)
	assert.NotZero(t, variant.CreatedAt)
	assert.Equal(t, variant.CreatedAt.Seconds, variant.UpdatedAt.Seconds)

	// get the flag again
	flag, err = s.store.GetFlag(context.TODO(), storage.NewResource(storage.DefaultNamespace, flag.Key))

	require.NoError(t, err)
	assert.NotNil(t, flag)

	assert.Len(t, flag.Variants, 1)
}

func (s *DBTestSuite) TestCreateVariantNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	attachment := `{"key":"value"}`
	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Attachment:   attachment,
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	assert.NotZero(t, variant.Id)
	assert.Equal(t, s.namespace, variant.NamespaceKey)
	assert.Equal(t, flag.Key, variant.FlagKey)
	assert.Equal(t, t.Name(), variant.Key)
	assert.Equal(t, "foo", variant.Name)
	assert.Equal(t, "bar", variant.Description)
	assert.Equal(t, attachment, variant.Attachment)
	assert.NotZero(t, variant.CreatedAt)
	assert.Equal(t, variant.CreatedAt.Seconds, variant.UpdatedAt.Seconds)

	// get the flag again
	flag, err = s.store.GetFlag(context.TODO(), storage.NewResource(s.namespace, flag.Key))

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

	assert.EqualError(t, err, "flag \"default/foo\" not found")
}

func (s *DBTestSuite) TestCreateVariantNamespace_FlagNotFound() {
	t := s.T()

	_, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      "foo",
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
	})

	assert.EqualError(t, err, fmt.Sprintf("flag \"%s/foo\" not found", s.namespace))
}

func (s *DBTestSuite) TestCreateVariant_DuplicateKey() {
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

	assert.EqualError(t, err, "variant \"foo\" is not unique for flag \"default/TestDBTestSuite/TestCreateVariant_DuplicateKey\"")
}

func (s *DBTestSuite) TestCreateVariantNamespace_DuplicateKey() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          "foo",
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	// try to create another variant with the same name for this flag
	_, err = s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          "foo",
		Name:         "foo",
		Description:  "bar",
	})

	assert.EqualError(t, err, fmt.Sprintf("variant \"foo\" is not unique for flag \"%s/%s\"", s.namespace, t.Name()))
}

func (s *DBTestSuite) TestCreateVariant_DuplicateKey_DifferentFlag() {
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

func (s *DBTestSuite) TestCreateVariantNamespace_DuplicateFlag_DuplicateKey() {
	t := s.T()

	flag1, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag1)

	variant1, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag1.Key,
		Key:          "foo",
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant1)

	assert.NotZero(t, variant1.Id)
	assert.Equal(t, s.namespace, variant1.NamespaceKey)
	assert.Equal(t, flag1.Key, variant1.FlagKey)
	assert.Equal(t, "foo", variant1.Key)

	flag2, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
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
	assert.Equal(t, storage.DefaultNamespace, variant2.NamespaceKey)
	assert.Equal(t, flag2.Key, variant2.FlagKey)
	assert.Equal(t, "foo", variant2.Key)
}

func (s *DBTestSuite) TestGetFlagWithVariantsMultiNamespace() {
	t := s.T()

	for _, namespace := range []string{storage.DefaultNamespace, s.namespace} {
		flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
			NamespaceKey: namespace,
			Key:          t.Name(),
			Name:         "foo",
			Description:  "bar",
			Enabled:      true,
		})

		require.NoError(t, err)
		assert.NotNil(t, flag)

		attachment := `{"key":"value"}`
		variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
			NamespaceKey: namespace,
			FlagKey:      flag.Key,
			Key:          t.Name(),
			Name:         "foo",
			Description:  "bar",
			Attachment:   attachment,
		})

		require.NoError(t, err)
		assert.NotNil(t, variant)

		assert.NotZero(t, variant.Id)
		assert.Equal(t, namespace, variant.NamespaceKey)
		assert.Equal(t, flag.Key, variant.FlagKey)
		assert.Equal(t, t.Name(), variant.Key)
		assert.Equal(t, "foo", variant.Name)
		assert.Equal(t, "bar", variant.Description)
		assert.Equal(t, attachment, variant.Attachment)
		assert.NotZero(t, variant.CreatedAt)
		assert.Equal(t, variant.CreatedAt.Seconds, variant.UpdatedAt.Seconds)
	}

	// get the default namespaced flag
	flag, err := s.store.GetFlag(context.TODO(), storage.NewResource(storage.DefaultNamespace, t.Name()))

	require.NoError(t, err)
	assert.NotNil(t, flag)

	assert.Len(t, flag.Variants, 1)

	variant := flag.Variants[0]
	assert.NotZero(t, variant.Id)
	assert.Equal(t, storage.DefaultNamespace, variant.NamespaceKey)
	assert.Equal(t, flag.Key, variant.FlagKey)
	assert.Equal(t, t.Name(), variant.Key)
	assert.Equal(t, "foo", variant.Name)
	assert.Equal(t, "bar", variant.Description)
	assert.Equal(t, `{"key":"value"}`, variant.Attachment)
	assert.NotZero(t, variant.CreatedAt)
	assert.Equal(t, variant.CreatedAt.Seconds, variant.UpdatedAt.Seconds)
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
	assert.Equal(t, storage.DefaultNamespace, variant.NamespaceKey)
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
	assert.Equal(t, storage.DefaultNamespace, updated.NamespaceKey)
	assert.Equal(t, variant.FlagKey, updated.FlagKey)
	assert.Equal(t, variant.Key, updated.Key)
	assert.Equal(t, variant.Name, updated.Name)
	assert.Equal(t, "foobar", updated.Description)
	assert.Equal(t, `{"key":"value2"}`, updated.Attachment)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotZero(t, updated.UpdatedAt)

	// get the flag again
	flag, err = s.store.GetFlag(context.TODO(), storage.NewResource(storage.DefaultNamespace, flag.Key))

	require.NoError(t, err)
	assert.NotNil(t, flag)

	assert.Len(t, flag.Variants, 1)
}

func (s *DBTestSuite) TestUpdateVariantNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	attachment1 := `{"key":"value1"}`
	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          "foo",
		Name:         "foo",
		Description:  "bar",
		Attachment:   attachment1,
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	assert.NotZero(t, variant.Id)
	assert.Equal(t, s.namespace, variant.NamespaceKey)
	assert.Equal(t, flag.Key, variant.FlagKey)
	assert.Equal(t, "foo", variant.Key)
	assert.Equal(t, "foo", variant.Name)
	assert.Equal(t, "bar", variant.Description)
	assert.Equal(t, attachment1, variant.Attachment)
	assert.NotZero(t, variant.CreatedAt)
	assert.Equal(t, variant.CreatedAt.Seconds, variant.UpdatedAt.Seconds)

	updated, err := s.store.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{
		NamespaceKey: s.namespace,
		Id:           variant.Id,
		FlagKey:      variant.FlagKey,
		Key:          variant.Key,
		Name:         variant.Name,
		Description:  "foobar",
		Attachment:   `{"key":      "value2"}`,
	})

	require.NoError(t, err)

	assert.Equal(t, variant.Id, updated.Id)
	assert.Equal(t, s.namespace, updated.NamespaceKey)
	assert.Equal(t, variant.FlagKey, updated.FlagKey)
	assert.Equal(t, variant.Key, updated.Key)
	assert.Equal(t, variant.Name, updated.Name)
	assert.Equal(t, "foobar", updated.Description)
	assert.Equal(t, `{"key":"value2"}`, updated.Attachment)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotZero(t, updated.UpdatedAt)

	// get the flag again
	flag, err = s.store.GetFlag(context.TODO(), storage.NewResource(s.namespace, flag.Key))

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

func (s *DBTestSuite) TestUpdateVariantNamespace_NotFound() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	_, err = s.store.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{
		NamespaceKey: s.namespace,
		Id:           "foo",
		FlagKey:      flag.Key,
		Key:          "foo",
		Name:         "foo",
		Description:  "bar",
	})

	assert.EqualError(t, err, "variant \"foo\" not found")
}

func (s *DBTestSuite) TestUpdateVariant_DuplicateKey() {
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

	assert.EqualError(t, err, "variant \"foo\" is not unique for flag \"default/TestDBTestSuite/TestUpdateVariant_DuplicateKey\"")
}

func (s *DBTestSuite) TestUpdateVariantNamespace_DuplicateKey() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant1, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          "foo",
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant1)

	variant2, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          "bar",
		Name:         "bar",
		Description:  "baz",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant2)

	_, err = s.store.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{
		NamespaceKey: s.namespace,
		Id:           variant2.Id,
		FlagKey:      variant2.FlagKey,
		Key:          variant1.Key,
		Name:         variant2.Name,
		Description:  "foobar",
	})

	assert.EqualError(t, err, fmt.Sprintf("variant \"foo\" is not unique for flag \"%s/%s\"", s.namespace, t.Name()))
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
	flag, err = s.store.GetFlag(context.TODO(), storage.NewResource(storage.DefaultNamespace, flag.Key))

	require.NoError(t, err)
	assert.NotNil(t, flag)

	assert.Empty(t, flag.Variants)
}

func (s *DBTestSuite) TestDeleteVariantNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	variant, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          "foo",
		Name:         "foo",
		Description:  "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, variant)

	err = s.store.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      variant.FlagKey,
		Id:           variant.Id,
	})
	require.NoError(t, err)

	// get the flag again
	flag, err = s.store.GetFlag(context.TODO(), storage.NewResource(s.namespace, flag.Key))

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

	require.EqualError(t, err, "atleast one rule exists that includes this variant")

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

func (s *DBTestSuite) TestDeleteVariantNamespace_NotFound() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)
	assert.NotNil(t, flag)

	err = s.store.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{
		NamespaceKey: s.namespace,
		Id:           "foo",
		FlagKey:      flag.Key,
	})

	require.NoError(t, err)
}

func (s *DBTestSuite) TestDeleteVariant_DefaultVariant() {
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

	_, err = s.store.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{
		Key:              flag.Key,
		Name:             flag.Name,
		Description:      flag.Description,
		Enabled:          true,
		DefaultVariantId: variant.Id,
	})

	require.NoError(t, err)

	err = s.store.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{FlagKey: variant.FlagKey, Id: variant.Id})
	require.NoError(t, err)

	// get the flag again
	flag, err = s.store.GetFlag(context.TODO(), storage.NewResource(storage.DefaultNamespace, flag.Key))

	require.NoError(t, err)
	assert.NotNil(t, flag)

	assert.Empty(t, flag.Variants)
	assert.Nil(t, flag.DefaultVariant)
}

func BenchmarkListFlags(b *testing.B) {
	s := new(DBTestSuite)
	t := &testing.T{}
	s.SetT(t)
	s.SetupSuite()

	for i := 0; i < 1000; i++ {
		reqs := []*flipt.CreateFlagRequest{
			{
				Key:     uuid.Must(uuid.NewV4()).String(),
				Name:    fmt.Sprintf("foo_%d", i),
				Enabled: true,
			},
		}

		for _, req := range reqs {
			f, err := s.store.CreateFlag(context.TODO(), req)
			require.NoError(t, err)
			assert.NotNil(t, f)

			for j := 0; j < 10; j++ {
				v, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
					FlagKey: f.Key,
					Key:     uuid.Must(uuid.NewV4()).String(),
					Name:    fmt.Sprintf("variant_%d", j),
				})

				require.NoError(t, err)
				assert.NotNil(t, v)
			}
		}
	}

	b.ResetTimer()

	req := storage.ListWithOptions(storage.NewNamespace(storage.DefaultNamespace))
	b.Run("no-pagination", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			flags, err := s.store.ListFlags(context.TODO(), req)
			require.NoError(t, err)
			assert.NotEmpty(t, flags)
		}
	})

	for _, pageSize := range []uint64{10, 25, 100, 500} {
		req := req
		req.QueryParams.Limit = pageSize
		b.Run(fmt.Sprintf("pagination-limit-%d", pageSize), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				flags, err := s.store.ListFlags(context.TODO(), req)
				require.NoError(t, err)
				assert.NotEmpty(t, flags)
			}
		})
	}

	b.Run("pagination", func(b *testing.B) {
		req := req
		req.QueryParams.Limit = 500
		req.QueryParams.Offset = 50
		req.QueryParams.Order = storage.OrderDesc
		for i := 0; i < b.N; i++ {
			flags, err := s.store.ListFlags(context.TODO(), req)
			require.NoError(t, err)
			assert.NotEmpty(t, flags)
		}
	})

	s.TearDownSuite()
}
