// Flipt Commercial Open Source Feature
// This file contains functionality that is licensed under the Flipt Fair Core License (FCL).
// You may NOT use, modify, or distribute this file or its contents without a valid paid license.
// For details: https://github.com/flipt-io/flipt/blob/v2/LICENSE

// Package webhook authenticates incoming SCM webhook requests and extracts the
// branches updated by push events.
//
// The package is independent of configuration wiring and secret resolution:
// callers pass the environment's SCM type and the already-resolved secret.
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
)

// MaxBodySize is the largest request body accepted. It matches GitHub's
// documented webhook payload cap; larger payloads are rejected rather than
// truncated, since truncation would invalidate the signature anyway.
const MaxBodySize int64 = 25 << 20

// Request headers inspected by the verifiers.
const (
	headerGitHubEvent     = "X-GitHub-Event"
	headerGitHubSignature = "X-Hub-Signature-256"
	headerGiteaEvent      = "X-Gitea-Event"
	headerGiteaSignature  = "X-Gitea-Signature"
	headerGitLabEvent     = "X-Gitlab-Event"
	headerGitLabToken     = "X-Gitlab-Token" //nolint:gosec // header name, not a credential
	headerBitbucketEvent  = "X-Event-Key"
	headerBitbucketSig    = "X-Hub-Signature"
)

const (
	refHeadsPrefix  = "refs/heads/"
	sha256SigPrefix = "sha256="
)

var (
	// ErrBodyTooLarge is returned when the request body exceeds MaxBodySize.
	ErrBodyTooLarge = errors.New("webhook: request body too large")

	errUnauthenticated = errs.ErrUnauthenticated("webhook: invalid or missing signature")
)

// Kind classifies an authenticated webhook event.
type Kind string

const (
	// KindPush is a push to one or more refs. Branches holds the pushed
	// branches; it is empty when only non-branch refs (e.g. tags) changed.
	KindPush Kind = "push"
	// KindPing is a provider connectivity check.
	KindPing Kind = "ping"
	// KindOther is any other event; it carries no branches.
	KindOther Kind = "other"
)

// Event is the result of handling an authenticated webhook request.
type Event struct {
	Kind Kind
	// Name is the provider's raw event name (e.g. "push", "Push Hook",
	// "repo:refs_changed", "git.push").
	Name string
	// Branches lists the distinct branch names (without refs/heads/) updated
	// by a push, in payload order.
	Branches []string
}

// Handle reads the request body (capped at MaxBodySize), authenticates the
// request for the given SCM using secret, and parses the event.
//
// Authentication failures return an errors.ErrUnauthenticated (map to 401).
// An oversized body returns ErrBodyTooLarge. An unsupported SCM type or an
// authentic but malformed payload returns an errors.ErrInvalid.
func Handle(scm config.SCMType, secret []byte, r *http.Request) (Event, error) {
	body, err := readBody(r.Body)
	if err != nil {
		return Event{}, err
	}

	if err := Verify(scm, secret, r, body); err != nil {
		return Event{}, err
	}

	return Parse(scm, r.Header, body)
}

func readBody(body io.Reader) ([]byte, error) {
	if body == nil {
		return nil, nil
	}

	b, err := io.ReadAll(io.LimitReader(body, MaxBodySize+1))
	if err != nil {
		return nil, fmt.Errorf("webhook: reading body: %w", err)
	}

	if int64(len(b)) > MaxBodySize {
		return nil, ErrBodyTooLarge
	}

	return b, nil
}

// Verify authenticates a webhook request for the given SCM type:
//
//   - github: HMAC-SHA256 of body in X-Hub-Signature-256 ("sha256=<hex>")
//   - gitea: HMAC-SHA256 of body in X-Gitea-Signature (bare hex)
//   - bitbucket: HMAC-SHA256 of body in X-Hub-Signature ("sha256=<hex>"),
//     for both Bitbucket Cloud and Server/Data Center
//   - gitlab: shared token in X-Gitlab-Token
//   - azure: HTTP Basic auth with secret as the password (username ignored)
//
// All comparisons are constant-time. An empty secret never authenticates.
func Verify(scm config.SCMType, secret []byte, r *http.Request, body []byte) error {
	switch scm {
	case config.GitHubSCMType, config.GiteaSCMType, config.GitLabSCMType,
		config.BitBucketSCMType, config.AzureSCMType:
	default:
		return errs.ErrInvalidf("webhook: unsupported SCM type %q", scm)
	}

	if len(secret) == 0 {
		return errUnauthenticated
	}

	var ok bool
	switch scm {
	case config.GitHubSCMType:
		ok = verifyHMAC(secret, body, r.Header.Get(headerGitHubSignature), true)
	case config.GiteaSCMType:
		ok = verifyHMAC(secret, body, r.Header.Get(headerGiteaSignature), false)
	case config.BitBucketSCMType:
		ok = verifyHMAC(secret, body, r.Header.Get(headerBitbucketSig), true)
	case config.GitLabSCMType:
		ok = constantTimeEqual([]byte(r.Header.Get(headerGitLabToken)), secret)
	case config.AzureSCMType:
		_, password, hasAuth := r.BasicAuth()
		ok = hasAuth && constantTimeEqual([]byte(password), secret)
	}

	if !ok {
		return errUnauthenticated
	}

	return nil
}

