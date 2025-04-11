package ext

import (
	"encoding/json"
	"errors"

	"github.com/blang/semver/v4"
)

var (
	// LatestVersion is the current latest supported export format for flag state
	LatestVersion = v1_5
	v1_5          = semver.Version{Major: 1, Minor: 5}
)

type Document struct {
	Version   string          `yaml:"version,omitempty" json:"version,omitempty"`
	Namespace *NamespaceEmbed `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Flags     []*Flag         `yaml:"flags,omitempty" json:"flags,omitempty"`
	Segments  []*Segment      `yaml:"segments,omitempty" json:"segments,omitempty"`
	Etag      string          `yaml:"-" json:"-"`
}

type Flag struct {
	Key         string         `yaml:"key,omitempty" json:"key,omitempty"`
	Name        string         `yaml:"name,omitempty" json:"name,omitempty"`
	Type        string         `yaml:"type,omitempty" json:"type,omitempty"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Enabled     bool           `yaml:"enabled" json:"enabled"`
	Metadata    map[string]any `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Variants    []*Variant     `yaml:"variants,omitempty" json:"variants,omitempty"`
	Rules       []*Rule        `yaml:"rules,omitempty" json:"rules,omitempty"`
	Rollouts    []*Rollout     `yaml:"rollouts,omitempty" json:"rollouts,omitempty"`
}

type Variant struct {
	Default     bool   `yaml:"default,omitempty" json:"default,omitempty"`
	Key         string `yaml:"key,omitempty" json:"key,omitempty"`
	Name        string `yaml:"name,omitempty" json:"name,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Attachment  any    `yaml:"attachment,omitempty" json:"attachment,omitempty"`
}

type Rule struct {
	Segment       *SegmentRule    `yaml:"segment,omitempty" json:"segment,omitempty"`
	Rank          uint            `yaml:"rank,omitempty" json:"rank,omitempty"`
	Distributions []*Distribution `yaml:"distributions,omitempty" json:"distributions,omitempty"`
}

type Distribution struct {
	VariantKey string  `yaml:"variant,omitempty" json:"variant,omitempty"`
	Rollout    float32 `yaml:"rollout,omitempty" json:"rollout,omitempty"`
}

type Rollout struct {
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Segment     *SegmentRule   `yaml:"segment,omitempty" json:"segment,omitempty"`
	Threshold   *ThresholdRule `yaml:"threshold,omitempty" json:"threshold,omitempty"`
}

type SegmentRule struct {
	Keys     []string `yaml:"-" json:"-"` // internal storage for keys for backwards compatibility.
	Operator string   `yaml:"operator,omitempty" json:"operator,omitempty"`
	Value    bool     `yaml:"value,omitempty" json:"value,omitempty"`
}

// MarshalYAML implements yaml.Marshaler
func (s *SegmentRule) MarshalYAML() (any, error) {
	type Alias SegmentRule
	return struct {
		*Alias
		Keys []string `yaml:"keys,omitempty"`
	}{
		Alias: (*Alias)(s),
		Keys:  s.Keys,
	}, nil
}

// UnmarshalYAML implements yaml.Unmarshaler
func (s *SegmentRule) UnmarshalYAML(unmarshal func(any) error) error {
	// Try full object with keys first
	aux := struct {
		Key      string   `yaml:"key"`
		Keys     []string `yaml:"keys"`
		Operator string   `yaml:"operator"`
		Value    bool     `yaml:"value"`
	}{}

	// Try to unmarshal the full object first
	if err := unmarshal(&aux); err == nil {
		*s = SegmentRule{
			Operator: aux.Operator,
			Value:    aux.Value,
		}

		// Handle the key/keys fields
		if len(aux.Keys) > 0 {
			s.Keys = aux.Keys
		} else if aux.Key != "" {
			s.Keys = []string{aux.Key}
		}
		return nil
	}

	// Try single key string
	var key string
	if err := unmarshal(&key); err == nil {
		s.Keys = []string{key}
		return nil
	}

	// Try array of keys
	var keys []string
	if err := unmarshal(&keys); err == nil {
		s.Keys = keys
		return nil
	}

	return errors.New("failed to unmarshal segment rule")
}

