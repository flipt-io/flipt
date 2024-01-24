package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/hashicorp/cap/oidc"
)

const (
	openidConfiguration = "/.well-known/openid-configuration"
	wellKnownJwks       = "/.well-known/jwks.json"
	domain              = "https://discover.svc"
)

func main() {
	privKeyPath := flag.String("private-key", "", "path to private key")
	addr := flag.String("addr", ":443", "Port for OIDC server")
	flag.Parse()

	p, err := os.ReadFile(*privKeyPath)
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode([]byte(p))
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		switch r.URL.Path {
		case openidConfiguration:
			reply := struct {
				Issuer        string   `json:"issuer"`
				JWKSURI       string   `json:"jwks_uri"`
				SupportedAlgs []string `json:"id_token_signing_alg_values_supported"`
			}{
				Issuer:        domain,
				JWKSURI:       domain + wellKnownJwks,
				SupportedAlgs: []string{string(oidc.RS256)},
			}

			if err := json.NewEncoder(w).Encode(&reply); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case wellKnownJwks:
			if err := json.NewEncoder(w).Encode(&jose.JSONWebKeySet{
				Keys: []jose.JSONWebKey{
					{
						Key:   &priv.PublicKey,
						KeyID: fmt.Sprintf("%d", time.Now().Unix()),
					},
				},
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	})

	_ = http.ListenAndServeTLS(*addr, "/server.crt", "/server.key", handler)
}
