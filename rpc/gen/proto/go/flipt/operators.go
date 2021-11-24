package flipt

const (
	OpEQ         = "eq"
	OpNEQ        = "neq"
	OpLT         = "lt"
	OpLTE        = "lte"
	OpGT         = "gt"
	OpGTE        = "gte"
	OpEmpty      = "empty"
	OpNotEmpty   = "notempty"
	OpTrue       = "true"
	OpFalse      = "false"
	OpPresent    = "present"
	OpNotPresent = "notpresent"
	OpPrefix     = "prefix"
	OpSuffix     = "suffix"
)

var (
	ValidOperators = map[string]struct{}{
		OpEQ:         {},
		OpNEQ:        {},
		OpLT:         {},
		OpLTE:        {},
		OpGT:         {},
		OpGTE:        {},
		OpEmpty:      {},
		OpNotEmpty:   {},
		OpTrue:       {},
		OpFalse:      {},
		OpPresent:    {},
		OpNotPresent: {},
		OpPrefix:     {},
		OpSuffix:     {},
	}
	NoValueOperators = map[string]struct{}{
		OpTrue:       {},
		OpFalse:      {},
		OpEmpty:      {},
		OpNotEmpty:   {},
		OpPresent:    {},
		OpNotPresent: {},
	}
	StringOperators = map[string]struct{}{
		OpEQ:       {},
		OpNEQ:      {},
		OpEmpty:    {},
		OpNotEmpty: {},
		OpPrefix:   {},
		OpSuffix:   {},
	}
	NumberOperators = map[string]struct{}{
		OpEQ:         {},
		OpNEQ:        {},
		OpLT:         {},
		OpLTE:        {},
		OpGT:         {},
		OpGTE:        {},
		OpPresent:    {},
		OpNotPresent: {},
	}
	BooleanOperators = map[string]struct{}{
		OpTrue:       {},
		OpFalse:      {},
		OpPresent:    {},
		OpNotPresent: {},
	}
)
