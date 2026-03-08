package environments

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/coss/license"
	"go.flipt.io/flipt/internal/product"
	"go.flipt.io/flipt/internal/server/authz"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ environments.EnvironmentsServiceServer = (*Server)(nil)

type Server struct {
	logger         *zap.Logger
	licenseManager license.Manager

	envs *EnvironmentStore

	environments.UnimplementedEnvironmentsServiceServer
}

func NewServer(logger *zap.Logger, licenseManager license.Manager, envs *EnvironmentStore) (_ *Server, err error) {
	return &Server{logger: logger, licenseManager: licenseManager, envs: envs}, nil
}

func (s *Server) requirePro() error {
	if s.licenseManager == nil || s.licenseManager.Product() != product.Pro {
		return errors.ErrUnauthenticatedf("this feature requires a Flipt Pro license")
	}
	return nil
}

// RegisterGRPC registers the *Server onto the provided grpc Server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	environments.RegisterEnvironmentsServiceServer(server, s)
}

// ListEnvironments returns a list of all environments.
func (s *Server) ListEnvironments(ctx context.Context, req *environments.ListEnvironmentsRequest) (el *environments.ListEnvironmentsResponse, err error) {
	el = &environments.ListEnvironmentsResponse{}

	// First collect all environments
	for env := range s.envs.List(ctx) {
		cfg := env.Configuration()

		el.Environments = append(el.Environments, &environments.Environment{
			Key:           env.Key(),
			Name:          env.Key(),
			Default:       new(env.Default()),
			Configuration: cfg,
		})
	}

	// Only filter if we have viewable environments in context
	viewableEnvironments, ok := ctx.Value(authz.EnvironmentsKey).([]string)
	if ok {
		// if user has access to all environments, return all environments
		if len(viewableEnvironments) == 1 && viewableEnvironments[0] == "*" {
			return el, nil
		}

		// filter environments based on viewable environments
		el.Environments = lo.Filter(el.Environments, func(env *environments.Environment, _ int) bool {
			return lo.Contains(viewableEnvironments, env.Key)
		})
	}

	return el, nil
}

func (s *Server) BranchEnvironment(ctx context.Context, req *environments.BranchEnvironmentRequest) (resp *environments.Environment, err error) {
	env, err := s.envs.Branch(ctx, req.EnvironmentKey, req.Key)
	if err != nil {
		return nil, err
	}

	return &environments.Environment{
		Key:           env.Key(),
		Name:          env.Key(),
		Default:       new(env.Default()),
		Configuration: env.Configuration(),
	}, nil
}

func (s *Server) DeleteBranchEnvironment(ctx context.Context, req *environments.DeleteBranchEnvironmentRequest) (resp *emptypb.Empty, err error) {
	return &emptypb.Empty{}, s.envs.DeleteBranch(ctx, req.EnvironmentKey, req.Key)
}

func (s *Server) ListEnvironmentBranches(ctx context.Context, req *environments.ListEnvironmentBranchesRequest) (br *environments.ListEnvironmentBranchesResponse, err error) {
	env, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, err
	}

	return env.ListBranches(ctx)
}

func (s *Server) ListBranchedEnvironmentChanges(ctx context.Context, req *environments.ListBranchedEnvironmentChangesRequest) (br *environments.ListBranchedEnvironmentChangesResponse, err error) {
	env, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, err
	}

	branch, err := env.Branch(ctx, req.Key)
	if err != nil {
		return nil, err
	}

	return env.ListBranchedChanges(ctx, branch)
}

func (s *Server) ProposeEnvironment(ctx context.Context, req *environments.ProposeEnvironmentRequest) (resp *environments.EnvironmentProposalDetails, err error) {
	env, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, err
	}

	branch, err := env.Branch(ctx, req.Key)
	if err != nil {
		return nil, err
	}

	opts := ProposalOptions{}

	if req.Title != nil {
		opts.Title = *req.Title
	}

	if req.Body != nil {
		opts.Body = *req.Body
	}

	if req.Draft != nil {
		opts.Draft = *req.Draft
	}

	return env.Propose(ctx, branch, opts)
}

