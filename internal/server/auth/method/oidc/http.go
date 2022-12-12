package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

var (
	stateCookieKey = "flipt_client_state"
	tokenCookieKey = "flipt_client_token"
)

// Middleware contains various extensions for appropriate integration of the OIDC services
// behind gRPC gateway. This includes forwarding cookies as gRPC metadata, adapting callback
// responses to http cookies, and establishing appropriate state parameters for csrf provention
// during the oauth/oidc flow.
type Middleware struct {
	Config config.AuthenticationSession
}

// NewHTTPMiddleware constructs and configures a new oidc HTTP middleware from the supplied
// authentication configuration struct.
func NewHTTPMiddleware(config config.AuthenticationSession) Middleware {
	return Middleware{
		Config: config,
	}
}

// ForwardCookies parses particular http cookies (Flipts state and client token) and
// forwards them as grpc metadata entries. This allows us to abstract away http
// constructs from the internal gRPC implementation.
func ForwardCookies(ctx context.Context, req *http.Request) metadata.MD {
	md := metadata.MD{}
	for _, key := range []string{stateCookieKey, tokenCookieKey} {
		if cookie, err := req.Cookie(key); err == nil {
			md[stateCookieKey] = []string{cookie.Value}
		}
	}

	return md
}

// ForwardResponseOption is a grpc gateway forward response option function implementation.
// The purpose of which is to intercept outgoing Callback operation responses.
// When intercepted the resulting clientToken is stripped from the response payload and instead
// added to a response header cookie (Set-Cookie).
// This ensures a secure browser session can be established.
// The user-agent is then redirected to the root of the domain.
func (m Middleware) ForwardResponseOption(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	r, ok := resp.(*auth.CallbackResponse)
	if ok {
		cookie := &http.Cookie{
			Name:     tokenCookieKey,
			Value:    r.ClientToken,
			Domain:   m.Config.Domain,
			Path:     "/",
			Expires:  time.Now().Add(m.Config.TokenLifetime),
			Secure:   m.Config.Secure,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		}

		http.SetCookie(w, cookie)

		// clear out token now that it is set via cookie
		r.ClientToken = ""

		w.Header().Set("Location", "/")
		w.WriteHeader(http.StatusFound)
	}

	return nil
}

// Handler is a http middleware used to decorate the OIDC provider gateway handler.
// The middleware intercepts authorize attempts and automatically establishes an
// appropriate state parameter. It does so by wrapping any provided state parameter
// in a JSON object with an additional cryptographically-random generated security
// token. The payload is then encoded in base64 and added back to the state query param.
// The payload is then also encoded as a http cookie which is bound to the callback path.
func (m Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		provider, method, match := parts(r.URL.Path)
		if !match {
			next.ServeHTTP(w, r)
			return
		}

		// rewrite URL paths for friendlier URL support
		enumName := "OIDC_PROVIDER_" + strings.ToUpper(provider)
		if _, ok := auth.OIDCProvider_value[enumName]; ok {
			r.URL.Path = fmt.Sprintf("/auth/v1/method/oidc/%s/%s", enumName, method)
		}

		if method == "authorize" {
			query := r.URL.Query()
			// create a random security token and bind it to
			// the state parameter while preserving any provided
			// state
			v, err := json.Marshal(struct {
				SecurityToken string `json:"security_token"`
				OriginalState string `json:"original_state"`
			}{
				// TODO(georgemac): handle redirect URL
				SecurityToken: generateSecurityToken(),
				// preserve and forward state
				OriginalState: query.Get("state"),
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// json marshal and base64 encode payload to url-safe string
			encoded := base64.URLEncoding.EncodeToString(v)

			// replace state parameter with generated value
			query.Set("state", encoded)
			r.URL.RawQuery = query.Encode()

			http.SetCookie(w, &http.Cookie{
				Name:   stateCookieKey,
				Value:  encoded,
				Domain: m.Config.Domain,
				// bind state cookie to provider callback
				Path:     "/auth/v1/method/oidc/" + provider + "/callback",
				Expires:  time.Now().Add(m.Config.StateLifetime),
				Secure:   m.Config.Secure,
				HttpOnly: true,
				// we need to support cookie forwarding when user
				// is being navigated from authorizing server
				SameSite: http.SameSiteLaxMode,
			})
		}

		// run decorated handler
		next.ServeHTTP(w, r)
	})
}

func parts(path string) (provider, method string, ok bool) {
	const prefix = "/auth/v1/method/oidc/"
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}

	return strings.Cut(path[len(prefix):], "/")
}

func generateSecurityToken() string {
	var token [64]byte
	if _, err := rand.Read(token[:]); err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(token[:])
}
