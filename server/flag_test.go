package server

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
	"github.com/stretchr/testify/assert"
)

var _ storage.FlagStore = &flagStoreMock{}

type flagStoreMock struct {
	getFlagFn       func(context.Context, *flipt.GetFlagRequest) (*flipt.Flag, error)
	listFlagsFn     func(context.Context, *flipt.ListFlagRequest) ([]*flipt.Flag, error)
	createFlagFn    func(context.Context, *flipt.CreateFlagRequest) (*flipt.Flag, error)
	updateFlagFn    func(context.Context, *flipt.UpdateFlagRequest) (*flipt.Flag, error)
	deleteFlagFn    func(context.Context, *flipt.DeleteFlagRequest) error
	createVariantFn func(context.Context, *flipt.CreateVariantRequest) (*flipt.Variant, error)
	updateVariantFn func(context.Context, *flipt.UpdateVariantRequest) (*flipt.Variant, error)
	deleteVariantFn func(context.Context, *flipt.DeleteVariantRequest) error
}

func (m *flagStoreMock) GetFlag(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
	return m.getFlagFn(ctx, r)
}

func (m *flagStoreMock) ListFlags(ctx context.Context, r *flipt.ListFlagRequest) ([]*flipt.Flag, error) {
	return m.listFlagsFn(ctx, r)
}

func (m *flagStoreMock) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	return m.createFlagFn(ctx, r)
}

func (m *flagStoreMock) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	return m.updateFlagFn(ctx, r)
}

func (m *flagStoreMock) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	return m.deleteFlagFn(ctx, r)
}

func (m *flagStoreMock) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	return m.createVariantFn(ctx, r)
}

func (m *flagStoreMock) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	return m.updateVariantFn(ctx, r)
}

func (m *flagStoreMock) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	return m.deleteVariantFn(ctx, r)
}

func TestGetFlag(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.GetFlagRequest
		f       func(context.Context, *flipt.GetFlagRequest) (*flipt.Flag, error)
		flag    *flipt.Flag
		wantErr error
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
			flag: &flipt.Flag{
				Key: "key",
			},
		},
		{
			name: "emptyKey",
			req:  &flipt.GetFlagRequest{Key: ""},
			f: func(_ context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.Key)

				return &flipt.Flag{
					Key: "",
				}, nil
			},
			flag:    nil,
			wantErr: emptyFieldError("key"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			wantErr = tt.wantErr
			flag    = tt.flag
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagStore: &flagStoreMock{
					getFlagFn: f,
				},
			}

			got, err := s.GetFlag(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, flag, got)
		})
	}
}

func TestListFlags(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.ListFlagRequest
		f       func(context.Context, *flipt.ListFlagRequest) ([]*flipt.Flag, error)
		flags   *flipt.FlagList
		wantErr error
	}{
		{
			name: "ok",
			req:  &flipt.ListFlagRequest{},
			f: func(context.Context, *flipt.ListFlagRequest) ([]*flipt.Flag, error) {
				return []*flipt.Flag{
					{Key: "flag"},
				}, nil
			},
			flags: &flipt.FlagList{
				Flags: []*flipt.Flag{
					{
						Key: "flag",
					},
				},
			},
		},
		{
			name: "err",
			req:  &flipt.ListFlagRequest{},
			f: func(context.Context, *flipt.ListFlagRequest) ([]*flipt.Flag, error) {
				return nil, errors.New("test error")
			},
			flags:   nil,
			wantErr: errors.New("test error"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			wantErr = tt.wantErr
			flags   = tt.flags
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagStore: &flagStoreMock{
					listFlagsFn: f,
				},
			}

			got, err := s.ListFlags(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, flags, got)
		})
	}
}

