
package flipt.authz.v1

import rego.v1
import data

viewable_namespaces contains namespace if {
	some role in input.roles
	some namespace in data.roles_to_namespaces[role]
}

default allow := false

allow if {
	input.request.namespace in viewable_namespaces
}
