import "list"

#versions: ["1.0", "1.1", "1.2", "1.3", "1.4", "1.5", "1.6"]

version: "1.0" | "1.1" | "1.2" | "1.3" | "1.4" | "1.5" | *"1.6"

close({
	version:    version
	namespace?: #Namespace
	flags: [...#Flag]
	segments: [...#Segment]
})

#Namespace: {
	key:          string & =~"^[-_,A-Za-z0-9]+$" | *"default"
	name?:        string & =~"^.+$"
	description?: string | null
} | string & =~"^[-_,A-Za-z0-9]+$"

#Flag: {
	key:          string & =~"^[-_,A-Za-z0-9]+$"
	name:         string & =~"^.+$"
	description?: string | null
	enabled:      bool | *false
	variants: [...#Variant]
	rules: [...#Rule]
	if list.Contains(#versions[1:], version) { // 1.1 or later
		type: "BOOLEAN_FLAG_TYPE" | *"VARIANT_FLAG_TYPE"
		if type == "BOOLEAN_FLAG_TYPE" {
			#FlagBoolean
		}
	}
	if list.Contains(#versions[3:], version) { // 1.3 or later
		metadata: [string]: !=null
	}
}

#FlagBoolean: {
	type: "BOOLEAN_FLAG_TYPE"
	rollouts: [...{
		description?: string | null
		#Rollout
	}]
}

#Variant: {
	key:          string & =~"^.+$"
	name?:        string & =~"^.+$"
	description?: string | null
	attachment: {...} | [...] | *null
	if list.Contains(#versions[3:], version) { // 1.3 or later
		default: bool | *false
	}
}

#RuleSegment: {
	keys: [...string]
	operator: "OR_SEGMENT_OPERATOR" | "AND_SEGMENT_OPERATOR" | *null
}

#Rule: {
	segment: string & =~"^[-_,A-Za-z0-9]+$" | #RuleSegment
	rank?:   int
	distributions: [...#Distribution]
}

#Distribution: {
	variant: string & =~"^.+$"
	rollout: >=0 & <=100
}

#RolloutSegment: {key: string & =~"^[-_,A-Za-z0-9]+$"} | {keys: [...string]}

#Rollout: {
	segment: {
		#RolloutSegment
		operator: "OR_SEGMENT_OPERATOR" | "AND_SEGMENT_OPERATOR" | *null
		value?:   bool | *false
	}
} | {
	threshold: {
		percentage: float | int
		value?:     bool | *false
	}
	// failure to add the following causes it not to close
} | *{} // I found a comment somewhere that this helps with distinguishing disjunctions

#Segment: {
	key:          string & =~"^[-_,A-Za-z0-9]+$"
	name:         string & =~"^.+$"
	match_type:   "ANY_MATCH_TYPE" | "ALL_MATCH_TYPE"
	description?: string | null
	constraints: [...#Constraint]
}

#Constraint: ({
	type:         "STRING_COMPARISON_TYPE"
	property:     string & =~"^.+$"
	value?:       string
	description?: string | null
	operator:     "eq" | "neq" | "empty" | "notempty" | "prefix" | "suffix" | "contains" | "notcontains"
} | {
	type:         "STRING_COMPARISON_TYPE"
	property:     string & =~"^.+$"
	if list.Contains(#versions[:6], version) { // before 1.6
		value: string
	}
	if !list.Contains(#versions[:6], version) { // 1.6 or later
		value: [...string]
	}
	description?: string | null
	operator:     "isoneof" | "isnotoneof"
} | {
	type:         "NUMBER_COMPARISON_TYPE"
	property:     string & =~"^.+$"
	value?:       string
	description?: string | null
	operator:     "eq" | "neq" | "present" | "notpresent" | "le" | "lte" | "gt" | "gte"
} | {
	type:         "NUMBER_COMPARISON_TYPE"
	property:     string & =~"^.+$"
	if list.Contains(#versions[:6], version) { // before 1.6
		value: string
	}
	if !list.Contains(#versions[:6], version) { // 1.6 or later
		value: [...number]
	}
	description?: string | null
	operator:     "isoneof" | "isnotoneof"
} | {
	type:         "BOOLEAN_COMPARISON_TYPE"
	property:     string & =~"^.+$"
	value?:       string
	description?: string | null
	operator:     "true" | "false" | "present" | "notpresent"
} | {
	type:         "DATETIME_COMPARISON_TYPE"
	property:     string & =~"^.+$"
	value?:       string
	description?: string | null
	operator:     "eq" | "neq" | "present" | "notpresent" | "le" | "lte" | "gt" | "gte"
} | {
	type:         "ENTITY_ID_COMPARISON_TYPE"
	property:     string & =~"^.+$"
	value?:       string
	description?: string | null
	operator:     "eq" | "neq" | "empty" | "notempty" | "prefix" | "suffix" | "contains" | "notcontains"
} | {
	type:         "ENTITY_ID_COMPARISON_TYPE"
	property:     string & =~"^.+$"
	if list.Contains(#versions[:6], version) { // before 1.6
		value: string
	}
	if !list.Contains(#versions[:6], version) { // 1.6 or later
		value: [...string]
	}
	description?: string | null
	operator:     "isoneof" | "isnotoneof"
})