func TestCreateFlag(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.CreateFlagRequest
		f       func(context.Context, *flipt.CreateFlagRequest) (*flipt.Flag, error)
		flag    *flipt.Flag
		wantErr error
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
			flag: &flipt.Flag{
				Key:         "key",
				Name:        "name",
				Description: "desc",
				Enabled:     true,
			},
		},
		{
			name: "emptyKey",
			req: &flipt.CreateFlagRequest{
				Key:         "",
				Name:        "name",
				Description: "desc",
				Enabled:     true,
			},
			f: func(_ context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)
				assert.True(t, r.Enabled)

				return &flipt.Flag{
					Key:         "",
					Name:        r.Name,
					Description: r.Description,
					Enabled:     r.Enabled,
				}, nil
			},
			wantErr: emptyFieldError("key"),
		},
		{
			name: "emptyName",
			req: &flipt.CreateFlagRequest{
				Key:         "key",
				Name:        "",
				Description: "desc",
				Enabled:     true,
			},
			f: func(_ context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)
				assert.Equal(t, "", r.Name)
				assert.Equal(t, "desc", r.Description)
				assert.True(t, r.Enabled)

				return &flipt.Flag{
					Key:         r.Key,
					Name:        "",
					Description: r.Description,
					Enabled:     r.Enabled,
				}, nil
			},
			wantErr: emptyFieldError("name"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			wantErr = tt.wantErr
			flag    = tt.flag
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagStore: &flagStoreMock{
					createFlagFn: f,
				},
			}

			got, err := s.CreateFlag(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, flag, got)
		})
	}
}

func TestUpdateFlag(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.UpdateFlagRequest
		f       func(context.Context, *flipt.UpdateFlagRequest) (*flipt.Flag, error)
		flag    *flipt.Flag
		wantErr error
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
			flag: &flipt.Flag{
				Key:         "key",
				Name:        "name",
				Description: "desc",
				Enabled:     true,
			},
		},
		{
			name: "emptyKey",
			req: &flipt.UpdateFlagRequest{
				Key:         "",
				Name:        "name",
				Description: "desc",
				Enabled:     true,
			},
			f: func(_ context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)
				assert.True(t, r.Enabled)

				return &flipt.Flag{
					Key:         "",
					Name:        r.Name,
					Description: r.Description,
					Enabled:     r.Enabled,
				}, nil
			},
			wantErr: emptyFieldError("key"),
		},
		{
			name: "emptyName",
			req: &flipt.UpdateFlagRequest{
				Key:         "key",
				Name:        "",
				Description: "desc",
				Enabled:     true,
			},
			f: func(_ context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)
				assert.Equal(t, "", r.Name)
				assert.Equal(t, "desc", r.Description)
				assert.True(t, r.Enabled)

				return &flipt.Flag{
					Key:         r.Key,
					Name:        "",
					Description: r.Description,
					Enabled:     r.Enabled,
				}, nil
			},
			wantErr: emptyFieldError("name"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			wantErr = tt.wantErr
			flag    = tt.flag
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagStore: &flagStoreMock{
					updateFlagFn: f,
				},
			}

			got, err := s.UpdateFlag(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, flag, got)
		})
	}
}

func TestDeleteFlag(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.DeleteFlagRequest
		f       func(context.Context, *flipt.DeleteFlagRequest) error
		empty   *empty.Empty
		wantErr error
	}{
		{
			name: "ok",
			req:  &flipt.DeleteFlagRequest{Key: "key"},
			f: func(_ context.Context, r *flipt.DeleteFlagRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)

				return nil
			},
			empty: &empty.Empty{},
		},
		{
			name: "emptyKey",
			req:  &flipt.DeleteFlagRequest{Key: ""},
			f: func(_ context.Context, r *flipt.DeleteFlagRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.Key)

				return nil
			},
			wantErr: emptyFieldError("key"),
		},
		{
			name: "error",
			req:  &flipt.DeleteFlagRequest{Key: "key"},
			f: func(_ context.Context, r *flipt.DeleteFlagRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "key", r.Key)

				return errors.New("test error")
			},
			wantErr: errors.New("test error"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			wantErr = tt.wantErr
			empty   = tt.empty
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagStore: &flagStoreMock{
					deleteFlagFn: f,
				},
			}

			got, err := s.DeleteFlag(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, empty, got)
		})
	}
}

func TestCreateVariant(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.CreateVariantRequest
		f       func(context.Context, *flipt.CreateVariantRequest) (*flipt.Variant, error)
		variant *flipt.Variant
		wantErr error
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
			variant: &flipt.Variant{
				FlagKey:     "flagKey",
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
		},
		{
			name: "emptyFlagKey",
			req: &flipt.CreateVariantRequest{
				FlagKey:     "",
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
			f: func(_ context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.FlagKey)
				assert.Equal(t, "key", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)

				return &flipt.Variant{
					FlagKey:     "",
					Key:         r.Key,
					Name:        r.Name,
					Description: r.Description,
				}, nil
			},
			wantErr: emptyFieldError("flagKey"),
		},
		{
			name: "emptyKey",
			req: &flipt.CreateVariantRequest{
				FlagKey:     "flagKey",
				Key:         "",
				Name:        "name",
				Description: "desc",
			},
			f: func(_ context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)

				return &flipt.Variant{
					FlagKey:     r.FlagKey,
					Key:         "",
					Name:        r.Name,
					Description: r.Description,
				}, nil
			},
			wantErr: emptyFieldError("key"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			wantErr = tt.wantErr
			variant = tt.variant
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagStore: &flagStoreMock{
					createVariantFn: f,
				},
			}

			got, err := s.CreateVariant(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, variant, got)
		})
	}
}

