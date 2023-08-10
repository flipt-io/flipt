package ext

import (
	"errors"
)

type Document struct {
	Version   string     `yaml:"version,omitempty"`
	Namespace string     `yaml:"namespace,omitempty"`
	Flags     []*Flag    `yaml:"flags,omitempty"`
	Segments  []*Segment `yaml:"segments,omitempty"`
}

type Flag struct {
	Key         string     `yaml:"key,omitempty"`
	Name        string     `yaml:"name,omitempty"`
	Type        string     `yaml:"type,omitempty"`
	Description string     `yaml:"description,omitempty"`
	Enabled     bool       `yaml:"enabled"`
	Variants    []*Variant `yaml:"variants,omitempty"`
	Rules       []*Rule    `yaml:"rules,omitempty"`
	Rollouts    []*Rollout `yaml:"rollouts,omitempty"`
}

type Variant struct {
	Key         string      `yaml:"key,omitempty"`
	Name        string      `yaml:"name,omitempty"`
	Description string      `yaml:"description,omitempty"`
	Attachment  interface{} `yaml:"attachment,omitempty"`
}

type Rule struct {
	Segment       *SegmentEmbed   `yaml:"segment,omitempty"`
	Rank          uint            `yaml:"rank,omitempty"`
	Distributions []*Distribution `yaml:"distributions,omitempty"`
}

type Distribution struct {
	VariantKey string  `yaml:"variant,omitempty"`
	Rollout    float32 `yaml:"rollout,omitempty"`
}

type Rollout struct {
	Description string         `yaml:"description,omitempty"`
	Segment     *SegmentRule   `yaml:"segment,omitempty"`
	Threshold   *ThresholdRule `yaml:"threshold,omitempty"`
}

type SegmentRule struct {
	Key      string   `yaml:"key,omitempty"`
	Keys     []string `yaml:"keys,omitempty"`
	Operator string   `yaml:"operator,omitempty"`
	Value    bool     `yaml:"value,omitempty"`
}

type ThresholdRule struct {
	Percentage float32 `yaml:"percentage,omitempty"`
	Value      bool    `yaml:"value,omitempty"`
}

type Segment struct {
	Key         string        `yaml:"key,omitempty"`
	Name        string        `yaml:"name,omitempty"`
	Description string        `yaml:"description,omitempty"`
	Constraints []*Constraint `yaml:"constraints,omitempty"`
	MatchType   string        `yaml:"match_type,omitempty"`
}

type Constraint struct {
	Type        string `yaml:"type,omitempty"`
	Property    string `yaml:"property,omitempty"`
	Operator    string `yaml:"operator,omitempty"`
	Value       string `yaml:"value,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type SegmentEmbed struct {
	IsSegment `yaml:"-"`
}

// MarshalYAML tries to type assert to either of the following types that implement
// IsSegment, and returns the marshaled value.
func (s *SegmentEmbed) MarshalYAML() (interface{}, error) {
	switch t := s.IsSegment.(type) {
	case SegmentKey:
		return string(t), nil
	case *Segments:
		sk := &Segments{
			Keys:            t.Keys,
			SegmentOperator: t.SegmentOperator,
		}
		return sk, nil
	}

	return nil, errors.New("failed to marshal to string or segmentKeys")
}

// UnmarshalYAML attempts to unmarshal a string or `SegmentKeys`, and fails if it can not
// do so.
func (s *SegmentEmbed) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sk SegmentKey

	if err := unmarshal(&sk); err == nil {
		s.IsSegment = sk
		return nil
	}

	var sks *Segments
	if err := unmarshal(&sks); err == nil {
		s.IsSegment = sks
		return nil
	}

	return errors.New("failed to unmarshal to string or segmentKeys")
}

// IsSegment is used to unify the two types of segments that can come in
// from the import.
type IsSegment interface {
	IsSegment()
}

type SegmentKey string

func (s SegmentKey) IsSegment() {}

type Segments struct {
	Keys            []string `yaml:"keys,omitempty"`
	SegmentOperator string   `yaml:"operator,omitempty"`
}

func (s *Segments) IsSegment() {}
