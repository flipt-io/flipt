flags: [...#Flag]

segments: [...#Segment]

#Flag: {
	key:         string
	name:        string
	description: string
	enabled:     bool | *false
	variants: [...#Variant]
	rules: [...#Rule]
}

#Variant: {
	key:        string
	name:       string
	attachment: {...} | *null
}

#Rule: {
	segment: string
	rank:    int
	distributions: [...#Distribution]
}

#Distribution: {
	variant: string
	rollout: >=0 & <=100
}

#Segment: {
	key:         string
	name:        string
	match_type:  string
	description: string
	constraints: [...#Constraint]
}

#Constraint: ({
	type:     "STRING_COMPARISON_TYPE"
	property: string
	value:    string
	operator: "eq" | "neq" | "empty" | "notempty" | "suffix" | "prefix"
} | {
	type:     "NUMBER_COMPARISON_TYPE"
	property: string
	value:    string
	operator: "eq" | "neq" | "present" | "notpresent" | "le" | "lte" | "gt" | "gte"
})
