package flipt.authz.v2

import future.keywords.if
import future.keywords.in

# Default deny
default allow = false
default viewable_environments = []

# Helper to get subject identifiers from input
subject_ids[id] {
    id := input.authentication.metadata["io.flipt.auth.user"]
}

subject_ids[id] {
    groups := input.authentication.metadata["io.flipt.auth.groups"]
    id := concat("group:", [groups[_]])
}

# Helper to check if subject has global access
has_global_access if {
    some binding in data.role_bindings
    some subject in binding.subjects
    some id in subject_ids
    subject == id
    binding.scope.type == "global"
}

# Helper to check if subject has namespace access
has_namespace_access(env, ns) {
    # Check global access first
    has_global_access
}

has_namespace_access(env, ns) {
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

has_namespace_access(env, _) {
    # Check wildcard namespace access
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

# Helper to check if subject has required permission for a scope type
has_permission(scope_type, env, ns, action) {
    # Global access has all permissions
    has_global_access
}

has_permission(scope_type, env, ns, action) {
    # Check namespace-level permissions
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

has_permission(scope_type, env, _, action) {
    # Check wildcard namespace permissions
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

# Get list of viewable environments
viewable_environments = envs if {
    # If global access, return all environments
    has_global_access
    envs := {env | data.environments[env]}
} else = envs if {
    # Return environments user has read access to based on bindings
    envs := {env |
        data.environments[env]
        some binding in data.role_bindings
        some subject in binding.subjects
        some id in subject_ids
        subject == id
        some b in binding.scope.bindings
        b.environment == env
        # Ensure user has read permission in at least one namespace
        some perm in b.permissions
        perm == "read"
    }
}

# Get list of viewable namespaces for a specific environment
viewable_namespaces_for_environment(env) = ns if {
    # If global access, return all namespaces for the environment
    has_global_access
    ns := {n | n := data.environments[env].namespaces[_]}
} else = ns if {
    # Return namespaces user has access to in the environment
    ns := {n |
        n := data.environments[env].namespaces[_]
        has_namespace_access(env, n)
    }
}

# Allow access based on scope type, environment, namespace and action
allow if {
    # Get the request details
    scope := input.request.scope      # "namespace" or "resource"
    env := input.request.environment
    ns := input.request.namespace
    action := input.request.action
    
    # Verify access and permission
    has_namespace_access(env, ns)
    has_permission(scope, env, ns, action)
} 