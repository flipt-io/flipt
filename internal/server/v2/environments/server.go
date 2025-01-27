package environments

import (
	"context"
	"fmt"
	"net/url"

	"go.flipt.io/flipt/errors"
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

func (s *Server) GetNamespace(ctx context.Context, req *environments.GetNamespaceRequest) (ns *environments.NamespaceResponse, err error) {
	env, err := s.envs.Get(ctx, req.Environment)
	if err != nil {
		return nil, err
	}

	return env.GetNamespace(ctx, req.Key)
}

func (s *Server) ListNamespaces(ctx context.Context, req *environments.ListNamespacesRequest) (nl *environments.ListNamespacesResponse, err error) {
	env, err := s.envs.Get(ctx, req.Environment)
	if err != nil {
		return nil, err
	}

	return env.ListNamespaces(ctx)
}

func (s *Server) CreateNamespace(ctx context.Context, ns *environments.UpdateNamespaceRequest) (*environments.NamespaceResponse, error) {
	env, err := s.envs.Get(ctx, ns.Environment)
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
	env, err := s.envs.Get(ctx, ns.Environment)
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
	env, err := s.envs.Get(ctx, req.Environment)
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
	env, err := s.envs.Get(ctx, req.Environment)
	if err != nil {
		return nil, err
	}

	typ, err := ParseResourceType(req.TypeUrl)
	if err != nil {
		return nil, err
	}

	return r, env.View(ctx, typ, func(ctx context.Context, sv ResourceStoreView) (err error) {
		r, err = sv.GetResource(ctx, req.Namespace, req.Key)
		if err != nil {
			return err
		}
		return err
	})
}

func (s *Server) ListResources(ctx context.Context, req *environments.ListResourcesRequest) (rl *environments.ListResourcesResponse, err error) {
	env, err := s.envs.Get(ctx, req.Environment)
	if err != nil {
		return nil, err
	}

	typ, err := ParseResourceType(req.TypeUrl)
	if err != nil {
		return nil, err
	}

	if err := env.View(ctx, typ, func(ctx context.Context, sv ResourceStoreView) (err error) {
		rl, err = sv.ListResources(ctx, req.Namespace)
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
	env, err := s.envs.Get(ctx, req.Environment)
	if err != nil {
		return nil, err
	}

	resp := &environments.ResourceResponse{
		Resource: &environments.Resource{
			Environment: req.Environment,
			Namespace:   req.Namespace,
			Key:         req.Key,
			Payload:     req.Payload,
		},
	}

	typ, err := ParseResourceType(req.Payload.GetTypeUrl())
	if err != nil {
		return nil, err
	}

	resp.Revision, err = env.Update(ctx, req.Revision, typ, func(ctx context.Context, sv ResourceStore) error {
		_, err := sv.GetResource(ctx, req.Namespace, req.Key)
		if err == nil {
			return errors.ErrAlreadyExistsf(`create resource "%s/%s/%s"`, req.Payload.GetTypeUrl(), req.Namespace, req.Key)
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
	env, err := s.envs.Get(ctx, req.Environment)
	if err != nil {
		return nil, err
	}

	resp := &environments.ResourceResponse{
		Resource: &environments.Resource{
			Environment: req.Environment,
			Namespace:   req.Namespace,
			Key:         req.Key,
			Payload:     req.Payload,
		},
	}

	typ, err := ParseResourceType(req.GetTypeUrl())
	if err != nil {
		return nil, err
	}

	resp.Revision, err = env.Update(ctx, req.Revision, typ, func(ctx context.Context, sv ResourceStore) error {
		_, err := sv.GetResource(ctx, req.Namespace, req.Key)
		if err != nil {
			return fmt.Errorf(`update resource "%s/%s/%s": %w`, req.Payload.GetTypeUrl(), req.Namespace, req.Key, err)
		}

		return sv.UpdateResource(ctx, resp.Resource)
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *Server) DeleteResource(ctx context.Context, req *environments.DeleteResourceRequest) (_ *environments.DeleteResourceResponse, err error) {
	env, err := s.envs.Get(ctx, req.Environment)
	if err != nil {
		return nil, err
	}

	typ, err := ParseResourceType(req.TypeUrl)
	if err != nil {
		return nil, err
	}

	resp := &environments.DeleteResourceResponse{}
	if resp.Revision, err = env.Update(ctx, req.Revision, typ, func(ctx context.Context, sv ResourceStore) (err error) {
		return sv.DeleteResource(ctx, req.Namespace, req.Key)
	}); err != nil {
		return nil, err
	}

	return resp, nil
}

// ensureFullURL ensures that the provided host is a valid URL with a scheme and host.
// If the host is already a valid URL, it returns it as is.
// If the host is not a valid URL, it returns a URL with the provided host and https as the scheme.
func ensureFullURL(envHost string) string {
	u, err := url.Parse(envHost)
	if err == nil && u.Host != "" {
		return envHost
	}
	if u.Host == "" {
		u = &url.URL{
			Scheme: "https",
			Host:   envHost,
		}
	}

	return u.String()
}
