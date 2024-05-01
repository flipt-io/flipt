package cloud

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
)

// Flow holds the state for the steps of our OAuth-like Web Application flow.
type Flow struct {
	server *localServer
	state  string
}

// InitFlow creates a new Flow instance by detecting a locally available port number.
func InitFlow() (*Flow, error) {
	server, err := bindLocalServer()
	if err != nil {
		return nil, err
	}

	state := randomString(20)

	return &Flow{
		server: server,
		state:  state,
	}, nil
}

// BrowserURL appends GET query parameters to baseURL and returns the url that the user should
// navigate to in their web browser.
func (f *Flow) BrowserURL(baseURL string) (string, error) {
	const callbackURL = "http://localhost:8080/cloud/auth/callback"

	ru, err := url.Parse(callbackURL)
	if err != nil {
		return "", err
	}

	ru.Host = fmt.Sprintf("%s:%d", ru.Hostname(), f.server.Port())
	f.server.CallbackPath = ru.Path

	q := url.Values{}
	q.Set("redirect_url", ru.String())
	q.Set("state", f.state)

	return fmt.Sprintf("%s?%s", baseURL, q.Encode()), nil
}

// StartServer starts the localhost server and blocks until it has received the web redirect. The
// writeSuccess function can be used to render a HTML page to the user upon completion.
func (f *Flow) StartServer(writeSuccess func(io.Writer)) error {
	f.server.WriteSuccessHTML = writeSuccess
	return f.server.Serve()
}

func (f *Flow) Close() error {
	return f.server.Close()
}

// Wait blocks until the browser flow has completed and returns the access token.
func (f *Flow) Wait(ctx context.Context) (string, error) {
	resp, err := f.server.Wait(ctx)
	if err != nil {
		return "", err
	}
	if resp.State != f.state {
		return "", errors.New("state mismatch")
	}

	return resp.Token, nil
}

func randomString(length int) string {
	b := make([]byte, length/2)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
