package config

import (
	"fmt"
	"text/template"
)

// TemplatesConfig contains templates for commit messages and pull request content.
// These templates are used when creating commits and pull requests via the Flipt UI.
// If specified, these templates override the defaults but can still be overridden
// by repository-level templates in flipt.yaml.
type TemplatesConfig struct {
	CommitMessage string `json:"commitMessage,omitempty" mapstructure:"commit_message" yaml:"commit_message,omitempty"`
	ProposalTitle string `json:"proposalTitle,omitempty" mapstructure:"proposal_title" yaml:"proposal_title,omitempty"`
	ProposalBody  string `json:"proposalBody,omitempty" mapstructure:"proposal_body" yaml:"proposal_body,omitempty"`
}

// Validate checks that all configured templates are valid Go text/templates.
// This provides early validation at config load time rather than at runtime.
func (c *TemplatesConfig) Validate() error {
	if c.CommitMessage != "" {
		if _, err := template.New("commit_message").Parse(c.CommitMessage); err != nil {
			return fmt.Errorf("invalid commit_message template: %w", err)
		}
	}
	if c.ProposalTitle != "" {
		if _, err := template.New("proposal_title").Parse(c.ProposalTitle); err != nil {
			return fmt.Errorf("invalid proposal_title template: %w", err)
		}
	}
	if c.ProposalBody != "" {
		if _, err := template.New("proposal_body").Parse(c.ProposalBody); err != nil {
			return fmt.Errorf("invalid proposal_body template: %w", err)
		}
	}
	return nil
}