func (s *Server) GetNamespace(ctx context.Context, req *environments.GetNamespaceRequest) (ns *environments.NamespaceResponse, err error) {
	env, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, err
	}

	return env.GetNamespace(ctx, req.Key)
}

func (s *Server) ListNamespaces(ctx context.Context, req *environments.ListNamespacesRequest) (nl *environments.ListNamespacesResponse, err error) {
	env, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, err
	}

	namespaces, err := env.ListNamespaces(ctx)
	if err != nil {
		return nil, err
	}

	viewableNamespaces, ok := ctx.Value(authz.NamespacesKey).([]string)
	if ok {
		// if user has access to all namespaces, return all namespaces
		if len(viewableNamespaces) == 1 && viewableNamespaces[0] == "*" {
			return namespaces, nil
		}

		namespaces.Items = lo.Filter(namespaces.Items, func(ns *environments.Namespace, _ int) bool {
			return lo.Contains(viewableNamespaces, ns.Key)
		})
	}

	return namespaces, nil
}

func (s *Server) CreateNamespace(ctx context.Context, ns *environments.UpdateNamespaceRequest) (*environments.NamespaceResponse, error) {
	env, err := s.envs.Get(ctx, ns.EnvironmentKey)
	if err != nil {
		return nil, err
	}

	_, err = env.GetNamespace(ctx, ns.Key)
	if err == nil {
		return nil, errors.ErrAlreadyExistsf("create namespace %q", ns.Key)
	}

	if !errors.AsMatch[errors.ErrNotFound](err) {
		return nil, err
	}

	resp := &environments.NamespaceResponse{
		Namespace: &environments.Namespace{
			Key:         ns.Key,
			Name:        ns.Name,
			Description: ns.Description,
			Protected:   ns.Protected,
		},
	}

	resp.Revision, err = env.CreateNamespace(ctx, ns.GetRevision(), resp.Namespace)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *Server) UpdateNamespace(ctx context.Context, ns *environments.UpdateNamespaceRequest) (*environments.NamespaceResponse, error) {
	env, err := s.envs.Get(ctx, ns.EnvironmentKey)
	if err != nil {
		return nil, err
	}

	_, err = env.GetNamespace(ctx, ns.Key)
	if err != nil {
		return nil, fmt.Errorf("update namespace %q: %w", ns.Key, err)
	}

	resp := &environments.NamespaceResponse{
		Namespace: &environments.Namespace{
			Key:         ns.Key,
			Name:        ns.Name,
			Description: ns.Description,
			Protected:   ns.Protected,
		},
	}

	resp.Revision, err = env.UpdateNamespace(ctx, ns.GetRevision(), resp.Namespace)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *Server) DeleteNamespace(ctx context.Context, req *environments.DeleteNamespaceRequest) (_ *environments.DeleteNamespaceResponse, err error) {
	env, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, err
	}

	resp := &environments.DeleteNamespaceResponse{}

	resp.Revision, err = env.DeleteNamespace(ctx, req.Revision, req.Key)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *Server) GetResource(ctx context.Context, req *environments.GetResourceRequest) (r *environments.ResourceResponse, err error) {
	env, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, err
	}

	typ, err := ParseResourceType(req.TypeUrl)
	if err != nil {
		return nil, err
	}

	return r, env.View(ctx, typ, func(ctx context.Context, sv ResourceStoreView) (err error) {
		r, err = sv.GetResource(ctx, req.NamespaceKey, req.Key)
		if err != nil {
			return err
		}
		return err
	})
}

func (s *Server) ListResources(ctx context.Context, req *environments.ListResourcesRequest) (rl *environments.ListResourcesResponse, err error) {
	env, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, err
	}

	typ, err := ParseResourceType(req.TypeUrl)
	if err != nil {
		return nil, err
	}

	if err := env.View(ctx, typ, func(ctx context.Context, sv ResourceStoreView) (err error) {
		rl, err = sv.ListResources(ctx, req.NamespaceKey)
		if err != nil {
			return err
		}

		return err
	}); err != nil {
		return nil, err
	}

	return rl, nil
}

