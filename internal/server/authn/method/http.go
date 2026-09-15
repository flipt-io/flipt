package method

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"path"
	"strings"
	"time"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

const (
	githubPrefix       = "/auth/v1/method/github"
	oidcPrefix         = "/auth/v1/method/oidc/"
	stateCookieKey     = "flipt_client_state"
	tokenCookieKey     = "flipt_client_token"
	ForwardedPrefixKey = "x-forwarded-prefix"
	xForwardedPrefix   = "X-Forwarded-Prefix"
)

// Middleware contains various extensions for appropriate integration of the OIDC services
// behind gRPC gateway. This includes forwarding cookies as gRPC metadata, adapting callback
// responses to http cookies, and establishing appropriate state parameters for csrf provention
// during the oauth/oidc flow.
type Middleware struct {
	config config.AuthenticationSessionConfig
}

// NewHTTPMiddleware constructs and configures a new oidc HTTP middleware from the supplied
// authentication configuration struct.
func NewHTTPMiddleware(config config.AuthenticationSessionConfig) Middleware {
	return Middleware{
		config: config,
	}
}

// ForwardCookies parses particular http cookies (Flipts state and client token) and
// forwards them as grpc metadata entries. This allows us to abstract away http
// constructs from the internal gRPC implementation.
func ForwardCookies(ctx context.Context, req *http.Request) metadata.MD {
	md := metadata.MD{}
	for _, key := range []string{stateCookieKey, tokenCookieKey} {
		if cookie, err := req.Cookie(key); err == nil {
			md[key] = []string{cookie.Value}
		}
	}

	return md
}

// ForwardPrefix extracts the "X-Forwarded-Prefix" header from an HTTP request
// and forwards them as grpc metadata entries.
func ForwardPrefix(ctx context.Context, req *http.Request) metadata.MD {
	md := metadata.MD{}
	values := req.Header.Values(xForwardedPrefix)
	if len(values) > 0 {
		md[ForwardedPrefixKey] = values
	}
	return md
}

// ForwardResponseOption is a grpc gateway forward response option function implementation.
// The purpose of which is to intercept outgoing Callback operation responses.
// When intercepted the resulting clientToken is stripped from the response payload and instead
// added to a response header cookie (Set-Cookie).
// This ensures a secure browser session can be established.
// The user-agent is then redirected to the original UI location captured in
// original_state (as "/#<original_state>" for the hash router),
// falling back to the root of the domain when absent or invalid.
func (m Middleware) ForwardResponseOption(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	r, ok := resp.(*auth.CallbackResponse)
	if ok {
		cookie := &http.Cookie{
			Name:     tokenCookieKey,
			Value:    r.GetClientToken(),
			Domain:   m.config.Domain,
			Path:     "/",
			Expires:  time.Now().Add(m.config.TokenLifetime),
			Secure:   m.config.Secure,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		}

		http.SetCookie(w, cookie)

		// clear out token now that it is set via cookie
		r.ClientToken = ""
		w.Header().Set("Location", resolveCallbackLocation(ctx))
		w.WriteHeader(http.StatusFound)
	}

	return nil
}

// clientStatePayload mirrors the JSON object written into the state
// parameter and state cookie by Handler.
type clientStatePayload struct {
	SecurityToken string `json:"security_token"`
	OriginalState string `json:"original_state"`
}

// resolveCallbackLocation returns the redirect target for a successful
// callback: "/#<original_state>" when a safe relative original_state can be
// decoded from the flipt_client_state metadata, otherwise the (possibly
// prefix-scoped) root path.
func resolveCallbackLocation(ctx context.Context) string {
	base := forwardedBasePath(ctx)

	original, ok := originalStateFromContext(ctx)
	if !ok {
		return base
	}

	// The UI uses a hash router, so the server redirect must include the
	// "#". e.g. original_state "/flags" -> "/#/flags".
	return strings.TrimSuffix(base, "/") + "/#" + original
}

