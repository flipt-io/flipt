package server

import (
	"context"

	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/core"
	"go.uber.org/zap"
)

// ListFlags lists all flags
func (s *Server) ListFlags(ctx context.Context, r *flipt.ListFlagRequest) (*flipt.FlagList, error) {
	s.logger.Debug("list flags", zap.Stringer("request", r))

	// Check for X-Environment header for backward compatibility
	environmentKey := r.EnvironmentKey
	if headerEnv, ok := common.FliptEnvironmentFromContext(ctx); ok && headerEnv != "" {
		environmentKey = headerEnv
	}

	ctx = common.WithFliptEnvironment(ctx, environmentKey)
	ctx = common.WithFliptNamespace(ctx, r.NamespaceKey)

	store, err := s.getStore(ctx)
	if err != nil {
		return nil, err
	}

	ns := storage.NewNamespace(r.NamespaceKey, storage.WithReference(r.GetReference()))
	results, err := store.ListFlags(ctx, storage.ListWithParameters(ns, r))
	if err != nil {
		return nil, err
	}

	resp := flipt.FlagList{}

	for _, flag := range results.Results {
		var defaultVariant *flipt.Variant
		if flag.DefaultVariant != nil {
			for _, variant := range flag.Variants {
				if *flag.DefaultVariant != variant.Key {
					continue
				}

				var attachment string
				if variant.Attachment != nil {
					v, err := variant.Attachment.MarshalJSON()
					if err != nil {
						return nil, err
					}

					attachment = string(v)
				}

				defaultVariant = &flipt.Variant{
					FlagKey:      flag.Key,
					NamespaceKey: r.NamespaceKey,
					Key:          variant.Key,
					Name:         variant.Name,
					Description:  variant.Description,
					Attachment:   attachment,
				}
				break
			}
		}

		resp.Flags = append(resp.Flags, &flipt.Flag{
			Key:            flag.Key,
			Name:           flag.Name,
			Description:    flag.Description,
			DefaultVariant: defaultVariant,
			Type:           toListFlagType(flag.Type),
			Enabled:        flag.Enabled,
			Metadata:       flag.Metadata,
		})
	}

	total, err := store.CountFlags(ctx, ns)
	if err != nil {
		return nil, err
	}

	resp.TotalCount = int32(total)
	resp.NextPageToken = results.NextPageToken

	s.logger.Debug("list flags", zap.Stringer("response", &resp))
	return &resp, nil
}

func toListFlagType(flagType core.FlagType) flipt.FlagType {
	if flagType == core.FlagType_BOOLEAN_FLAG_TYPE {
		return flipt.FlagType_BOOLEAN_FLAG_TYPE
	}
	return flipt.FlagType_VARIANT_FLAG_TYPE
}
