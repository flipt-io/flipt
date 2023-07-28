close({
	version:   "1.0" | *"1.1"
	namespace: string & =~"^[-_,A-Za-z0-9]+$" | *"default"
	flags: [...{_version: version} & #Flag]
	segments: [...#Segment]
})

#Flag: {
	_version:     string
	key:          string & =~"^[-_,A-Za-z0-9]+$"
	name:         string & =~"^.+$"
	description?: string
	enabled:      bool | *false
	variants: [...#Variant]
	rules: [...#Rule]
	if _version == "1.1" {
		type: "BOOLEAN_FLAG_TYPE" | *"VARIANT_FLAG_TYPE"
		#FlagBoolean | *{}
	}
}

#FlagBoolean: {
	type: "BOOLEAN_FLAG_TYPE"
	rollouts: [...{
		description: string
		#Rollout
	}]
}

#Variant: {
	key:        string & =~"^.+$"
	name:       string & =~"^.+$"
	description?: string
	attachment: {...} | *null
}

#Rule: {
	segment: string & =~"^.+$"
	rank?:   int
	distributions: [...#Distribution]
}

#Distribution: {
	variant: string & =~"^.+$"
	rollout: >=0 & <=100
}

#Rollout: {
	segment: {
		key:   string
		value: bool
	}
} | {
	threshold: {
		percentage: float
		value:      bool
	}
	// failure to add the following causes it not to close
} | *{} // I found a comment somewhere that this helps with distinguishing disjunctions

#Segment: {
	key:          string & =~"^[-_,A-Za-z0-9]+$"
	name:         string & =~"^.+$"
	match_type:   "ANY_MATCH_TYPE" | "ALL_MATCH_TYPE"
	description?: string
	constraints: [...#Constraint]
}

#Constraint: ({
	type:         "STRING_COMPARISON_TYPE"
	property:     string & =~"^.+$"
	value?:       string
	description?: string
	operator:     "eq" | "neq" | "empty" | "notempty" | "prefix" | "suffix"
} | {
	type:         "NUMBER_COMPARISON_TYPE"
	property:     string & =~"^.+$"
	value?:       string
	description?: string
	operator:     "eq" | "neq" | "present" | "notpresent" | "le" | "lte" | "gt" | "gte"
} | {
	type:         "BOOLEAN_COMPARISON_TYPE"
	property:     string & =~"^.+$"
	value?:       string
	operator:     "true" | "false" | "present" | "notpresent"
	description?: string
} | {
	type:         "DATETIME_COMPARISON_TYPE"
	property:     string & =~"^.+$"
	value?:       string
	description?: string
	operator:     "eq" | "neq" | "present" | "notpresent" | "le" | "lte" | "gt" | "gte"
})
