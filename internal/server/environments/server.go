package environments

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/samber/lo"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/coss/license"
	"go.flipt.io/flipt/internal/product"
	"go.flipt.io/flipt/internal/server/authz"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
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
		// Pre-check which resources exist in the target to separate
		// resources that need writing from those that can be skipped/failed.
		existingKeys := make(map[string]bool)
		if err := tgtEnv.View(ctx, tr.typ, func(ctx context.Context, sv ResourceStoreView) error {
			for _, src := range tr.resources {
				_, err := sv.GetResource(ctx, req.NamespaceKey, src.Key)
				if err == nil {
					existingKeys[src.Key] = true
				} else if !errors.AsMatch[errors.ErrNotFound](err) {
					msg := err.Error()
					resp.Results = append(resp.Results, &environments.CopyNamespaceResourceResult{
						TypeUrl: tr.typ.String(),
						Key:     src.Key,
						Status:  environments.OperationStatus_OPERATION_STATUS_FAILED,
						Error:   &msg,
					})
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}

		// Collect resources that need to be written.
		var toWrite []*environments.Resource
		for _, src := range tr.resources {
			exists := existingKeys[src.Key]

			// Check if already handled as a failed lookup above.
			alreadyHandled := false
			for _, r := range resp.Results {
				if r.TypeUrl == tr.typ.String() && r.Key == src.Key {
					alreadyHandled = true
					break
				}
			}
			if alreadyHandled {
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
			default:
				// Needs to be written (create or overwrite).
				toWrite = append(toWrite, &environments.Resource{
					NamespaceKey: req.NamespaceKey,
					Key:          src.Key,
					Payload:      src.Payload,
				})
			}

			// Only append non-write results here; write results are appended after Update.
			if result.Status != 0 {
				resp.Results = append(resp.Results, result)
			}
		}

		// Only call Update if there are resources to write.
		if len(toWrite) > 0 {
			rev, err := tgtEnv.Update(ctx, req.Revision, tr.typ, func(ctx context.Context, sv ResourceStore) error {
				for _, tgtResource := range toWrite {
					result := &environments.CopyNamespaceResourceResult{
						TypeUrl: tr.typ.String(),
						Key:     tgtResource.Key,
					}

					exists := existingKeys[tgtResource.Key]
					var writeErr error
					if exists {
						writeErr = sv.UpdateResource(ctx, tgtResource)
					} else {
						writeErr = sv.CreateResource(ctx, tgtResource)
					}

					if writeErr != nil {
						result.Status = environments.OperationStatus_OPERATION_STATUS_FAILED
						msg := writeErr.Error()
						result.Error = &msg
					} else {
						result.Status = environments.OperationStatus_OPERATION_STATUS_SUCCESS
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
	}

	resp.Revision = req.Revision
	return resp, nil
}

// CompareEnvironments compares resources across source and target env/namespace pairs.
func (s *Server) CompareEnvironments(ctx context.Context, req *environments.CompareEnvironmentsRequest) (*environments.CompareEnvironmentsResponse, error) {
	srcEnv, err := s.envs.Get(ctx, req.EnvironmentKey)
	if err != nil {
		return nil, fmt.Errorf("source environment: %w", err)
	}

	tgtEnv, err := s.envs.Get(ctx, req.TargetEnvironmentKey)
	if err != nil {
		return nil, fmt.Errorf("target environment: %w", err)
	}

	types := knownResourceTypes
	if len(req.TypeUrls) > 0 {
		types = make([]ResourceType, 0, len(req.TypeUrls))
		for _, typeURL := range req.TypeUrls {
			typ, parseErr := ParseResourceType(typeURL)
			if parseErr != nil {
				return nil, parseErr
			}
			types = append(types, typ)
		}
	}

	resp := &environments.CompareEnvironmentsResponse{}
	for _, typ := range types {
		sourceByKey := map[string]*environments.Resource{}
		if err := srcEnv.View(ctx, typ, func(ctx context.Context, sv ResourceStoreView) error {
			resources, listErr := sv.ListResources(ctx, req.NamespaceKey)
			if listErr != nil {
				return listErr
			}
			for _, resource := range resources.Resources {
				sourceByKey[resource.Key] = resource
			}
			return nil
		}); err != nil {
			return nil, err
		}

		targetByKey := map[string]*environments.Resource{}
		if err := tgtEnv.View(ctx, typ, func(ctx context.Context, sv ResourceStoreView) error {
			resources, listErr := sv.ListResources(ctx, req.TargetNamespaceKey)
			if listErr != nil {
				return listErr
			}
			for _, resource := range resources.Resources {
				targetByKey[resource.Key] = resource
			}
			return nil
		}); err != nil {
			return nil, err
		}

		keys := make([]string, 0, len(sourceByKey)+len(targetByKey))
		seen := make(map[string]struct{}, len(sourceByKey)+len(targetByKey))
		for key := range sourceByKey {
			seen[key] = struct{}{}
			keys = append(keys, key)
		}
		for key := range targetByKey {
			if _, ok := seen[key]; ok {
				continue
			}
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			source := sourceByKey[key]
			target := targetByKey[key]
			result := &environments.CompareResourceResult{
				TypeUrl: typ.String(),
				Key:     key,
				Source:  source,
				Target:  target,
			}

			switch {
			case source != nil && target == nil:
				result.Status = environments.CompareStatus_COMPARE_STATUS_SOURCE_ONLY
			case source == nil && target != nil:
				result.Status = environments.CompareStatus_COMPARE_STATUS_TARGET_ONLY
			case source != nil && target != nil && !proto.Equal(source.Payload, target.Payload):
				result.Status = environments.CompareStatus_COMPARE_STATUS_DIFFERENT
			default:
				result.Status = environments.CompareStatus_COMPARE_STATUS_IDENTICAL
			}

			resp.Results = append(resp.Results, result)
		}
	}

	return resp, nil
}

// BulkApplyResources applies an operation to a resource across multiple namespaces.
func (s *Server) BulkApplyResources(ctx context.Context, req *environments.BulkApplyResourcesRequest) (*environments.BulkApplyResourcesResponse, error) {
	if err := s.requirePro(); err != nil {
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
	environmentKeys := req.GetEnvironmentKeys()
	if len(environmentKeys) == 0 {
		environmentKeys = []string{req.EnvironmentKey}
	}

	seen := make(map[string]struct{}, len(environmentKeys))
	for _, environmentKey := range environmentKeys {
		if _, ok := seen[environmentKey]; ok {
			continue
		}

		seen[environmentKey] = struct{}{}
		env, err := s.envs.Get(ctx, environmentKey)
		if err != nil {
			for _, nsKey := range req.NamespaceKeys {
				msg := err.Error()
				resp.Results = append(resp.Results, &environments.BulkApplyNamespaceResult{
					EnvironmentKey: environmentKey,
					NamespaceKey:   nsKey,
					Status:         environments.OperationStatus_OPERATION_STATUS_FAILED,
					Error:          &msg,
				})
			}
			continue
		}

		revision := ""
		if environmentKey == req.EnvironmentKey {
			revision = req.Revision
		}

		wroteAny := false
		envResults := make([]*environments.BulkApplyNamespaceResult, 0, len(req.NamespaceKeys))
		envRevision, err := env.Update(ctx, revision, typ, func(ctx context.Context, sv ResourceStore) error {
			for _, nsKey := range req.NamespaceKeys {
				result := &environments.BulkApplyNamespaceResult{
					EnvironmentKey: environmentKey,
					NamespaceKey:   nsKey,
				}

				status, wrote, opErr := applyBulkOperation(ctx, sv, nsKey, req)
				if opErr != nil {
					result.Status = environments.OperationStatus_OPERATION_STATUS_FAILED
					msg := opErr.Error()
					result.Error = &msg
				} else {
					result.Status = status
					wroteAny = wroteAny || wrote
				}

				envResults = append(envResults, result)
			}
			return nil
		})
		if err != nil {
			// No-op conflict paths can legitimately produce no writes. In that case
			// ignore empty commit errors and preserve per-namespace results.
			if !wroteAny && strings.Contains(err.Error(), "cannot create empty commit") {
				resp.Results = append(resp.Results, envResults...)
				if environmentKey == req.EnvironmentKey {
					resp.Revision = req.Revision
				}
				continue
			}

			msg := err.Error()
			if len(envResults) == 0 {
				for _, nsKey := range req.NamespaceKeys {
					resp.Results = append(resp.Results, &environments.BulkApplyNamespaceResult{
						EnvironmentKey: environmentKey,
						NamespaceKey:   nsKey,
						Status:         environments.OperationStatus_OPERATION_STATUS_FAILED,
						Error:          &msg,
					})
				}
				continue
			}

			for _, result := range envResults {
				result.Status = environments.OperationStatus_OPERATION_STATUS_FAILED
				result.Error = &msg
			}

			resp.Results = append(resp.Results, envResults...)
			continue
		}

		resp.Results = append(resp.Results, envResults...)
		if environmentKey == req.EnvironmentKey {
			resp.Revision = envRevision
		}
	}

	if resp.Revision == "" {
		resp.Revision = req.Revision
	}

	return resp, nil
}

func applyBulkOperation(ctx context.Context, sv ResourceStore, namespaceKey string, req *environments.BulkApplyResourcesRequest) (environments.OperationStatus, bool, error) {
	switch req.Operation {
	case environments.BulkOperation_BULK_OPERATION_CREATE:
		resource := &environments.Resource{
			NamespaceKey: namespaceKey,
			Key:          req.Key,
			Payload:      req.Payload,
		}

		exists, err := resourceExists(ctx, sv, namespaceKey, req.Key)
		if err != nil {
			return environments.OperationStatus_OPERATION_STATUS_FAILED, false, err
		}

		status, wrote, err := applyCreateWithConflict(ctx, sv, resource, req.OnConflict, exists)
		return status, wrote, err
	case environments.BulkOperation_BULK_OPERATION_UPDATE:
		resource := &environments.Resource{
			NamespaceKey: namespaceKey,
			Key:          req.Key,
			Payload:      req.Payload,
		}

		if err := sv.UpdateResource(ctx, resource); err != nil {
			return environments.OperationStatus_OPERATION_STATUS_FAILED, false, err
		}
		return environments.OperationStatus_OPERATION_STATUS_SUCCESS, true, nil
	case environments.BulkOperation_BULK_OPERATION_DELETE:
		if err := sv.DeleteResource(ctx, namespaceKey, req.Key); err != nil {
			return environments.OperationStatus_OPERATION_STATUS_FAILED, false, err
		}
		return environments.OperationStatus_OPERATION_STATUS_SUCCESS, true, nil
	case environments.BulkOperation_BULK_OPERATION_UPSERT:
		resource := &environments.Resource{
			NamespaceKey: namespaceKey,
			Key:          req.Key,
			Payload:      req.Payload,
		}

		exists, err := resourceExists(ctx, sv, namespaceKey, req.Key)
		if err != nil {
			return environments.OperationStatus_OPERATION_STATUS_FAILED, false, err
		}

		if exists {
			if err := sv.UpdateResource(ctx, resource); err != nil {
				return environments.OperationStatus_OPERATION_STATUS_FAILED, false, err
			}
			return environments.OperationStatus_OPERATION_STATUS_SUCCESS, true, nil
		} else if err := sv.CreateResource(ctx, resource); err != nil {
			return environments.OperationStatus_OPERATION_STATUS_FAILED, false, err
		}
		return environments.OperationStatus_OPERATION_STATUS_SUCCESS, true, nil
	default:
		return environments.OperationStatus_OPERATION_STATUS_FAILED, false, fmt.Errorf("unsupported bulk operation: %s", req.Operation)
	}
}

func applyCreateWithConflict(
	ctx context.Context,
	sv ResourceStore,
	resource *environments.Resource,
	onConflict environments.ConflictStrategy,
	exists bool,
) (environments.OperationStatus, bool, error) {
	if exists {
		switch onConflict {
		case environments.ConflictStrategy_CONFLICT_STRATEGY_SKIP:
			return environments.OperationStatus_OPERATION_STATUS_SKIPPED, false, nil
		case environments.ConflictStrategy_CONFLICT_STRATEGY_OVERWRITE:
			if err := sv.UpdateResource(ctx, resource); err != nil {
				return environments.OperationStatus_OPERATION_STATUS_FAILED, false, err
			}
			return environments.OperationStatus_OPERATION_STATUS_SUCCESS, true, nil
		default:
			return environments.OperationStatus_OPERATION_STATUS_FAILED, false, errors.ErrAlreadyExistsf("resource %q in namespace %q already exists", resource.Key, resource.NamespaceKey)
		}
	}

	if err := sv.CreateResource(ctx, resource); err != nil {
		return environments.OperationStatus_OPERATION_STATUS_FAILED, false, err
	}

	return environments.OperationStatus_OPERATION_STATUS_SUCCESS, true, nil
}

func resourceExists(ctx context.Context, sv ResourceStoreView, namespaceKey, key string) (bool, error) {
	_, err := sv.GetResource(ctx, namespaceKey, key)
	if err == nil {
		return true, nil
	}

	if errors.AsMatch[errors.ErrNotFound](err) {
		return false, nil
	}

	return false, err
}