func TestUpdateVariant(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.UpdateVariantRequest
		f       func(context.Context, *flipt.UpdateVariantRequest) (*flipt.Variant, error)
		variant *flipt.Variant
		wantErr error
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
			variant: &flipt.Variant{
				Id:          "id",
				FlagKey:     "flagKey",
				Key:         "key",
				Name:        "name",
				Description: "desc",
			},
		},
		{
			name: "emptyID",
			req:  &flipt.UpdateVariantRequest{Id: "", FlagKey: "flagKey", Key: "key", Name: "name", Description: "desc"},
			f: func(_ context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "", r.Id)
				assert.Equal(t, "key", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)

				return &flipt.Variant{
					Id:          "",
					FlagKey:     r.FlagKey,
					Key:         r.Key,
					Name:        r.Name,
					Description: r.Description,
				}, nil
			},
			wantErr: emptyFieldError("id"),
		},
		{
			name: "emptyFlagKey",
			req:  &flipt.UpdateVariantRequest{Id: "id", FlagKey: "", Key: "key", Name: "name", Description: "desc"},
			f: func(_ context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.FlagKey)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "key", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)

				return &flipt.Variant{
					Id:          r.Id,
					FlagKey:     "",
					Key:         r.Key,
					Name:        r.Name,
					Description: r.Description,
				}, nil
			},
			wantErr: emptyFieldError("flagKey"),
		},
		{
			name: "emptyKey",
			req:  &flipt.UpdateVariantRequest{Id: "id", FlagKey: "flagKey", Key: "", Name: "name", Description: "desc"},
			f: func(_ context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
				assert.NotNil(t, r)
				assert.Equal(t, "flagKey", r.FlagKey)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "", r.Key)
				assert.Equal(t, "name", r.Name)
				assert.Equal(t, "desc", r.Description)

				return &flipt.Variant{
					Id:          r.Id,
					FlagKey:     r.FlagKey,
					Key:         "",
					Name:        r.Name,
					Description: r.Description,
				}, nil
			},
			wantErr: emptyFieldError("key"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			wantErr = tt.wantErr
			variant = tt.variant
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagStore: &flagStoreMock{
					updateVariantFn: f,
				},
			}

			got, err := s.UpdateVariant(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, variant, got)
		})
	}
}

func TestDeleteVariant(t *testing.T) {
	tests := []struct {
		name    string
		req     *flipt.DeleteVariantRequest
		f       func(context.Context, *flipt.DeleteVariantRequest) error
		empty   *empty.Empty
		wantErr error
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
			empty: &empty.Empty{},
		},
		{
			name: "emptyID",
			req:  &flipt.DeleteVariantRequest{Id: "", FlagKey: "flagKey"},
			f: func(_ context.Context, r *flipt.DeleteVariantRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)

				return nil
			},
			wantErr: emptyFieldError("id"),
		},
		{
			name: "emptyFlagKey",
			req:  &flipt.DeleteVariantRequest{Id: "id", FlagKey: ""},
			f: func(_ context.Context, r *flipt.DeleteVariantRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "", r.FlagKey)

				return nil
			},
			wantErr: emptyFieldError("flagKey"),
		},
		{
			name: "error",
			req:  &flipt.DeleteVariantRequest{Id: "id", FlagKey: "flagKey"},
			f: func(_ context.Context, r *flipt.DeleteVariantRequest) error {
				assert.NotNil(t, r)
				assert.Equal(t, "id", r.Id)
				assert.Equal(t, "flagKey", r.FlagKey)

				return errors.New("error test")
			},
			wantErr: errors.New("error test"),
		},
	}

	for _, tt := range tests {
		var (
			f       = tt.f
			req     = tt.req
			wantErr = tt.wantErr
			empty   = tt.empty
		)

		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				FlagStore: &flagStoreMock{
					deleteVariantFn: f,
				},
			}

			got, err := s.DeleteVariant(context.TODO(), req)
			assert.Equal(t, wantErr, err)
			assert.Equal(t, empty, got)
		})
	}
}
