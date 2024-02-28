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
	OpIsOneOf    = "isoneof"
	OpIsNotOneOf = "isnotoneof"
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
		OpIsOneOf:    {},
		OpIsNotOneOf: {},
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
		OpEQ:         {},
		OpNEQ:        {},
		OpEmpty:      {},
		OpNotEmpty:   {},
		OpPrefix:     {},
		OpSuffix:     {},
		OpIsOneOf:    {},
		OpIsNotOneOf: {},
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
		OpIsOneOf:    {},
		OpIsNotOneOf: {},
	}
	BooleanOperators = map[string]struct{}{
		OpTrue:       {},
		OpFalse:      {},
		OpPresent:    {},
		OpNotPresent: {},
	}
	EntityIdOperators = map[string]struct{}{
		OpEQ:         {},
		OpNEQ:        {},
		OpIsOneOf:    {},
		OpIsNotOneOf: {},
	}
)
