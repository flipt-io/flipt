package testing

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"math/big"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TestSubject struct {
	Password     string
	CustomClaims map[string]any
	UserInfo     map[string]any
}

type TestSigningKey struct {
	PrivKey *rsa.PrivateKey
	PubKey  crypto.PublicKey
	Alg     string
}

type TestProviderDefaults struct {
	CustomClaims        map[string]any
	SubjectInfo         map[string]*TestSubject
	SigningKey          *TestSigningKey
	AllowedRedirectURIs []string
	ClientID            *string
	ClientSecret        *string
	ExpectedNonce       *string
	PKCEVerifier        string
}

type TestProvider struct {
	server *httptest.Server
	config TestProviderDefaults
	addr   string

	mu       sync.Mutex
	codes    map[string]authCodeEntry
	sessions map[string]sessionEntry
	keyID    string
}

type authCodeEntry struct {
	subject  string
	redirect string
	nonce    string
	claims   map[string]any
}

type sessionEntry struct {
	subject string
	claims  map[string]any
}

func StartTestProvider(t *testing.T, opts ...TestProviderOption) *TestProvider {
	t.Helper()

	config := TestProviderDefaults{
		CustomClaims: map[string]any{},
		SubjectInfo:  map[string]*TestSubject{},
	}

	for _, opt := range opts {
		opt(&config)
	}

	if config.SigningKey == nil {
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err)
		}
		config.SigningKey = &TestSigningKey{
			PrivKey: priv,
			PubKey:  priv.Public(),
			Alg:     "RS256",
		}
	}

	if config.ClientID == nil {
		id := "test-client"
		config.ClientID = &id
	}

	tp := &TestProvider{
		config:   config,
		codes:    make(map[string]authCodeEntry),
		sessions: make(map[string]sessionEntry),
		keyID:    fmt.Sprintf("key-%d", time.Now().UnixNano()),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", tp.handleDiscovery)
	mux.HandleFunc("/.well-known/jwks.json", tp.handleJWKS)
	mux.HandleFunc("/authorize", tp.handleAuthorize)
	mux.HandleFunc("/login", tp.handleLogin)
	mux.HandleFunc("/token", tp.handleToken)
	mux.HandleFunc("/userinfo", tp.handleUserInfo)

	tp.server = httptest.NewUnstartedServer(mux)
	tp.server.Start()
	tp.addr = tp.server.URL

	t.Cleanup(tp.Stop)

	return tp
}

func (tp *TestProvider) Addr() string {
	return tp.addr
}

func (tp *TestProvider) Stop() {
	tp.server.Close()
}

func (tp *TestProvider) SetSigningKeys(priv *rsa.PrivateKey, pub crypto.PublicKey, alg, keyID string) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.config.SigningKey = &TestSigningKey{
		PrivKey: priv,
		PubKey:  pub,
		Alg:     alg,
	}
	tp.keyID = keyID
}

func (tp *TestProvider) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	baseURL := tp.addr
	reply := map[string]any{
		"issuer":                                baseURL,
		"authorization_endpoint":                baseURL + "/authorize",
		"token_endpoint":                        baseURL + "/token",
		"userinfo_endpoint":                     baseURL + "/userinfo",
		"jwks_uri":                              baseURL + "/.well-known/jwks.json",
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{tp.config.SigningKey.Alg},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(reply) //nolint:errchkjson
}

func (tp *TestProvider) handleJWKS(w http.ResponseWriter, r *http.Request) {
	tp.mu.Lock()
	key := tp.config.SigningKey
	keyID := tp.keyID
	tp.mu.Unlock()

	rsaPub, ok := key.PubKey.(*rsa.PublicKey)
	if !ok {
		http.Error(w, "not an RSA key", http.StatusInternalServerError)
		return
	}

	n := base64.RawURLEncoding.EncodeToString(rsaPub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(rsaPub.E)).Bytes())

	keys := []map[string]any{
		{
			"kty": "RSA",
			"alg": key.Alg,
			"use": "sig",
			"kid": keyID,
			"n":   n,
			"e":   e,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"keys": keys}) //nolint:errchkjson
}

func (tp *TestProvider) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tp.renderLoginPage(w, r)
		return
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (tp *TestProvider) renderLoginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<body>
<form action="/login" method="POST">
`)
	for key, values := range r.URL.Query() {
		for _, v := range values {
			fmt.Fprintf(w, `<input type="hidden" name="%s" value="%s" />`, key, v)
		}
	}
	fmt.Fprintf(w, `
