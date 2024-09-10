package oidc_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/cap/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	authoidc "go.flipt.io/flipt/internal/server/authn/method/oidc"
	oidctesting "go.flipt.io/flipt/internal/server/authn/method/oidc/testing"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"golang.org/x/net/html"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
)

func Test_Server_ImplicitFlow(t *testing.T) {
	var (
		router = chi.NewRouter()
		// httpServer is the test server used for hosting
		// Flipts oidc authorize and callback handles
		httpServer = httptest.NewServer(router)
		// rewriting http server to use localhost as it is a domain and
		// the <=go1.18 implementation will propagate cookies on it.
		// From go1.19+ cookiejar support IP addresses as cookie domains.
		clientAddress = strings.Replace(httpServer.URL, "127.0.0.1", "localhost", 1)

		id, secret = "client_id", "client_secret"
		nonce      = "static"

		logger = zaptest.NewLogger(t)
		ctx    = context.Background()
	)

	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n\n", err)
		return
	}

	tp := oidc.StartTestProvider(t, oidc.WithNoTLS(), oidc.WithTestDefaults(&oidc.TestProviderDefaults{
		CustomClaims: map[string]interface{}{},
		SubjectInfo: map[string]*oidc.TestSubject{
			"mark": {
				Password: "phelps",
				UserInfo: map[string]interface{}{
					"email": "mark@flipt.io",
					"name":  "Mark Phelps",
				},
				CustomClaims: map[string]interface{}{
					"email": "mark@flipt.io",
					"name":  "Mark Phelps",
					"roles": []string{"admin"},
				},
			},
			"george": {
				Password: "macrorie",
				UserInfo: map[string]interface{}{
					"email": "george@flipt.io",
					"name":  "George MacRorie",
				},
				CustomClaims: map[string]interface{}{
					"email": "george@flipt.io",
					"name":  "George MacRorie",
					"roles": []string{"editor"},
				},
			},
		},
		SigningKey: &oidc.TestSigningKey{
			PrivKey: priv,
			PubKey:  priv.Public(),
			Alg:     oidc.RS256,
		},
		AllowedRedirectURIs: []string{
			fmt.Sprintf("%s/auth/v1/method/oidc/google/callback", clientAddress),
		},
		ClientID:      &id,
		ClientSecret:  &secret,
		ExpectedNonce: &nonce,
	}))

	defer tp.Stop()

	var (
		authConfig = config.AuthenticationConfig{
			Session: config.AuthenticationSession{
				Domain:        "localhost",
				Secure:        false,
				TokenLifetime: 1 * time.Hour,
				StateLifetime: 10 * time.Minute,
			},
			Methods: config.AuthenticationMethods{
				OIDC: config.AuthenticationMethod[config.AuthenticationMethodOIDCConfig]{
					Enabled: true,
					Method: config.AuthenticationMethodOIDCConfig{
						Providers: map[string]config.AuthenticationMethodOIDCProvider{
							"google": {
								IssuerURL:       tp.Addr(),
								ClientID:        id,
								ClientSecret:    secret,
								RedirectAddress: clientAddress,
								Nonce:           nonce,
							},
						},
					},
				},
			},
		}
		server = oidctesting.StartHTTPServer(t, ctx, logger, authConfig, router)
	)

	t.Cleanup(func() { _ = server.Stop() })

	jar, err := cookiejar.New(&cookiejar.Options{})
	require.NoError(t, err)

	client := &http.Client{
		// skip redirects
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
		// establish a cookie jar
		Jar: jar,
	}

	testOIDCFlow(t, ctx, tp.Addr(), clientAddress, client, server)
}

