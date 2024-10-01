package sql_test

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/sql/common"
	flipt "go.flipt.io/flipt/rpc/flipt"
)

func (s *DBTestSuite) TestGetNamespace() {
	t := s.T()

	ns, err := s.store.CreateNamespace(context.TODO(), &flipt.CreateNamespaceRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, ns)

	got, err := s.store.GetNamespace(context.TODO(), storage.NewNamespace(ns.Key))

	require.NoError(t, err)
	assert.NotNil(t, got)

	assert.Equal(t, ns.Key, got.Key)
	assert.Equal(t, ns.Name, got.Name)
	assert.Equal(t, ns.Description, got.Description)
	assert.NotZero(t, ns.CreatedAt)
	assert.NotZero(t, ns.UpdatedAt)
}

func (s *DBTestSuite) TestGetNamespaceNotFound() {
	t := s.T()

	_, err := s.store.GetNamespace(context.TODO(), storage.NewNamespace("foo"))
	assert.EqualError(t, err, "namespace \"foo\" not found")
}

func (s *DBTestSuite) TestListNamespaces() {
	t := s.T()

	reqs := []*flipt.CreateNamespaceRequest{
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
		},
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
		},
	}

	for _, req := range reqs {
		_, err := s.store.CreateNamespace(context.TODO(), req)
		require.NoError(t, err)
	}

	_, err := s.store.ListNamespaces(context.TODO(), storage.ListWithOptions(
		storage.ReferenceRequest{},
		storage.ListWithQueryParamOptions[storage.ReferenceRequest](
			storage.WithPageToken("Hello World"),
		),
	))
	require.EqualError(t, err, "pageToken is not valid: \"Hello World\"")

	res, err := s.store.ListNamespaces(context.TODO(), storage.ListWithOptions(
		storage.ReferenceRequest{},
	))
	require.NoError(t, err)

	got := res.Results
	assert.NotEmpty(t, got)

	for _, ns := range got {
		assert.NotZero(t, ns.CreatedAt)
		assert.NotZero(t, ns.UpdatedAt)
	}
}

func (s *DBTestSuite) TestListNamespacesPagination_LimitOffset() {
	t := s.T()

	reqs := []*flipt.CreateNamespaceRequest{
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
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
		},
	}

	for _, req := range reqs {
		_, err := s.store.CreateNamespace(context.TODO(), req)
		require.NoError(t, err)
	}

	oldest, middle, newest := reqs[0], reqs[1], reqs[2]

	// TODO: the ordering (DESC) is required because the default ordering is ASC and we are not clearing the DB between tests
	// get middle namespace
	res, err := s.store.ListNamespaces(context.TODO(),
		storage.ListWithOptions(
			storage.ReferenceRequest{},
			storage.ListWithQueryParamOptions[storage.ReferenceRequest](
				storage.WithOrder(storage.OrderDesc), storage.WithLimit(1), storage.WithOffset(1)),
		),
	)

	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)

	assert.Equal(t, middle.Key, got[0].Key)

	// get first (newest) namespace
	res, err = s.store.ListNamespaces(context.TODO(),
		storage.ListWithOptions(
			storage.ReferenceRequest{},
			storage.ListWithQueryParamOptions[storage.ReferenceRequest](
				storage.WithOrder(storage.OrderDesc), storage.WithLimit(1)),
		),
	)

	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)

	assert.Equal(t, newest.Key, got[0].Key)

	// get last (oldest) namespace
	res, err = s.store.ListNamespaces(context.TODO(),
		storage.ListWithOptions(
			storage.ReferenceRequest{},
			storage.ListWithQueryParamOptions[storage.ReferenceRequest](
				storage.WithOrder(storage.OrderDesc), storage.WithLimit(1), storage.WithOffset(2)),
		),
	)
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)

	assert.Equal(t, oldest.Key, got[0].Key)

	// get all namespaces
	res, err = s.store.ListNamespaces(context.TODO(),
		storage.ListWithOptions(
			storage.ReferenceRequest{},
			storage.ListWithQueryParamOptions[storage.ReferenceRequest](
				storage.WithOrder(storage.OrderDesc),
			),
		),
	)
	require.NoError(t, err)

	got = res.Results

	assert.Equal(t, newest.Key, got[0].Key)
	assert.Equal(t, middle.Key, got[1].Key)
	assert.Equal(t, oldest.Key, got[2].Key)
}

