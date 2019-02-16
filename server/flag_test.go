package server

import (
	"context"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/markphelps/flipt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ flipt.FlagService = &flagServiceMock{}

type flagServiceMock struct {
	flagFn          func(context.Context, *flipt.GetFlagRequest) (*flipt.Flag, error)
	flagsFn         func(context.Context, *flipt.ListFlagRequest) ([]*flipt.Flag, error)
	createFlagFn    func(context.Context, *flipt.CreateFlagRequest) (*flipt.Flag, error)
	updateFlagFn    func(context.Context, *flipt.UpdateFlagRequest) (*flipt.Flag, error)
	deleteFlagFn    func(context.Context, *flipt.DeleteFlagRequest) error
	createVariantFn func(context.Context, *flipt.CreateVariantRequest) (*flipt.Variant, error)
	updateVariantFn func(context.Context, *flipt.UpdateVariantRequest) (*flipt.Variant, error)
	deleteVariantFn func(context.Context, *flipt.DeleteVariantRequest) error
}

func (srv *flagServiceMock) Flag(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
	return srv.flagFn(ctx, r)
}

func (srv *flagServiceMock) Flags(ctx context.Context, r *flipt.ListFlagRequest) ([]*flipt.Flag, error) {
	return srv.flagsFn(ctx, r)
}

func (srv *flagServiceMock) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	return srv.createFlagFn(ctx, r)
}

func (srv *flagServiceMock) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	return srv.updateFlagFn(ctx, r)
}

func (srv *flagServiceMock) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	return srv.deleteFlagFn(ctx, r)
}

func (srv *flagServiceMock) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	return srv.createVariantFn(ctx, r)
}

func (srv *flagServiceMock) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	return srv.updateVariantFn(ctx, r)
}

func (srv *flagServiceMock) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	return srv.deleteVariantFn(ctx, r)
}

func TestGetFlag(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.GetFlagRequest
		f    func(context.Context, *flipt.GetFlagRequest) (*flipt.Flag, error)
	}{
		{
			name: "ok",
			req:  &flipt.GetFlagRequest{Key: "key"},
			f: func(_ context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)

				return &flipt.Flag{
					Key: "key",
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagService: &flagServiceMock{
					flagFn: tt.f,
				},
			}

			flag, err := s.GetFlag(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.NotNil(t, flag)
		})
	}
}

func TestListFlags(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.ListFlagRequest
		f    func(context.Context, *flipt.ListFlagRequest) ([]*flipt.Flag, error)
	}{
		{
			name: "ok",
			req:  &flipt.ListFlagRequest{},
			f: func(context.Context, *flipt.ListFlagRequest) ([]*flipt.Flag, error) {
				return []*flipt.Flag{
					{Key: "flag"},
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagService: &flagServiceMock{
					flagsFn: tt.f,
				},
			}

			resp, err := s.ListFlags(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

func TestCreateFlag(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.CreateFlagRequest
		f    func(context.Context, *flipt.CreateFlagRequest) (*flipt.Flag, error)
	}{
		{
			name: "ok",
			req: &flipt.CreateFlagRequest{
				Key:         "key",
				Name:        "name",
				Description: "desc",
				Enabled:     true,
			},
			f: func(_ context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)
				assert.True(t, r.Enabled)

				return &flipt.Flag{
					Key:         r.Key,
					Name:        r.Name,
					Description: r.Description,
					Enabled:     r.Enabled,
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagService: &flagServiceMock{
					createFlagFn: tt.f,
				},
			}

			flag, err := s.CreateFlag(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.NotNil(t, flag)
		})
	}
}

func TestUpdateFlag(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.UpdateFlagRequest
		f    func(context.Context, *flipt.UpdateFlagRequest) (*flipt.Flag, error)
	}{
		{
			name: "ok",
			req: &flipt.UpdateFlagRequest{
				Key:         "key",
				Name:        "name",
				Description: "desc",
				Enabled:     true,
			},
			f: func(_ context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)
				assert.True(t, r.Enabled)

				return &flipt.Flag{
					Key:         r.Key,
					Name:        r.Name,
					Description: r.Description,
					Enabled:     r.Enabled,
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagService: &flagServiceMock{
					updateFlagFn: tt.f,
				},
			}

			flag, err := s.UpdateFlag(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.NotNil(t, flag)
		})
	}
}

func TestDeleteFlag(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.DeleteFlagRequest
		f    func(context.Context, *flipt.DeleteFlagRequest) error
	}{
		{
			name: "ok",
			req:  &flipt.DeleteFlagRequest{Key: "key"},
			f: func(_ context.Context, r *flipt.DeleteFlagRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagService: &flagServiceMock{
					deleteFlagFn: tt.f,
				},
			}

			resp, err := s.DeleteFlag(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.Equal(t, &empty.Empty{}, resp)
		})
	}
}

func TestCreateVariant(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.CreateVariantRequest
		f    func(context.Context, *flipt.CreateVariantRequest) (*flipt.Variant, error)
	}{
		{
			name: "ok",
			req: &flipt.CreateVariantRequest{
				FlagKey:     "flagKey",
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
			f: func(_ context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "key", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)

				return &flipt.Variant{
					FlagKey:     r.FlagKey,
					Key:         r.Key,
					Name:        r.Name,
					Description: r.Description,
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagService: &flagServiceMock{
					createVariantFn: tt.f,
				},
			}

			variant, err := s.CreateVariant(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.NotNil(t, variant)
		})
	}
}

func TestUpdateVariant(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.UpdateVariantRequest
		f    func(context.Context, *flipt.UpdateVariantRequest) (*flipt.Variant, error)
	}{
		{
			name: "ok",
			req:  &flipt.UpdateVariantRequest{Id: "id", FlagKey: "flagKey", Key: "key", Name: "name", Description: "desc"},
			f: func(_ context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "key", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)

				return &flipt.Variant{
					Id:          r.Id,
					FlagKey:     r.FlagKey,
					Key:         r.Key,
					Name:        r.Name,
					Description: r.Description,
				}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagService: &flagServiceMock{
					updateVariantFn: tt.f,
				},
			}

			variant, err := s.UpdateVariant(context.TODO(), tt.req)
			require.NoError(t, err)
			require.NotNil(t, variant)
		})
	}
}

func TestDeleteVariant(t *testing.T) {
	tests := []struct {
		name string
		req  *flipt.DeleteVariantRequest
		f    func(context.Context, *flipt.DeleteVariantRequest) error
	}{
		{
			name: "ok",
			req:  &flipt.DeleteVariantRequest{Id: "id", FlagKey: "flagKey"},
			f: func(_ context.Context, r *flipt.DeleteVariantRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagService: &flagServiceMock{
					deleteVariantFn: tt.f,
				},
			}

			resp, err := s.DeleteVariant(context.TODO(), tt.req)
			require.NoError(t, err)
			assert.Equal(t, &empty.Empty{}, resp)
		})
	}
}
