package cloud

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

const callbackPath = "/cloud/auth/callback"

// TokenResponse represents the token received by the local server's callback handler.
type TokenResponse struct {
	Token string
	State string
}

// bindLocalServer initializes a LocalServer that will listen on a randomly available TCP port.
func bindLocalServer() (*localServer, error) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	return &localServer{
		listener:   listener,
		resultChan: make(chan TokenResponse, 1),
	}, nil
}

type localServer struct {
	WriteSuccessHTML func(w io.Writer)

	resultChan chan (TokenResponse)
	listener   net.Listener
}

func (s *localServer) Port() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

func (s *localServer) Close() error {
	return safeClose(s.listener)
}

func (s *localServer) Serve() error {
	return http.Serve(s.listener, s) //nolint:gosec
}

func (s *localServer) Wait(ctx context.Context) (TokenResponse, error) {
	select {
	case <-ctx.Done():
		return TokenResponse{}, ctx.Err()
	case code, ok := <-s.resultChan:
		if !ok {
			return TokenResponse{}, errors.New("server shutdown")
		}
		return code, nil
	}
}

// ServeHTTP implements http.Handler.
func (s *localServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != callbackPath {
		w.WriteHeader(404)
		return
	}
	defer func() {
		_ = s.Close()
		close(s.resultChan)
	}()

	w.Header().Add("content-type", "text/html")
	if s.WriteSuccessHTML != nil {
		s.WriteSuccessHTML(w)
	} else {
		defaultSuccessHTML(w)
	}

	params := r.URL.Query()
	s.resultChan <- TokenResponse{
		Token: params.Get("id_token"),
		State: params.Get("state"),
	}
}

func defaultSuccessHTML(w io.Writer) {
	fmt.Fprintf(w, "<p>You may now close this page and return to the client app.</p>")
}

func safeClose(c io.Closer) error {
	if c == nil {
		return nil
	}
	return c.Close()
}
