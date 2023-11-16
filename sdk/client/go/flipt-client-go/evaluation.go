package evaluation

/*
#cgo LDFLAGS: -L. -lengine
#include <stdlib.h>
#include <string.h>

void* initialize_engine(char** namespaces);

char* variant(void* engine, char* evaluation_request);

char* boolean(void* engine, char* evaluation_request);

void destroy_engine(void* engine);
*/
import "C"
import (
	"context"
	"encoding/json"
	"unsafe"
)

// Client wraps the functionality of making variant and boolean evaluation of Flipt feature flags
// using an engine that is compiled to a dynamically linked library.
type Client struct {
	engine    unsafe.Pointer
	namespace string
}

// NewClient constructs an Client.
func NewClient(opts ...clientOption) *Client {
	client := &Client{
		namespace: "default",
	}

	for _, opt := range opts {
		opt(client)
	}

	ns := []*C.char{C.CString(client.namespace)}

	// Free the memory of the C Strings that were created to initialize the engine.
	defer func() {
		for _, n := range ns {
			C.free(unsafe.Pointer(n))
		}
	}()

	nsPtr := (**C.char)(unsafe.Pointer(&ns[0]))

	eng := C.initialize_engine(nsPtr)

	client.engine = eng

	return client
}

// clientOption adds additional configuraiton for Client parameters
type clientOption func(*Client)

// WithNamespace allows for specifying which namespace the clients wants to make evaluations from.
func WithNamespace(namespace string) clientOption {
	return func(c *Client) {
		c.namespace = namespace
	}
}

// Variant makes an evaluation on a variant flag using the allocated Rust engine.
func (e *Client) Variant(_ context.Context, flagKey, entityID string, evalContext map[string]string) (*VariantResult, error) {
	eb, err := json.Marshal(evalContext)
	if err != nil {
		return nil, err
	}

	ereq, err := json.Marshal(evaluationRequest{
		NamespaceKey: e.namespace,
		FlagKey:      flagKey,
		EntityId:     entityID,
		Context:      string(eb),
	})
	if err != nil {
		return nil, err
	}

	variant := C.variant(e.engine, C.CString(string(ereq)))
	defer C.free(unsafe.Pointer(variant))

	b := C.GoBytes(unsafe.Pointer(variant), (C.int)(C.strlen(variant)))

	var vr *VariantResult

	if err := json.Unmarshal(b, &vr); err != nil {
		return nil, err
	}

	return vr, nil
}

// Boolean makes an evaluation on a boolean flag using the allocated Rust engine.
func (e *Client) Boolean(_ context.Context, flagKey, entityID string, evalContext map[string]string) (*BooleanResult, error) {
	eb, err := json.Marshal(evalContext)
	if err != nil {
		return nil, err
	}

	ereq, err := json.Marshal(evaluationRequest{
		NamespaceKey: e.namespace,
		FlagKey:      flagKey,
		EntityId:     entityID,
		Context:      string(eb),
	})
	if err != nil {
		return nil, err
	}

	boolean := C.boolean(e.engine, C.CString(string(ereq)))
	defer C.free(unsafe.Pointer(boolean))

	b := C.GoBytes(unsafe.Pointer(boolean), (C.int)(C.strlen(boolean)))

	var br *BooleanResult

	if err := json.Unmarshal(b, &br); err != nil {
		return nil, err
	}

	return br, nil
}

// Close cleans up the allocated engine as it was initialized in the constructor.
func (e *Client) Close() error {
	// Destroy the engine to clean up allocated memory on dynamic library side.
	C.destroy_engine(e.engine)
	return nil
}
