package audit

import (
	"go.flipt.io/flipt/rpc/flipt"
)

// All types in this file represent an audit representation of the Flipt type that we will send to
// the different sinks.

type Flag struct {
	Key            string `json:"key"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Enabled        bool   `json:"enabled"`
	NamespaceKey   string `json:"namespace_key"`
	DefaultVariant string `json:"default_variant,omitempty"`
}

func NewFlag(f *flipt.Flag) *Flag {
	af := &Flag{
		Key:          f.Key,
		Name:         f.Name,
		Description:  f.Description,
		Enabled:      f.Enabled,
		NamespaceKey: f.NamespaceKey,
	}
	if f.DefaultVariant != nil {
		af.DefaultVariant = f.DefaultVariant.Key
	}
	return af
}

type Variant struct {
	Id           string `json:"id"`
	FlagKey      string `json:"flag_key"`
	Key          string `json:"key"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Attachment   string `json:"attachment"`
	NamespaceKey string `json:"namespace_key"`
}

func NewVariant(v *flipt.Variant) *Variant {
	return &Variant{
		Id:           v.Id,
		FlagKey:      v.FlagKey,
		Key:          v.Key,
		Name:         v.Name,
		Description:  v.Description,
		Attachment:   v.Attachment,
		NamespaceKey: v.NamespaceKey,
	}
}

type Constraint struct {
	Id           string `json:"id"`
	SegmentKey   string `json:"segment_key"`
	Type         string `json:"type"`
	Property     string `json:"property"`
	Operator     string `json:"operator"`
	Value        string `json:"value"`
	NamespaceKey string `json:"namespace_key"`
}

func NewConstraint(c *flipt.Constraint) *Constraint {
	return &Constraint{
		Id:           c.Id,
		SegmentKey:   c.SegmentKey,
		Type:         c.Type.String(),
		Property:     c.Property,
		Operator:     c.Operator,
		Value:        c.Value,
		NamespaceKey: c.NamespaceKey,
	}
}

type Namespace struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Protected   bool   `json:"protected"`
}

func NewNamespace(n *flipt.Namespace) *Namespace {
	return &Namespace{
		Key:         n.Key,
		Name:        n.Name,
		Description: n.Description,
		Protected:   n.Protected,
	}
}

type Distribution struct {
	Id        string  `json:"id"`
	RuleId    string  `json:"rule_id"`
	VariantId string  `json:"variant_id"`
	Rollout   float32 `json:"rollout"`
}

func NewDistribution(d *flipt.Distribution) *Distribution {
	return &Distribution{
		Id:        d.Id,
		RuleId:    d.RuleId,
		VariantId: d.VariantId,
		Rollout:   d.Rollout,
	}
}

type Segment struct {
	Key          string        `json:"key"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Constraints  []*Constraint `json:"constraints"`
	MatchType    string        `json:"match_type"`
	NamespaceKey string        `json:"namespace_key"`
}

func NewSegment(s *flipt.Segment) *Segment {
	c := make([]*Constraint, 0, len(s.Constraints))
	for _, sc := range s.Constraints {
		c = append(c, NewConstraint(sc))
	}

	return &Segment{
		Key:          s.Key,
		Name:         s.Name,
		Description:  s.Description,
		Constraints:  c,
		MatchType:    s.MatchType.String(),
		NamespaceKey: s.NamespaceKey,
	}
}

type Rule struct {
	Id            string          `json:"id"`
	FlagKey       string          `json:"flag_key"`
	SegmentKey    string          `json:"segment_key"`
	Distributions []*Distribution `json:"distributions"`
	Rank          int32           `json:"rank"`
	NamespaceKey  string          `json:"namespace_key"`
}

func NewRule(r *flipt.Rule) *Rule {
	d := make([]*Distribution, 0, len(r.Distributions))
	for _, rd := range r.Distributions {
		d = append(d, NewDistribution(rd))
	}

	return &Rule{
		Id:            r.Id,
		FlagKey:       r.FlagKey,
		SegmentKey:    r.SegmentKey,
		Distributions: d,
		Rank:          r.Rank,
		NamespaceKey:  r.NamespaceKey,
	}
}

type Rollout struct {
	NamespaceKey string            `json:"namespace_key"`
	FlagKey      string            `json:"flag_key"`
	Rank         int32             `json:"rank"`
	Description  string            `json:"description"`
	Threshold    *RolloutThreshold `json:"threshold,omitempty"`
	Segment      *RolloutSegment   `json:"segment,omitempty"`
}

type RolloutThreshold struct {
	Percentage float32 `json:"percentage"`
	Value      bool    `json:"value"`
}

type RolloutSegment struct {
	Key   string `json:"key"`
	Value bool   `json:"value"`
}

func NewRollout(r *flipt.Rollout) *Rollout {
	rollout := &Rollout{
		NamespaceKey: r.NamespaceKey,
		FlagKey:      r.FlagKey,
		Rank:         r.Rank,
		Description:  r.Description,
	}

	switch rout := r.Rule.(type) {
	case *flipt.Rollout_Segment:
		rollout.Segment = &RolloutSegment{
			Key:   rout.Segment.SegmentKey,
			Value: rout.Segment.Value,
		}
	case *flipt.Rollout_Threshold:
		rollout.Threshold = &RolloutThreshold{
			Percentage: rout.Threshold.Percentage,
			Value:      rout.Threshold.Value,
		}
	}

	return rollout
}
