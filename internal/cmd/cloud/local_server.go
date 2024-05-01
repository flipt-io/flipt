package cloud

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

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
	CallbackPath     string
	WriteSuccessHTML func(w io.Writer)

	resultChan chan (TokenResponse)
	server     *http.Server
	listener   net.Listener
}

func (s *localServer) Port() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

func (s *localServer) Close() error {
	return safeClose(s.server)
}

func (s *localServer) Serve() error {
	s.server = &http.Server{
		Handler:           s,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return s.server.Serve(s.listener)
}

func (s *localServer) Wait(ctx context.Context) (TokenResponse, error) {
	select {
	case <-ctx.Done():
		return TokenResponse{}, ctx.Err()
	case code := <-s.resultChan:
		return code, nil
	}
}

// ServeHTTP implements http.Handler.
func (s *localServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.CallbackPath != "" && r.URL.Path != s.CallbackPath {
		w.WriteHeader(404)
		return
	}
	defer func() {
		_ = s.Close()
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