// MarshalJSON implements json.Marshaler
func (s *SegmentRule) MarshalJSON() ([]byte, error) {
	type Alias SegmentRule
	return json.Marshal(struct {
		*Alias
		Keys []string `json:"keys,omitempty"`
	}{
		Alias: (*Alias)(s),
		Keys:  s.Keys,
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (s *SegmentRule) UnmarshalJSON(data []byte) error {
	aux := struct {
		Key      string   `json:"key"`
		Keys     []string `json:"keys"`
		Operator string   `json:"operator"`
		Value    bool     `json:"value"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	*s = SegmentRule{
		Operator: aux.Operator,
		Value:    aux.Value,
	}

	// Handle the key/keys fields
	if len(aux.Keys) > 0 {
		s.Keys = aux.Keys
	} else if aux.Key != "" {
		s.Keys = []string{aux.Key}
	}

	return nil
}

type ThresholdRule struct {
	Percentage float32 `yaml:"percentage,omitempty" json:"percentage,omitempty"`
	Value      bool    `yaml:"value,omitempty" json:"value,omitempty"`
}

type Segment struct {
	Key         string        `yaml:"key,omitempty" json:"key,omitempty"`
	Name        string        `yaml:"name,omitempty" json:"name,omitempty"`
	Description string        `yaml:"description,omitempty" json:"description,omitempty"`
	Constraints []*Constraint `yaml:"constraints,omitempty" json:"constraints,omitempty"`
	MatchType   string        `yaml:"match_type,omitempty" json:"match_type,omitempty"`
}

type Constraint struct {
	Type        string `yaml:"type,omitempty" json:"type,omitempty"`
	Property    string `yaml:"property,omitempty" json:"property,omitempty"`
	Operator    string `yaml:"operator,omitempty" json:"operator,omitempty"`
	Value       string `yaml:"value,omitempty" json:"value,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// IsNamespace is used to unify the two types of namespaces that can come in
// from the import.
type IsNamespace interface {
	IsNamespace()
	GetKey() string
}

type NamespaceEmbed struct {
	IsNamespace `yaml:"-"`
}

var DefaultNamespace = &NamespaceEmbed{&Namespace{Key: "default", Name: "Default"}}

func (n *NamespaceEmbed) String() string {
	return n.IsNamespace.GetKey()
}

// MarshalYAML tries to type assert to either of the following types that implement
// IsNamespace, and returns the marshaled value.
func (n *NamespaceEmbed) MarshalYAML() (any, error) {
	switch t := n.IsNamespace.(type) {
	case NamespaceKey:
		return string(t), nil
	case *Namespace:
		ns := &Namespace{
			Key:         t.Key,
			Name:        t.Name,
			Description: t.Description,
		}
		return ns, nil
	}

	return nil, errors.New("failed to marshal to string or namespace")
}

// UnmarshalYAML attempts to unmarshal a string or `Namespace`, and fails if it can not
// do so.
func (n *NamespaceEmbed) UnmarshalYAML(unmarshal func(any) error) error {
	var nk NamespaceKey

	if err := unmarshal(&nk); err == nil {
		n.IsNamespace = nk
		return nil
	}

	var ns *Namespace
	if err := unmarshal(&ns); err == nil {
		n.IsNamespace = ns
		return nil
	}

	return errors.New("failed to unmarshal to string or namespace")
}

// MarshalJSON tries to type assert to either of the following types that implement
// IsNamespace, and returns the marshaled value.
func (n *NamespaceEmbed) MarshalJSON() ([]byte, error) {
	switch t := n.IsNamespace.(type) {
	case NamespaceKey:
		return json.Marshal(string(t))
	case *Namespace:
		ns := &Namespace{
			Key:         t.Key,
			Name:        t.Name,
			Description: t.Description,
		}
		return json.Marshal(ns)
	}

	return nil, errors.New("failed to marshal to string or namespace")
}

// UnmarshalJSON attempts to unmarshal a string or `Namespace`, and fails if it can not
// do so.
func (n *NamespaceEmbed) UnmarshalJSON(v []byte) error {
	var nk NamespaceKey

	if err := json.Unmarshal(v, &nk); err == nil {
		n.IsNamespace = nk
		return nil
	}

	var ns *Namespace
	if err := json.Unmarshal(v, &ns); err == nil {
		n.IsNamespace = ns
		return nil
	}

	return errors.New("failed to unmarshal to string or namespace")
}

type NamespaceKey string

func (n NamespaceKey) IsNamespace() {}

func (n NamespaceKey) GetKey() string {
	return string(n)
}

type Namespace struct {
	Key         string `yaml:"key,omitempty" json:"key,omitempty"`
	Name        string `yaml:"name,omitempty" json:"name,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

func (n *Namespace) IsNamespace() {}

func (n *Namespace) GetKey() string {
	if n == nil {
		return ""
	}
	return n.Key
}