func (s *Server) CreateResource(ctx context.Context, req *environments.UpdateResourceRequest) (_ *environments.ResourceResponse, err error) {
	env, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, err
	}

	resp := &environments.ResourceResponse{
		Resource: &environments.Resource{
			NamespaceKey: req.NamespaceKey,
			Key:          req.Key,
			Payload:      req.Payload,
		},
	}

	typ, err := ParseResourceType(req.Payload.GetTypeUrl())
	if err != nil {
		return nil, err
	}

	resp.Revision, err = env.Update(ctx, req.Revision, typ, func(ctx context.Context, sv ResourceStore) error {
		_, err := sv.GetResource(ctx, req.NamespaceKey, req.Key)
		if err == nil {
			return errors.ErrAlreadyExistsf(`create resource "%s/%s/%s"`, req.Payload.GetTypeUrl(), req.NamespaceKey, req.Key)
		}

		if !errors.AsMatch[errors.ErrNotFound](err) {
			return err
		}

		return sv.CreateResource(ctx, resp.Resource)
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *Server) UpdateResource(ctx context.Context, req *environments.UpdateResourceRequest) (_ *environments.ResourceResponse, err error) {
	env, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, err
	}

	resp := &environments.ResourceResponse{
		Resource: &environments.Resource{
			NamespaceKey: req.NamespaceKey,
			Key:          req.Key,
			Payload:      req.Payload,
		},
	}

	typ, err := ParseResourceType(req.Payload.GetTypeUrl())
	if err != nil {
		return nil, err
	}

	resp.Revision, err = env.Update(ctx, req.Revision, typ, func(ctx context.Context, sv ResourceStore) error {
		_, err := sv.GetResource(ctx, req.NamespaceKey, req.Key)
		if err != nil {
			return fmt.Errorf(`update resource "%s/%s/%s": %w`, req.Payload.GetTypeUrl(), req.NamespaceKey, req.Key, err)
		}

		return sv.UpdateResource(ctx, resp.Resource)
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CopyResource copies a single resource from one environment/namespace to another.
func (s *Server) CopyResource(ctx context.Context, req *environments.CopyResourceRequest) (*environments.CopyResourceResponse, error) {
	if err := s.requirePro(); err != nil {
		return nil, err
	}

	typ, err := ParseResourceType(req.TypeUrl)
	if err != nil {
		return nil, err
	}

	// Read from source environment.
	srcEnv, err := s.envs.Get(ctx, req.SourceEnvironmentKey)
	if err != nil {
		return nil, fmt.Errorf("source environment: %w", err)
	}

	var srcResource *environments.Resource
	if err := srcEnv.View(ctx, typ, func(ctx context.Context, sv ResourceStoreView) error {
		resp, err := sv.GetResource(ctx, req.SourceNamespaceKey, req.Key)
		if err != nil {
			return fmt.Errorf("source resource %q/%q: %w", req.SourceNamespaceKey, req.Key, err)
		}
		srcResource = resp.Resource
		return nil
	}); err != nil {
		return nil, err
	}

	// Write to target environment.
	tgtEnv, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, fmt.Errorf("target environment: %w", err)
	}

	resp := &environments.CopyResourceResponse{
		Resource: &environments.Resource{
			NamespaceKey: req.NamespaceKey,
			Key:          srcResource.Key,
			Payload:      srcResource.Payload,
		},
	}

	resp.Revision, err = tgtEnv.Update(ctx, req.Revision, typ, func(ctx context.Context, sv ResourceStore) error {
		_, err := sv.GetResource(ctx, req.NamespaceKey, req.Key)
		exists := err == nil

		if err != nil && !errors.AsMatch[errors.ErrNotFound](err) {
			return err
		}

		switch {
		case exists && req.OnConflict == environments.ConflictStrategy_CONFLICT_STRATEGY_FAIL:
			return errors.ErrAlreadyExistsf("resource %q/%q/%q", req.TypeUrl, req.NamespaceKey, req.Key)
		case exists && req.OnConflict == environments.ConflictStrategy_CONFLICT_STRATEGY_SKIP:
			resp.Status = environments.OperationStatus_OPERATION_STATUS_SKIPPED
			return nil
		case exists: // OVERWRITE
			resp.Status = environments.OperationStatus_OPERATION_STATUS_SUCCESS
			return sv.UpdateResource(ctx, resp.Resource)
		default: // doesn't exist — create
			resp.Status = environments.OperationStatus_OPERATION_STATUS_SUCCESS
			return sv.CreateResource(ctx, resp.Resource)
		}
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *Server) DeleteResource(ctx context.Context, req *environments.DeleteResourceRequest) (_ *environments.DeleteResourceResponse, err error) {
	env, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, err
	}

	typ, err := ParseResourceType(req.TypeUrl)
	if err != nil {
		return nil, err
	}

	resp := &environments.DeleteResourceResponse{}
	if resp.Revision, err = env.Update(ctx, req.Revision, typ, func(ctx context.Context, sv ResourceStore) (err error) {
		return sv.DeleteResource(ctx, req.NamespaceKey, req.Key)
	}); err != nil {
		return nil, err
	}

	return resp, nil
}

// knownResourceTypes lists all resource types that should be copied during namespace copy.
var knownResourceTypes = []ResourceType{
	NewResourceType("flipt.core", "Flag"),
	NewResourceType("flipt.core", "Segment"),
}

// CopyNamespace copies all resources from a source namespace/environment to a target.
func (s *Server) CopyNamespace(ctx context.Context, req *environments.CopyNamespaceRequest) (*environments.CopyNamespaceResponse, error) {
	if err := s.requirePro(); err != nil {
		return nil, err
	}

	// Read all resources from source namespace.
	srcEnv, err := s.envs.Get(ctx, req.SourceEnvironmentKey)
	if err != nil {
		return nil, fmt.Errorf("source environment: %w", err)
	}

	type typedResources struct {
		typ       ResourceType
		resources []*environments.Resource
	}

	var allResources []typedResources
	for _, typ := range knownResourceTypes {
		var resources []*environments.Resource
		if err := srcEnv.View(ctx, typ, func(ctx context.Context, sv ResourceStoreView) error {
			resp, err := sv.ListResources(ctx, req.SourceNamespaceKey)
			if err != nil {
				return err
			}
			resources = resp.Resources
			return nil
		}); err != nil {
			return nil, fmt.Errorf("listing source %s resources: %w", typ, err)
		}
		if len(resources) > 0 {
			allResources = append(allResources, typedResources{typ: typ, resources: resources})
		}
	}

	// Resolve target environment.
	tgtEnv, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, fmt.Errorf("target environment: %w", err)
	}

	// Ensure target namespace exists.
	_, err = tgtEnv.GetNamespace(ctx, req.NamespaceKey)
	if errors.AsMatch[errors.ErrNotFound](err) {
		rev, createErr := tgtEnv.CreateNamespace(ctx, req.Revision, &environments.Namespace{
			Key:  req.NamespaceKey,
			Name: req.NamespaceKey,
		})
		if createErr != nil {
			return nil, fmt.Errorf("creating target namespace: %w", createErr)
		}
		req.Revision = rev
	} else if err != nil {
		return nil, fmt.Errorf("checking target namespace: %w", err)
	}

	resp := &environments.CopyNamespaceResponse{}

	for _, tr := range allResources {
		rev, err := tgtEnv.Update(ctx, req.Revision, tr.typ, func(ctx context.Context, sv ResourceStore) error {
			for _, src := range tr.resources {
				tgtResource := &environments.Resource{
					NamespaceKey: req.NamespaceKey,
					Key:          src.Key,
					Payload:      src.Payload,
				}

				_, getErr := sv.GetResource(ctx, req.NamespaceKey, src.Key)
				exists := getErr == nil

				if getErr != nil && !errors.AsMatch[errors.ErrNotFound](getErr) {
					msg := getErr.Error()
					resp.Results = append(resp.Results, &environments.CopyNamespaceResourceResult{
						TypeUrl: tr.typ.String(),
						Key:     src.Key,
						Status:  environments.OperationStatus_OPERATION_STATUS_FAILED,
						Error:   &msg,
					})
					continue
				}

				result := &environments.CopyNamespaceResourceResult{
					TypeUrl: tr.typ.String(),
					Key:     src.Key,
				}

				switch {
				case exists && req.OnConflict == environments.ConflictStrategy_CONFLICT_STRATEGY_FAIL:
					result.Status = environments.OperationStatus_OPERATION_STATUS_FAILED
					msg := fmt.Sprintf("resource already exists: %s/%s", req.NamespaceKey, src.Key)
					result.Error = &msg
				case exists && req.OnConflict == environments.ConflictStrategy_CONFLICT_STRATEGY_SKIP:
					result.Status = environments.OperationStatus_OPERATION_STATUS_SKIPPED
				case exists: // OVERWRITE
					if err := sv.UpdateResource(ctx, tgtResource); err != nil {
						result.Status = environments.OperationStatus_OPERATION_STATUS_FAILED
						msg := err.Error()
						result.Error = &msg
					} else {
						result.Status = environments.OperationStatus_OPERATION_STATUS_SUCCESS
					}
				default: // create
					if err := sv.CreateResource(ctx, tgtResource); err != nil {
						result.Status = environments.OperationStatus_OPERATION_STATUS_FAILED
						msg := err.Error()
						result.Error = &msg
					} else {
						result.Status = environments.OperationStatus_OPERATION_STATUS_SUCCESS
					}
				}

				resp.Results = append(resp.Results, result)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		req.Revision = rev
	}

	resp.Revision = req.Revision
	return resp, nil
}

// BulkApplyResources applies an operation to a resource across multiple namespaces.
func (s *Server) BulkApplyResources(ctx context.Context, req *environments.BulkApplyResourcesRequest) (*environments.BulkApplyResourcesResponse, error) {
	if err := s.requirePro(); err != nil {
		return nil, err
	}

	env, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, err
	}

	typeUrl := req.TypeUrl
	if typeUrl == "" && req.Payload != nil {
		typeUrl = req.Payload.GetTypeUrl()
	}

	typ, err := ParseResourceType(typeUrl)
	if err != nil {
		return nil, err
	}

	resp := &environments.BulkApplyResourcesResponse{}

	resp.Revision, err = env.Update(ctx, req.Revision, typ, func(ctx context.Context, sv ResourceStore) error {
		for _, nsKey := range req.NamespaceKeys {
			result := &environments.BulkApplyNamespaceResult{
				NamespaceKey: nsKey,
			}

			var opErr error
			switch req.Operation {
			case environments.BulkOperation_BULK_OPERATION_CREATE:
				resource := &environments.Resource{
					NamespaceKey: nsKey,
					Key:          req.Key,
					Payload:      req.Payload,
				}
				_, getErr := sv.GetResource(ctx, nsKey, req.Key)
				if getErr == nil {
					switch req.OnConflict {
					case environments.ConflictStrategy_CONFLICT_STRATEGY_SKIP:
						result.Status = environments.OperationStatus_OPERATION_STATUS_SKIPPED
					case environments.ConflictStrategy_CONFLICT_STRATEGY_OVERWRITE:
						opErr = sv.UpdateResource(ctx, resource)
					default:
						opErr = errors.ErrAlreadyExistsf("resource %q in namespace %q", req.Key, nsKey)
					}
				} else {
					opErr = sv.CreateResource(ctx, resource)
				}

			case environments.BulkOperation_BULK_OPERATION_UPDATE:
				resource := &environments.Resource{
					NamespaceKey: nsKey,
					Key:          req.Key,
					Payload:      req.Payload,
				}
				opErr = sv.UpdateResource(ctx, resource)

			case environments.BulkOperation_BULK_OPERATION_DELETE:
				opErr = sv.DeleteResource(ctx, nsKey, req.Key)

			case environments.BulkOperation_BULK_OPERATION_UPSERT:
				resource := &environments.Resource{
					NamespaceKey: nsKey,
					Key:          req.Key,
					Payload:      req.Payload,
				}
				_, getErr := sv.GetResource(ctx, nsKey, req.Key)
				if getErr == nil {
					opErr = sv.UpdateResource(ctx, resource)
				} else {
					opErr = sv.CreateResource(ctx, resource)
				}
			}

			if opErr != nil {
				result.Status = environments.OperationStatus_OPERATION_STATUS_FAILED
				msg := opErr.Error()
				result.Error = &msg
			} else if result.Status != environments.OperationStatus_OPERATION_STATUS_SKIPPED {
				result.Status = environments.OperationStatus_OPERATION_STATUS_SUCCESS
			}

			resp.Results = append(resp.Results, result)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
