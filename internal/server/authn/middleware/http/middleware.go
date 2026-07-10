package http_middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.flipt.io/flipt/internal/config"
	middlewarecommon "go.flipt.io/flipt/internal/server/authn/middleware/common"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const stateCookieKey = "flipt_client_state"

// Middleware contains various extensions for appropriate integration of the generic auth services
// behind gRPC gateway. This currently includes clearing the appropriate cookies on logout.
type Middleware struct {
	config            config.AuthenticationSessionConfig
	defaultErrHandler runtime.ErrorHandlerFunc
}

// NewHTTPMiddleware constructs a new auth HTTP middleware.
func NewHTTPMiddleware(config config.AuthenticationSessionConfig) *Middleware {
	return &Middleware{
		config:            config,
		defaultErrHandler: runtime.DefaultHTTPErrorHandler,
	}
}

// Handler is a http middleware used to decorate the auth provider gateway handler.
// This is used to clear the appropriate cookies on logout.
func (m Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/auth/v1/self/expire" {
			m.clearAllCookies(w, http.SameSiteStrictMode)
		} else if r.Method == http.MethodDelete && r.URL.Path == "/auth/v1/self/revoke" {
			m.clearAllCookies(w, http.SameSiteStrictMode)
		}
		next.ServeHTTP(w, r)
	})
}

// ForwardRevokeOIDCResponseOption is a grpc-gateway forward response option function.
// When the response is a successful RevokeOIDCResponse for a GET request
// (OIDC front-channel logout), it clears session cookies so the browser
// discards them.
func (m Middleware) ForwardRevokeOIDCResponseOption(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	if _, ok := resp.(*auth.RevokeOIDCResponse); !ok {
		return nil
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md, ok = metadata.FromIncomingContext(ctx)
	}
	if !ok {
		return nil
	}

	methods := md.Get("x-http-method")
	if len(methods) != 1 || methods[0] != http.MethodGet {
		return nil
	}

	w.Header().Set("Cache-Control", "no-store")
	m.clearAllCookies(w, http.SameSiteNoneMode)
	return nil
}

// ErrorHandler ensures cookies are cleared when cookie auth is attempted but leads to
// an unauthenticated response. This ensures well behaved user-agents won't attempt to
// supply the same token via a cookie again in a subsequent call.
func (m Middleware) ErrorHandler(ctx context.Context, sm *runtime.ServeMux, ms runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	// given a token cookie was supplied and the resulting error was unauthenticated
	// then we clear all cookies to instruct the user agent to not attempt to use them
	// again in a subsequent call
	if _, cerr := r.Cookie(middlewarecommon.TokenCookieKey); status.Code(err) == codes.Unauthenticated &&
		!errors.Is(cerr, http.ErrNoCookie) {
		m.clearAllCookies(w, http.SameSiteStrictMode)
	}

	// always delegate to default handler
	m.defaultErrHandler(ctx, sm, ms, w, r, err)
}

// clearAllCookies clears both the state and token cookies for the session domain.
// The sameSite parameter controls the cookie SameSite attribute:
//   - SameSiteNoneMode is used for OIDC front-channel logout because the request
//     arrives as a third-party cross-site load in a hidden iframe/img from the OP;
//     Lax/Strict would prevent the cookies from being attached and thus cleared.
//   - SameSiteStrictMode is used for direct user-initiated expire/revoke requests
//     where the caller is on the same site.
func (m Middleware) clearAllCookies(w http.ResponseWriter, sameSite http.SameSite) {
	for _, cookieName := range []string{stateCookieKey, middlewarecommon.TokenCookieKey} {
		cookie := &http.Cookie{
			Name:     cookieName,
			Value:    "",
			Domain:   m.config.Domain,
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   m.config.Secure,
			SameSite: sameSite,
		}

		http.SetCookie(w, cookie)
	}
}
