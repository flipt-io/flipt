namespace: string & =~"^[-_,A-Za-z0-9]+$" | *"default"

flags: [...#Flag]

segments: [...#Segment]

#Flag: {
	key:         string & =~"^[-_,A-Za-z0-9]+$"
	name:        string & =~"^.+$"
	description?: string
	enabled:     bool | *false
	variants: [...#Variant]
	rules: [...#Rule]
}

#Variant: {
	key:        string & =~"^.+$"
	name:       string & =~"^.+$"
	attachment: {...} | *null
}

#Rule: {
	segment: string & =~"^.+$"
	rank:    int
	distributions: [...#Distribution]
}

#Distribution: {
	variant: string & =~"^.+$"
	rollout: >=0 & <=100
}

#Segment: {
	key:         string & =~"^[-_,A-Za-z0-9]+$"
	name:        string & =~"^.+$"
	match_type:  "ANY_MATCH_TYPE" | "ALL_MATCH_TYPE"
	description?: string
	constraints: [...#Constraint]
}

#Constraint: ({
	type:     "STRING_COMPARISON_TYPE"
	property: string & =~"^.+$"
	value?:   string 
	operator: "eq" | "neq" | "empty" | "notempty" | "prefix" | "suffix"
	description?: string 
} | {
	type:     "NUMBER_COMPARISON_TYPE"
	property: string & =~"^.+$"
	value?:   string 
	operator: "eq" | "neq" | "present" | "notpresent" | "le" | "lte" | "gt" | "gte"
	description?: string 
} | {
	type:     "BOOLEAN_COMPARISON_TYPE"
	property: string & =~"^.+$"
	value?:   string 
	operator: "true" | "false" | "present" | "notpresent"
	description?: string 
} | {
	type:     "DATETIME_COMPARISON_TYPE"
	property: string & =~"^.+$"
	value?:   string 
	operator: "eq" | "neq" | "present" | "notpresent" | "le" | "lte" | "gt" | "gte"
	description?: string 
})