<label for="uname"><b>Username</b></label>
<input type="text" placeholder="Enter Username" name="uname" required />
<label for="psw"><b>Password</b></label>
<input type="password" placeholder="Enter Password" name="psw" required />
<button type="submit">Login</button>
</form>
</body>
</html>`)
}

func (tp *TestProvider) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	username := r.FormValue("uname")
	password := r.FormValue("psw")

	tp.mu.Lock()
	subject, ok := tp.config.SubjectInfo[username]
	tp.mu.Unlock()

	if !ok || subject.Password != password {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	redirectURI := r.FormValue("redirect_uri")
	state := r.FormValue("state")
	nonce := r.FormValue("nonce")

	allowed := slices.Contains(tp.config.AllowedRedirectURIs, redirectURI)
	if !allowed {
		http.Error(w, "redirect_uri not allowed", http.StatusBadRequest)
		return
	}

	claims := make(map[string]any)
	maps.Copy(claims, tp.config.CustomClaims)
	maps.Copy(claims, subject.CustomClaims)
	claims["sub"] = username
	if nonce != "" {
		claims["nonce"] = nonce
	}

	code := fmt.Sprintf("auth-code-%d", time.Now().UnixNano())

	expiresAt := time.Now().Add(5 * time.Minute)
	claims["exp"] = expiresAt.Unix()
	claims["iat"] = time.Now().Unix()
	claims["iss"] = tp.addr
	claims["aud"] = *tp.config.ClientID

	codeEntry := authCodeEntry{
		subject:  username,
		redirect: redirectURI,
		nonce:    nonce,
		claims:   claims,
	}

	tp.mu.Lock()
	tp.codes[code] = codeEntry
	tp.mu.Unlock()

	redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, state)
	http.Redirect(w, r, redirectURL, http.StatusFound) //nolint:gosec
}

func (tp *TestProvider) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")
	codeVerifier := r.FormValue("code_verifier")

	if tp.config.PKCEVerifier != "" && codeVerifier != tp.config.PKCEVerifier {
		http.Error(w, "invalid code_verifier", http.StatusBadRequest)
		return
	}

	tp.mu.Lock()
	entry, ok := tp.codes[code]
	tp.mu.Unlock()

	if !ok {
		http.Error(w, "invalid code", http.StatusBadRequest)
		return
	}

	if entry.redirect != redirectURI {
		http.Error(w, "redirect_uri mismatch", http.StatusBadRequest)
		return
	}

	tp.mu.Lock()
	delete(tp.codes, code)
	tp.mu.Unlock()

	claims := jwt.MapClaims{}
	maps.Copy(claims, entry.claims)
	claims["jti"] = fmt.Sprintf("id-token-%d", time.Now().UnixNano())
	claims["auth_time"] = time.Now().Unix()

	tp.mu.Lock()
	signingKey := tp.config.SigningKey
	tp.mu.Unlock()

	token := jwt.NewWithClaims(jwt.GetSigningMethod(signingKey.Alg), claims)
	token.Header["kid"] = tp.keyID

	idToken, err := token.SignedString(signingKey.PrivKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	accessToken := fmt.Sprintf("access-token-%s", code)

	tp.mu.Lock()
	tp.sessions[accessToken] = sessionEntry{
		subject: entry.subject,
		claims:  entry.claims,
	}
	tp.mu.Unlock()

	response := map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"id_token":     idToken,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response) //nolint:errchkjson
}

func (tp *TestProvider) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	accessToken := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		accessToken = authHeader[7:]
	}

	tp.mu.Lock()
	session, ok := tp.sessions[accessToken]
	tp.mu.Unlock()

	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tp.mu.Lock()
	subject, ok := tp.config.SubjectInfo[session.subject]
	tp.mu.Unlock()

	userInfo := map[string]any{
		"sub": session.subject,
	}
	if ok && subject != nil {
		maps.Copy(userInfo, subject.UserInfo)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(userInfo) //nolint:errchkjson
}

type TestProviderOption func(*TestProviderDefaults)

func WithNoTLS() TestProviderOption {
	return func(d *TestProviderDefaults) {}
}

func WithTestDefaults(defaults *TestProviderDefaults) TestProviderOption {
	return func(d *TestProviderDefaults) {
		*d = *defaults
	}
}
