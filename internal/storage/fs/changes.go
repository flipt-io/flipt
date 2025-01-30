package fs

import (
	"fmt"
	"strings"

	"go.flipt.io/flipt/internal/server/environments"
)

type Verb string

const (
	VerbCreate = Verb("create")
	VerbUpdate = Verb("update")
	VerbDelete = Verb("delete")
)

type Change struct {
	Verb     Verb
	Resource Resource
}

func (c Change) String() string {
	return fmt.Sprintf("%s %s", c.Verb, c.Resource)
}

type Resource struct {
	Type      environments.ResourceType
	Namespace string
	Key       string
}

func (r Resource) String() string {
	if r.Namespace == "" {
		return fmt.Sprintf("%s %s", strings.ToLower(r.Type.Name), r.Key)
	}

	return fmt.Sprintf("%s %s/%s", strings.ToLower(r.Type.Name), r.Namespace, r.Key)
}