// forwardedBasePath returns "/" or "<prefix>/" when an X-Forwarded-Prefix
// is present in the request metadata.
func forwardedBasePath(ctx context.Context) string {
	for _, mdGetter := range []func(context.Context) (metadata.MD, bool){
		metadata.FromOutgoingContext,
		metadata.FromIncomingContext,
	} {
		if md, ok := mdGetter(ctx); ok {
			if prefixes := md.Get(ForwardedPrefixKey); len(prefixes) > 0 {
				return path.Join(prefixes...) + "/"
			}
		}
	}

	return "/"
}

// originalStateFromContext decodes and validates the original_state value
// wrapped in the flipt_client_state metadata entry.
func originalStateFromContext(ctx context.Context) (string, bool) {
	var encoded string

	for _, mdGetter := range []func(context.Context) (metadata.MD, bool){
		metadata.FromIncomingContext,
		metadata.FromOutgoingContext,
	} {
		if md, ok := mdGetter(ctx); ok {
			if values := md.Get(stateCookieKey); len(values) > 0 && values[0] != "" {
				encoded = values[0]
				break
			}
		}
	}

	if encoded == "" {
		return "", false
	}

	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", false
	}

	var payload clientStatePayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return "", false
	}

	return sanitizeOriginalState(payload.OriginalState)
}

// sanitizeOriginalState ensures the redirect stays same-origin: only relative
// paths are allowed. It returns the normalized path and whether it is usable.
func sanitizeOriginalState(s string) (string, bool) {
	const maxOriginalStateLen = 2048

	s = strings.TrimSpace(s)
	if s == "" || len(s) > maxOriginalStateLen {
		return "", false
	}

	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}

	if strings.HasPrefix(s, "//") || strings.Contains(s, "\\") || strings.Contains(s, "://") {
		return "", false
	}

	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\x00' {
			return "", false
		}
	}

	return s, true
}

// Handler is a http middleware used to decorate the OIDC provider gateway handler.
// The middleware intercepts authorize attempts and automatically establishes an
// appropriate state parameter. It does so by wrapping any provided state parameter
// in a JSON object with an additional cryptographically-random generated security
// token. The payload is then encoded in base64 and added back to the state query param.
// The payload is then also encoded as a http cookie which is bound to the callback path.
func (m Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefix, method := path.Split(r.URL.Path)

		//nolint:staticcheck
		if !((strings.HasPrefix(prefix, oidcPrefix) || strings.HasPrefix(prefix, githubPrefix)) && method == "authorize") {
			next.ServeHTTP(w, r)
			return
		}

		if method == "authorize" {
			prefix = path.Join(path.Join(r.Header.Values(xForwardedPrefix)...), prefix)
			query := r.URL.Query()
			// create a random security token and bind it to
			// the state parameter while preserving any provided
			// state
			v, err := json.Marshal(clientStatePayload{
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

			cookie := &http.Cookie{
				Name:  stateCookieKey,
				Value: encoded,
				// bind state cookie to provider callback
				Path:     prefix + "/callback",
				Expires:  time.Now().Add(m.config.StateLifetime),
				Secure:   m.config.Secure,
				HttpOnly: true,
				// we need to support cookie forwarding when user
				// is being navigated from authorizing server
				SameSite: http.SameSiteLaxMode,
			}

			// domains must have at least two dots to be considered valid, so we
			// `localhost` is not a valid domain. See:
			// https://curl.se/rfc/cookie_spec.html
			if !strings.HasPrefix(m.config.Domain, "localhost") {
				cookie.Domain = m.config.Domain
			}

			http.SetCookie(w, cookie)
		}

		// run decorated handler
		next.ServeHTTP(w, r)
	})
}

func generateSecurityToken() string {
	var token [64]byte
	if _, err := rand.Read(token[:]); err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(token[:])
}
