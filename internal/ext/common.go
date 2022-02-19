package ext

type Document struct {
	Flags    []*Flag    `yaml:"flags,omitempty"`
	Segments []*Segment `yaml:"segments,omitempty"`
}

type Flag struct {
	Key         string     `yaml:"key,omitempty"`
	Name        string     `yaml:"name,omitempty"`
	Description string     `yaml:"description,omitempty"`
	Enabled     bool       `yaml:"enabled"`
	Variants    []*Variant `yaml:"variants,omitempty"`
	Rules       []*Rule    `yaml:"rules,omitempty"`
}

type Variant struct {
	Key         string      `yaml:"key,omitempty"`
	Name        string      `yaml:"name,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Attachment  interface{} `yaml:"attachment,omitempty"`
}

type Rule struct {
	SegmentKey    string          `yaml:"segment,omitempty"`
	Rank          uint            `yaml:"rank,omitempty"`
	Distributions []*Distribution `yaml:"distributions,omitempty"`
}

type Distribution struct {
	VariantKey string  `yaml:"variant,omitempty"`
	Rollout    float32 `yaml:"rollout,omitempty"`
}

type Segment struct {
	Key         string        `yaml:"key,omitempty"`
	Name        string        `yaml:"name,omitempty"`
	Description string        `yaml:"description,omitempty"`
	Constraints []*Constraint `yaml:"constraints,omitempty"`
	MatchType   string        `yaml:"match_type,omitempty"`
}

type Constraint struct {
	Type     string `yaml:"type,omitempty"`
	Property string `yaml:"property,omitempty"`
	Operator string `yaml:"operator,omitempty"`
	Value    string `yaml:"value,omitempty"`
}
