package ecr

import "net/http"

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (fn RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := fn(r)
	if err != nil {
		return resp, err
	}
	// when authorization token has expired, AWS ECR response with http code 403 but oras expected
	// http code 401 in this case.
	if resp.StatusCode == http.StatusForbidden {
		resp.StatusCode = http.StatusUnauthorized
	}
	return resp, nil
}
