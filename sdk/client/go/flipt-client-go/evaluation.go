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
// using an engine that is compiled to a dynamic linking library.
type Client struct {
	engine unsafe.Pointer
}

// NewClient constructs an Client.
func NewClient(namespace string) *Client {
	ns := []*C.char{C.CString(namespace)}

	// Free the memory of the C Strings that were created to initialize the engine.
	defer func() {
		for _, n := range ns {
			C.free(unsafe.Pointer(n))
		}
	}()

	nsPtr := (**C.char)(unsafe.Pointer(&ns[0]))

	eng := C.initialize_engine(nsPtr)

	return &Client{
		engine: eng,
	}
}

// Variant makes an evaluation on a variant flag using the allocated Rust engine.
func (e *Client) Variant(_ context.Context, evaluationRequest *EvaluationRequest) (*VariantEvaluationResponse, error) {
	ereq, err := json.Marshal(evaluationRequest)
	if err != nil {
		return nil, err
	}

	variant := C.variant(e.engine, C.CString(string(ereq)))
	defer C.free(unsafe.Pointer(variant))

	b := C.GoBytes(unsafe.Pointer(variant), (C.int)(C.strlen(variant)))

	var ver *VariantEvaluationResponse

	if err := json.Unmarshal(b, &ver); err != nil {
		return nil, err
	}

	return ver, nil
}

// Boolean makes an evaluation on a boolean flag using the allocated Rust engine.
func (e *Client) Boolean(_ context.Context, evaluationRequest *EvaluationRequest) (*BooleanEvaluationResponse, error) {
	ereq, err := json.Marshal(evaluationRequest)
	if err != nil {
		return nil, err
	}

	boolean := C.boolean(e.engine, C.CString(string(ereq)))
	defer C.free(unsafe.Pointer(boolean))

	b := C.GoBytes(unsafe.Pointer(boolean), (C.int)(C.strlen(boolean)))

	var ber *BooleanEvaluationResponse

	if err := json.Unmarshal(b, &ber); err != nil {
		return nil, err
	}

	return ber, nil
}

// Close cleans up the allocated engine as it was initialized in the constructor.
func (e *Client) Close() error {
	// Destroy the engine to clean up allocated memory on dynamic library side.
	C.destroy_engine(e.engine)
	return nil
}
