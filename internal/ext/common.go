package ext

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/blang/semver/v4"
	"go.flipt.io/flipt/rpc/flipt"
)

var (
	// LatestVersion is the current latest supported export format for flag state
	LatestVersion = v1_6
	v1_6          = semver.Version{Major: 1, Minor: 6}
)

// LatestVersionString returns the latest supported version string
// in "major.minor" format (e.g. "1.6").
func LatestVersionString() string {
	version := LatestVersion.FinalizeVersion()
	if LatestVersion.Patch == 0 {
		version = fmt.Sprintf("%d.%d", LatestVersion.Major, LatestVersion.Minor)
	}
	return version
}

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
	Rollout    float32 `yaml:"rollout" json:"rollout"`
}

type Rollout struct {
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Segment     *SegmentRule   `yaml:"segment,omitempty" json:"segment,omitempty"`
	Threshold   *ThresholdRule `yaml:"threshold,omitempty" json:"threshold,omitempty"`
}

type SegmentRule struct {
	Keys     []string `yaml:"keys,omitempty" json:"keys,omitempty"`
	Operator string   `yaml:"operator,omitempty" json:"operator,omitempty"`
	Value    bool     `yaml:"value,omitempty" json:"value,omitempty"`
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
	Percentage float32 `yaml:"percentage" json:"percentage"`
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

// Must be kept in sync with Constraint fields.
type constraintYAMLList struct {
	Type        string `yaml:"type,omitempty"`
	Property    string `yaml:"property,omitempty"`
	Operator    string `yaml:"operator,omitempty"`
	Value       []any  `yaml:"value"`
	Description string `yaml:"description,omitempty"`
}

func (c Constraint) MarshalYAML() (any, error) {
	if c.Operator == "isoneof" || c.Operator == "isnotoneof" {
		list := []any{}
		if c.Value != "" && json.Unmarshal([]byte(c.Value), &list) == nil {
			slices.SortFunc(list, func(a, b any) int {
				return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
			})
		}

		return &constraintYAMLList{
			Type:        c.Type,
			Property:    c.Property,
			Operator:    c.Operator,
			Value:       list,
			Description: c.Description,
		}, nil
	}

	type alias Constraint
	return alias(c), nil
}

func (c *Constraint) UnmarshalYAML(unmarshal func(any) error) error {
	var aux struct {
		Type        string `yaml:"type"`
		Property    string `yaml:"property"`
		Operator    string `yaml:"operator"`
		Value       any    `yaml:"value"`
		Description string `yaml:"description"`
	}

	if err := unmarshal(&aux); err != nil {
		return err
	}

	c.Type = aux.Type
	c.Property = aux.Property
	c.Operator = aux.Operator
	c.Description = aux.Description

	if err := setConstraintValue(c, aux.Value); err != nil {
		return err
	}

	return nil
}

func (c Constraint) MarshalJSON() ([]byte, error) {
	if c.Operator == "isoneof" || c.Operator == "isnotoneof" {
		list := []any{}
		if c.Value != "" && json.Unmarshal([]byte(c.Value), &list) == nil {
			slices.SortFunc(list, func(a, b any) int {
				return strings.Compare(fmt.Sprint(a), fmt.Sprint(b))
			})
		}

		return json.Marshal(struct {
			Type        string `json:"type,omitempty"`
			Property    string `json:"property,omitempty"`
			Operator    string `json:"operator,omitempty"`
			Value       []any  `json:"value"`
			Description string `json:"description,omitempty"`
		}{
			Type:        c.Type,
			Property:    c.Property,
			Operator:    c.Operator,
			Value:       list,
			Description: c.Description,
		})
	}

	type alias Constraint
	return json.Marshal(alias(c))
}

func (c *Constraint) UnmarshalJSON(data []byte) error {
	var aux struct {
		Type        string `json:"type"`
		Property    string `json:"property"`
		Operator    string `json:"operator"`
		Value       any    `json:"value"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	c.Type = aux.Type
	c.Property = aux.Property
	c.Operator = aux.Operator
	c.Description = aux.Description

	if err := setConstraintValue(c, aux.Value); err != nil {
		return err
	}

	return nil
}

// setConstraintValue populates c.Value from the decoded value v, handling
// string, array, scalar, and nil forms. Shared by UnmarshalYAML and UnmarshalJSON.
func setConstraintValue(c *Constraint, v any) error {
	switch val := v.(type) {
	case string:
		c.Value = val
	case []any:
		// Coerce bool/nil (e.g. bare true/false/null in hand-written YAML) to
		// strings. Leave int/float64 alone — number arrays are valid for
		// NUMBER_COMPARISON_TYPE isoneof constraints.
		for i, elem := range val {
			switch e := elem.(type) {
			case string, int, float64:
			case bool:
				val[i] = strconv.FormatBool(e)
			case nil:
				val[i] = ""
			default:
				val[i] = fmt.Sprintf("%v", e)
			}
		}
		data, err := json.Marshal(val)
		if err != nil {
			return fmt.Errorf("marshaling constraint value list: %w", err)
		}
		c.Value = string(data)
	case int:
		c.Value = strconv.Itoa(val)
	case float64:
		c.Value = strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		c.Value = strconv.FormatBool(val)
	case nil:
		c.Value = ""
	default:
		c.Value = fmt.Sprintf("%v", val)
	}

	return nil
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

var DefaultNamespace = &NamespaceEmbed{&Namespace{Key: flipt.DefaultNamespace, Name: "Default"}}

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