func Test_Server_PKCE(t *testing.T) {
	var (
		router        = chi.NewRouter()
		httpServer    = httptest.NewServer(router)
		clientAddress = strings.Replace(httpServer.URL, "127.0.0.1", "localhost", 1)

		id, secret = "client_id", "client_secret"
		nonce      = "static"

		logger = zaptest.NewLogger(t)
		ctx    = context.Background()
	)

	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n\n", err)
		return
	}

	tp := oidc.StartTestProvider(t, oidc.WithNoTLS(), oidc.WithTestDefaults(&oidc.TestProviderDefaults{
		CustomClaims: map[string]interface{}{},
		SubjectInfo: map[string]*oidc.TestSubject{
			"mark": {
				Password: "phelps",
				UserInfo: map[string]interface{}{
					"email": "mark@flipt.io",
					"name":  "Mark Phelps",
				},
				CustomClaims: map[string]interface{}{
					"email": "mark@flipt.io",
					"name":  "Mark Phelps",
					"roles": []string{"admin"},
				},
			},
			"george": {
				Password: "macrorie",
				UserInfo: map[string]interface{}{
					"email": "george@flipt.io",
					"name":  "George MacRorie",
				},
				CustomClaims: map[string]interface{}{
					"email": "george@flipt.io",
					"name":  "George MacRorie",
					"roles": []string{"editor"},
				},
			},
		},
		SigningKey: &oidc.TestSigningKey{
			PrivKey: priv,
			PubKey:  priv.Public(),
			Alg:     oidc.RS256,
		},
		AllowedRedirectURIs: []string{
			fmt.Sprintf("%s/auth/v1/method/oidc/google/callback", clientAddress),
		},
		ClientID:      &id,
		ClientSecret:  &secret,
		ExpectedNonce: &nonce,
		PKCEVerifier:  authoidc.PKCEVerifier,
	}))

	defer tp.Stop()

	var (
		authConfig = config.AuthenticationConfig{
			Session: config.AuthenticationSession{
				Domain:        "localhost",
				Secure:        false,
				TokenLifetime: 1 * time.Hour,
				StateLifetime: 10 * time.Minute,
			},
			Methods: config.AuthenticationMethods{
				OIDC: config.AuthenticationMethod[config.AuthenticationMethodOIDCConfig]{
					Enabled: true,
					Method: config.AuthenticationMethodOIDCConfig{
						Providers: map[string]config.AuthenticationMethodOIDCProvider{
							"google": {
								IssuerURL:       tp.Addr(),
								ClientID:        id,
								ClientSecret:    secret,
								RedirectAddress: clientAddress,
								Nonce:           nonce,
								UsePKCE:         true,
							},
						},
					},
				},
			},
		}
		server = oidctesting.StartHTTPServer(t, ctx, logger, authConfig, router)
	)

	t.Cleanup(func() { _ = server.Stop() })

	jar, err := cookiejar.New(&cookiejar.Options{})
	require.NoError(t, err)

	client := &http.Client{
		// skip redirects
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
		// establish a cookie jar
		Jar: jar,
	}

	testOIDCFlow(t, ctx, tp.Addr(), clientAddress, client, server)
}

func Test_Server_Nonce(t *testing.T) {
	var (
		router        = chi.NewRouter()
		httpServer    = httptest.NewServer(router)
		clientAddress = strings.Replace(httpServer.URL, "127.0.0.1", "localhost", 1)

		id, secret = "client_id", "client_secret"
		nonce      = "random-nonce"

		logger = zaptest.NewLogger(t)
		ctx    = context.Background()
	)

	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)

	tp := oidc.StartTestProvider(t, oidc.WithNoTLS(), oidc.WithTestDefaults(&oidc.TestProviderDefaults{
		CustomClaims: map[string]interface{}{},
		SubjectInfo: map[string]*oidc.TestSubject{
			"mark": {
				Password: "phelps",
				UserInfo: map[string]interface{}{
					"email": "mark@flipt.io",
					"name":  "Mark Phelps",
				},
				CustomClaims: map[string]interface{}{
					"email": "mark@flipt.io",
					"name":  "Mark Phelps",
					"roles": []string{"admin"},
				},
			},
			"george": {
				Password: "macrorie",
				UserInfo: map[string]interface{}{
					"email": "george@flipt.io",
					"name":  "George MacRorie",
				},
				CustomClaims: map[string]interface{}{
					"email": "george@flipt.io",
					"name":  "George MacRorie",
					"roles": []string{"editor"},
				},
			},
		},
		SigningKey: &oidc.TestSigningKey{
			PrivKey: priv,
			PubKey:  priv.Public(),
			Alg:     oidc.RS256,
		},
		AllowedRedirectURIs: []string{
			fmt.Sprintf("%s/auth/v1/method/oidc/google/callback", clientAddress),
		},
		ClientID:      &id,
		ClientSecret:  &secret,
		ExpectedNonce: &nonce,
	}))

	defer tp.Stop()

	var (
		authConfig = config.AuthenticationConfig{
			Session: config.AuthenticationSession{
				Domain:        "localhost",
				Secure:        false,
				TokenLifetime: 1 * time.Hour,
				StateLifetime: 10 * time.Minute,
			},
			Methods: config.AuthenticationMethods{
				OIDC: config.AuthenticationMethod[config.AuthenticationMethodOIDCConfig]{
					Enabled: true,
					Method: config.AuthenticationMethodOIDCConfig{
						Providers: map[string]config.AuthenticationMethodOIDCProvider{
							"google": {
								IssuerURL:       tp.Addr(),
								ClientID:        id,
								ClientSecret:    secret,
								RedirectAddress: clientAddress,
								Nonce:           nonce,
							},
						},
					},
				},
			},
		}
		server = oidctesting.StartHTTPServer(t, ctx, logger, authConfig, router)
	)

	t.Cleanup(func() { _ = server.Stop() })

	jar, err := cookiejar.New(&cookiejar.Options{})
	require.NoError(t, err)

	client := &http.Client{
		// skip redirects
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
		// establish a cookie jar
		Jar: jar,
	}

	testOIDCFlow(t, ctx, tp.Addr(), clientAddress, client, server)
}

