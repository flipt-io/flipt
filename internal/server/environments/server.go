package environments

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/authz"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var _ environments.EnvironmentsServiceServer = (*Server)(nil)

type Server struct {
	logger *zap.Logger

	envs *EnvironmentStore

	environments.UnimplementedEnvironmentsServiceServer
}

func NewServer(logger *zap.Logger, envs *EnvironmentStore) (_ *Server, err error) {
	return &Server{logger: logger, envs: envs}, nil
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

		if cfg.Base != nil {
			// ignore branched environments
			continue
		}

		el.Environments = append(el.Environments, &environments.Environment{
			Key:           env.Key(),
			Name:          env.Key(),
			Default:       ptr(env.Default()),
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
	env, err := s.envs.Branch(ctx, req.BaseEnvironmentKey)
	if err != nil {
		return nil, err
	}

	return &environments.Environment{
		Key:           env.Key(),
		Name:          env.Key(),
		Default:       ptr(env.Default()),
		Configuration: env.Configuration(),
	}, nil
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

func ptr[T any](t T) *T {
	return &t
}
