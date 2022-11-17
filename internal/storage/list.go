package storage

import (
	"context"
)

// ListFunc is a function which can return a set of results for a list request.
type ListFunc[P, V any] func(context.Context, *ListRequest[P]) (ResultSet[V], error)

type ListAllParams struct {
	PerPage int
	Order   Order
}

// ListAll can return the entire contents of some generic storage layer if given
// a ListFunc implementation for that store.
// It performs an entire paginated walk until an empty next page token is returned.
func ListAll[P, V any](ctx context.Context, fn ListFunc[P, V], params ListAllParams) (res []V, err error) {
	var req *ListRequest[P]

	for {
		if req != nil && req.QueryParams.PageToken == "" {
			break
		}

		if req == nil {
			req = &ListRequest[P]{
				QueryParams: QueryParams{
					Limit: uint64(params.PerPage),
					Order: params.Order,
				},
			}
		}

		set, err := fn(ctx, req)
		if err != nil {
			return nil, err
		}

		req.QueryParams.PageToken = set.NextPageToken
		res = append(res, set.Results...)
	}

	return res, nil
}
