package config

const (
	// configuration keys
	uiEnabled = "ui.enabled"
)

// UIConfig contains fields, which control the behaviour
// of Flipt's user interface.
type UIConfig struct {
	Enabled bool `json:"enabled"`
}