func testOIDCFlow(t *testing.T, ctx context.Context, tpAddr, clientAddress string, client *http.Client, server *oidctesting.HTTPServer) {
	var authURL *url.URL

	t.Run("AuthorizeURL", func(t *testing.T) {
		authorizeURL := clientAddress + "/auth/v1/method/oidc/google/authorize"

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, authorizeURL, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var authorize auth.AuthorizeURLResponse
		err = protojson.Unmarshal(body, &authorize)
		require.NoError(t, err)

		authURL, err = url.Parse(authorize.AuthorizeUrl)
		require.NoError(t, err)
	})

	t.Log("Navigating to authorize URL:", authURL.String())

	var location string
	t.Run("Login as Mark", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, authURL.String(), nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		values, err := parseLoginFormHiddenValues(resp.Body)
		require.NoError(t, err)

		// add login credentials
		values.Set("uname", "mark")
		values.Set("psw", "phelps")

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tpAddr+"/login", strings.NewReader(values.Encode()))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusFound, resp.StatusCode)

		location = resp.Header.Get("Location")
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	require.NoError(t, err)

	t.Run("Callback (missing state)", func(t *testing.T) {
		// using the default client which has no state cookie
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Callback (invalid state)", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
		require.NoError(t, err)

		req.Header.Set("Cookie", "flipt_client_state=abcdef")
		// using the default client which has no state cookie
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Callback", func(t *testing.T) {
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var response auth.CallbackResponse
		if !assert.NoError(t, protojson.Unmarshal(data, &response)) {
			t.Log("Unexpected response", string(data))
			t.FailNow()
		}

		assert.Empty(t, response.ClientToken) // middleware moves it to cookie
		assert.Equal(t, auth.Method_METHOD_OIDC, response.Authentication.Method)

		claims := response.Authentication.Metadata["io.flipt.auth.claims"]
		delete(response.Authentication.Metadata, "io.flipt.auth.claims")

		assert.Equal(t, map[string]string{
			"io.flipt.auth.oidc.provider": "google",
			"io.flipt.auth.oidc.email":    "mark@flipt.io",
			"io.flipt.auth.email":         "mark@flipt.io",
			"io.flipt.auth.oidc.name":     "Mark Phelps",
			"io.flipt.auth.name":          "Mark Phelps",
			"io.flipt.auth.oidc.sub":      "mark",
		}, response.Authentication.Metadata)

		assert.NotEmpty(t, claims)

		// ensure expiry is set
		assert.NotNil(t, response.Authentication.ExpiresAt)

		// obtain returned cookie
		cookie, err := (&http.Request{
			Header: http.Header{"Cookie": resp.Header["Set-Cookie"]},
		}).Cookie("flipt_client_token")
		require.NoError(t, err)

		// check authentication in store matches
		storedAuth, err := server.GRPCServer.Store.GetAuthenticationByClientToken(ctx, cookie.Value)
		require.NoError(t, err)

		// ensure stored auth can be retrieved by cookie abd matches response body auth
		if diff := cmp.Diff(storedAuth, response.Authentication, protocmp.Transform()); err != nil {
			t.Errorf("-exp/+got:\n%s", diff)
		}
	})
}

// parseLoginFormHiddenValues parses the contents of the supplied reader as HTML.
// It descends into the document looking for the hidden values associated
// with the login form.
// It collecs the hidden values into a url.Values so that we can post the form
// using our Go client.
func parseLoginFormHiddenValues(r io.Reader) (url.Values, error) {
	values := url.Values{}

	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	findNode := func(visit func(*html.Node), name string, attrs ...html.Attribute) func(*html.Node) {
		var f func(*html.Node)
		f = func(n *html.Node) {
			var hasAttrs bool
			if n.Type == html.ElementNode && n.Data == name {
				hasAttrs = true
				for _, want := range attrs {
					for _, has := range n.Attr {
						if has.Key == want.Key {
							hasAttrs = hasAttrs && (want.Val == has.Val)
						}
					}
				}
			}

			if hasAttrs {
				visit(n)
				return
			}

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}

		return f
	}

	findNode(func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findNode(func(n *html.Node) {
				var (
					name  string
					value string
				)
				for _, a := range n.Attr {
					switch a.Key {
					case "name":
						name = a.Val
					case "value":
						value = a.Val
					}
				}

				values.Set(name, value)
			}, "input", html.Attribute{Key: "type", Val: "hidden"})(c)
		}
	}, "form", html.Attribute{Key: "action", Val: "/login"})(doc)

	return values, nil
}
