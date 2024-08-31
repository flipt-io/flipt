package ext

import (
	"errors"
	"fmt"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
)

func init() {
	rego.RegisterBuiltin2(&rego.Function{
		Name: "flipt.is_auth_method",
		Decl: types.NewFunction(types.Args(types.A, types.S), types.B),
	}, isAuthMethod)
}

var labelMethodTable = map[string]*ast.Term{
	"token":      ast.IntNumberTerm(int(authrpc.Method_METHOD_TOKEN.Number())),
	"oidc":       ast.IntNumberTerm(int(authrpc.Method_METHOD_OIDC.Number())),
	"kubernetes": ast.IntNumberTerm(int(authrpc.Method_METHOD_KUBERNETES.Number())),
	"k8s":        ast.IntNumberTerm(int(authrpc.Method_METHOD_KUBERNETES.Number())),
	"github":     ast.IntNumberTerm(int(authrpc.Method_METHOD_GITHUB.Number())),
	"jwt":        ast.IntNumberTerm(int(authrpc.Method_METHOD_JWT.Number())),
	"cloud":      ast.IntNumberTerm(int(authrpc.Method_METHOD_CLOUD.Number())),
}

var (
	errNoAuthenticationFound = errors.New("no authentication found")
	authTerm                 = ast.StringTerm("authentication")
	methodTerm               = ast.StringTerm("method")
)

func isAuthMethod(_ rego.BuiltinContext, input, key *ast.Term) (*ast.Term, error) {
	var authMethod string
	if err := ast.As(key.Value, &authMethod); err != nil {
		return nil, err
	}

	methodCode, ok := labelMethodTable[authMethod]
	if !ok {
		return nil, fmt.Errorf("unsupported auth method %s", authMethod)
	}

	auth := input.Get(authTerm)
	if auth == nil {
		return nil, errNoAuthenticationFound
	}

	return ast.BooleanTerm(methodCode.Equal(auth.Get(methodTerm))), nil
}
