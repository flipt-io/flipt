package flipt.authz.v2

import rego.v1

# Default deny
default allow := false

default viewable_environments := []

# Helper to get subject identifiers from input
subject_ids contains id if {
	user := input.authentication.metadata["io.flipt.auth.user"]
	id := sprintf("user:%s", [user])
}

subject_ids contains id if {
	groups := input.authentication.metadata["io.flipt.auth.groups"]
	id := sprintf("group:%s", [groups[_]])
}

# Helper to check if subject has global access
has_global_access if {
	some binding in data.role_bindings
	some subject in binding.subjects
	some id in subject_ids
	subject == id
	binding.scope.type == "global"
}

# Helper to check if subject has wildcard access to an environment
has_wildcard_access(env) if {
	# Global access means wildcard access to all environments
	has_global_access
}

has_wildcard_access(env) if {
	# Check for wildcard namespace access in role bindings
	some binding in data.role_bindings
	some subject in binding.subjects
	some id in subject_ids
	subject == id
	binding.scope.type == "namespace"
	some b in binding.scope.bindings
	b.environment == env
	some n in b.namespaces
	n == "*"
}

# Helper to check if subject has namespace access
has_namespace_access(env, ns) if {
	# Global access means access to any environment and namespace
	has_global_access
}

has_namespace_access(env, ns) if {
	# Check namespace-level access
	some binding in data.role_bindings
	some subject in binding.subjects
	some id in subject_ids
	subject == id
	binding.scope.type == "namespace"
	some b in binding.scope.bindings
	b.environment == env
	some n in b.namespaces
	n == ns
}

has_namespace_access(env, ns) if {
	# Check resource-level access
	some binding in data.role_bindings
	some subject in binding.subjects
	some id in subject_ids
	subject == id
	binding.scope.type == "resource"
	some b in binding.scope.bindings
	b.environment == env
	some n in b.namespaces
	n == ns
}

# Helper to check if subject has required permission for a scope type
has_permission(scope_type, env, ns, action) if {
	# Global access has all permissions
	has_global_access
}

has_permission(scope_type, env, _, action) if {
	# Check namespace-level permissions with wildcard access
	some binding in data.role_bindings
	some subject in binding.subjects
	some id in subject_ids
	subject == id
	binding.scope.type == scope_type
	some b in binding.scope.bindings
	b.environment == env
	some n in b.namespaces
	n == "*"
	some perm in b.permissions
	perm == action
}

has_permission(scope_type, env, ns, action) if {
	# Check namespace-level permissions for specific namespace
	some binding in data.role_bindings
	some subject in binding.subjects
	some id in subject_ids
	subject == id
	binding.scope.type == scope_type
	some b in binding.scope.bindings
	b.environment == env
	some n in b.namespaces
	n == ns
	some perm in b.permissions
	perm == action
}

# Get list of viewable environments
viewable_environments := ["*"] if {
	# If global access, return wildcard
	has_global_access
} else := envs if {
	# Return environments user has access to based on bindings
	envs := {env |
		some binding in data.role_bindings
		binding.scope.type in {"resource", "namespace"}
		some subject in binding.subjects
		some id in subject_ids
		subject == id
		some b in binding.scope.bindings
		env := b.environment
	}
}

# Get list of viewable namespaces for a specific environment
viewable_namespaces_for_environment(env) := ["*"] if {
	# If wildcard access for this environment, return wildcard
	has_wildcard_access(env)
} else := ns if {
	# Return namespaces user has access to in the environment
	ns := {n |
		some binding in data.role_bindings
		some subject in binding.subjects
		some id in subject_ids
		subject == id
		binding.scope.type in {"resource", "namespace"}
		some b in binding.scope.bindings
		b.environment == env
		some n in b.namespaces
		n != "*"
		n == n # ensure n is bound
	}
}

# Allow access based on scope type, environment, namespace and action
allow if {
	has_global_access
}

allow if {
	# Get the request details
	scope := input.request.scope
	env := input.request.environment
	action := input.request.action

	# For namespace creation only, just check permission
	scope == "namespace"
	action == "create"
	has_permission(scope, env, "", action)
}

allow if {
	# Get the request details
	scope := input.request.scope
	env := input.request.environment
	ns := input.request.namespace
	action := input.request.action

	# For all other operations, verify both access and permission
	has_namespace_access(env, ns)
	has_permission(scope, env, ns, action)
}