func verifyHMAC(secret, body []byte, signature string, prefixed bool) bool {
	if prefixed {
		var found bool
		if signature, found = strings.CutPrefix(signature, sha256SigPrefix); !found {
			return false
		}
	}

	got, err := hex.DecodeString(signature)
	if err != nil || len(got) != sha256.Size {
		return false
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write(body)

	return hmac.Equal(got, mac.Sum(nil))
}

// constantTimeEqual compares a and b in time independent of their contents.
// Lengths are hashed first so the comparison doesn't leak the secret length.
func constantTimeEqual(a, b []byte) bool {
	ha, hb := sha256.Sum256(a), sha256.Sum256(b)
	return subtle.ConstantTimeCompare(ha[:], hb[:]) == 1
}

// Parse classifies an already-authenticated webhook and extracts the pushed
// branches. Non-push events return KindPing or KindOther with no branches.
func Parse(scm config.SCMType, header http.Header, body []byte) (Event, error) {
	switch scm {
	case config.GitHubSCMType:
		return parseRefEvent(header.Get(headerGitHubEvent), "push", "ping", body)
	case config.GiteaSCMType:
		return parseRefEvent(header.Get(headerGiteaEvent), "push", "", body)
	case config.GitLabSCMType:
		return parseRefEvent(header.Get(headerGitLabEvent), "Push Hook", "", body)
	case config.BitBucketSCMType:
		return parseBitbucket(header.Get(headerBitbucketEvent), body)
	case config.AzureSCMType:
		return parseAzure(body)
	default:
		return Event{}, errs.ErrInvalidf("webhook: unsupported SCM type %q", scm)
	}
}

// parseRefEvent handles the GitHub, Gitea and GitLab push payloads, which all
// carry the updated ref in a top-level "ref" field.
func parseRefEvent(name, pushName, pingName string, body []byte) (Event, error) {
	switch {
	case name == pushName:
	case pingName != "" && name == pingName:
		return Event{Kind: KindPing, Name: name}, nil
	default:
		return Event{Kind: KindOther, Name: name}, nil
	}

	var payload struct {
		Ref string `json:"ref"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return Event{}, errs.ErrInvalidf("webhook: decoding %s payload: %v", name, err)
	}

	return newPushEvent(name, payload.Ref), nil
}

func parseBitbucket(name string, body []byte) (Event, error) {
	switch name {
	case "repo:push":
		return parseBitbucketCloud(name, body)
	case "repo:refs_changed":
		return parseBitbucketServer(name, body)
	case "diagnostics:ping":
		return Event{Kind: KindPing, Name: name}, nil
	default:
		return Event{Kind: KindOther, Name: name}, nil
	}
}

type bitbucketCloudRef struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func parseBitbucketCloud(name string, body []byte) (Event, error) {
	var payload struct {
		Push struct {
			Changes []struct {
				New *bitbucketCloudRef `json:"new"`
				Old *bitbucketCloudRef `json:"old"`
			} `json:"changes"`
		} `json:"push"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return Event{}, errs.ErrInvalidf("webhook: decoding %s payload: %v", name, err)
	}

	refs := make([]string, 0, len(payload.Push.Changes))
	for _, change := range payload.Push.Changes {
		// new is null when a branch is deleted; fall back to old so the
		// deletion is still reported against its branch.
		ref := change.New
		if ref == nil {
			ref = change.Old
		}
		if ref != nil && ref.Type == "branch" && ref.Name != "" {
			refs = append(refs, refHeadsPrefix+ref.Name)
		}
	}

	return newPushEvent(name, refs...), nil
}

func parseBitbucketServer(name string, body []byte) (Event, error) {
	var payload struct {
		Changes []struct {
			RefID string `json:"refId"`
			Ref   struct {
				ID string `json:"id"`
			} `json:"ref"`
		} `json:"changes"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return Event{}, errs.ErrInvalidf("webhook: decoding %s payload: %v", name, err)
	}

	refs := make([]string, 0, len(payload.Changes))
	for _, change := range payload.Changes {
		ref := change.Ref.ID
		if ref == "" {
			ref = change.RefID
		}
		refs = append(refs, ref)
	}

	return newPushEvent(name, refs...), nil
}

func parseAzure(body []byte) (Event, error) {
	var payload struct {
		EventType string `json:"eventType"`
		Resource  struct {
			RefUpdates []struct {
				Name string `json:"name"`
			} `json:"refUpdates"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return Event{}, errs.ErrInvalidf("webhook: decoding azure payload: %v", err)
	}

	if payload.EventType != "git.push" {
		return Event{Kind: KindOther, Name: payload.EventType}, nil
	}

	refs := make([]string, 0, len(payload.Resource.RefUpdates))
	for _, update := range payload.Resource.RefUpdates {
		refs = append(refs, update.Name)
	}

	return newPushEvent(payload.EventType, refs...), nil
}

// newPushEvent builds a push event from full ref names, keeping only distinct
// branches (refs/heads/*) and dropping tags and other refs.
func newPushEvent(name string, refs ...string) Event {
	ev := Event{Kind: KindPush, Name: name}
	for _, ref := range refs {
		branch, ok := strings.CutPrefix(ref, refHeadsPrefix)
		if !ok || branch == "" || slices.Contains(ev.Branches, branch) {
			continue
		}
		ev.Branches = append(ev.Branches, branch)
	}

	return ev
}