func (s *DBTestSuite) TestListNamespacesPagination_LimitWithNextPage() {
	t := s.T()

	reqs := []*flipt.CreateNamespaceRequest{
		{
			Key:         uuid.Must(uuid.NewV4()).String(),
			Name:        "foo",
			Description: "bar",
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
		},
	}

	for _, req := range reqs {
		_, err := s.store.CreateNamespace(context.TODO(), req)
		require.NoError(t, err)
	}

	oldest, middle, newest := reqs[0], reqs[1], reqs[2]

	// TODO: the ordering (DESC) is required because the default ordering is ASC and we are not clearing the DB between tests
	// get newest namespace
	opts := []storage.QueryOption{storage.WithOrder(storage.OrderDesc), storage.WithLimit(1)}

	res, err := s.store.ListNamespaces(context.TODO(), storage.ListWithOptions(
		storage.ReferenceRequest{},
		storage.ListWithQueryParamOptions[storage.ReferenceRequest](opts...),
	))
	require.NoError(t, err)

	got := res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, newest.Key, got[0].Key)
	assert.NotEmpty(t, res.NextPageToken)

	pTokenB, err := base64.StdEncoding.DecodeString(res.NextPageToken)
	require.NoError(t, err)

	pageToken := &common.PageToken{}
	err = json.Unmarshal(pTokenB, pageToken)
	require.NoError(t, err)
	// next page should be the middle namespace
	assert.Equal(t, middle.Key, pageToken.Key)
	assert.NotZero(t, pageToken.Offset)

	opts = append(opts, storage.WithPageToken(res.NextPageToken))

	// get middle namespace
	res, err = s.store.ListNamespaces(context.TODO(), storage.ListWithOptions(
		storage.ReferenceRequest{},
		storage.ListWithQueryParamOptions[storage.ReferenceRequest](opts...),
	))
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, middle.Key, got[0].Key)

	pTokenB, err = base64.StdEncoding.DecodeString(res.NextPageToken)
	require.NoError(t, err)

	err = json.Unmarshal(pTokenB, pageToken)
	require.NoError(t, err)
	// next page should be the oldest namespace
	assert.Equal(t, oldest.Key, pageToken.Key)
	assert.NotZero(t, pageToken.Offset)

	opts = []storage.QueryOption{storage.WithOrder(storage.OrderDesc), storage.WithLimit(1), storage.WithPageToken(res.NextPageToken)}

	// get oldest namespace
	res, err = s.store.ListNamespaces(context.TODO(), storage.ListWithOptions(
		storage.ReferenceRequest{},
		storage.ListWithQueryParamOptions[storage.ReferenceRequest](opts...),
	))
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 1)
	assert.Equal(t, oldest.Key, got[0].Key)

	opts = []storage.QueryOption{storage.WithOrder(storage.OrderDesc), storage.WithLimit(3)}
	// get all namespaces
	res, err = s.store.ListNamespaces(context.TODO(), storage.ListWithOptions(
		storage.ReferenceRequest{},
		storage.ListWithQueryParamOptions[storage.ReferenceRequest](opts...),
	))
	require.NoError(t, err)

	got = res.Results
	assert.Len(t, got, 3)
	assert.Equal(t, newest.Key, got[0].Key)
	assert.Equal(t, middle.Key, got[1].Key)
	assert.Equal(t, oldest.Key, got[2].Key)
}

func (s *DBTestSuite) TestCreateNamespace() {
	t := s.T()

	ns, err := s.store.CreateNamespace(context.TODO(), &flipt.CreateNamespaceRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)

	assert.Equal(t, t.Name(), ns.Key)
	assert.Equal(t, "foo", ns.Name)
	assert.Equal(t, "bar", ns.Description)
	assert.NotZero(t, ns.CreatedAt)
	assert.Equal(t, ns.CreatedAt.Seconds, ns.UpdatedAt.Seconds)
}

func (s *DBTestSuite) TestCreateNamespace_DuplicateKey() {
	t := s.T()

	_, err := s.store.CreateNamespace(context.TODO(), &flipt.CreateNamespaceRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)

	_, err = s.store.CreateNamespace(context.TODO(), &flipt.CreateNamespaceRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "namespace \"TestDBTestSuite/TestCreateNamespace_DuplicateKey\" is not unique")
}

func (s *DBTestSuite) TestUpdateNamespace() {
	t := s.T()

	ns, err := s.store.CreateNamespace(context.TODO(), &flipt.CreateNamespaceRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)

	assert.Equal(t, t.Name(), ns.Key)
	assert.Equal(t, "foo", ns.Name)
	assert.Equal(t, "bar", ns.Description)
	assert.NotZero(t, ns.CreatedAt)
	assert.Equal(t, ns.CreatedAt.Seconds, ns.UpdatedAt.Seconds)

	updated, err := s.store.UpdateNamespace(context.TODO(), &flipt.UpdateNamespaceRequest{
		Key:         ns.Key,
		Name:        ns.Name,
		Description: "foobar",
	})

	require.NoError(t, err)

	assert.Equal(t, ns.Key, updated.Key)
	assert.Equal(t, ns.Name, updated.Name)
	assert.Equal(t, "foobar", updated.Description)
	assert.NotZero(t, updated.CreatedAt)
	assert.NotZero(t, updated.UpdatedAt)
}

func (s *DBTestSuite) TestUpdateNamespace_NotFound() {
	t := s.T()

	_, err := s.store.UpdateNamespace(context.TODO(), &flipt.UpdateNamespaceRequest{
		Key:         "foo",
		Name:        "foo",
		Description: "bar",
	})

	assert.EqualError(t, err, "namespace \"foo\" not found")
}

func (s *DBTestSuite) TestDeleteNamespace() {
	t := s.T()

	ns, err := s.store.CreateNamespace(context.TODO(), &flipt.CreateNamespaceRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
	})

	require.NoError(t, err)
	assert.NotNil(t, ns)

	err = s.store.DeleteNamespace(context.TODO(), &flipt.DeleteNamespaceRequest{Key: ns.Key})
	require.NoError(t, err)
}

func (s *DBTestSuite) TestDeleteNamespace_NotFound() {
	t := s.T()

	err := s.store.DeleteNamespace(context.TODO(), &flipt.DeleteNamespaceRequest{Key: "foo"})
	require.NoError(t, err)
}
