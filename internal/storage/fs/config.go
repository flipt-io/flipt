package fs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"text/template"

	"github.com/blang/semver/v4"
	"github.com/gobwas/glob"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const (
	configFileNameYAML = "flipt.yaml"
	configFileNameYML  = "flipt.yml"
)

var (
	configVersion = semver.Version{Major: 2, Minor: 0}

	defaultCommitMsgTmpl string = `{{- if eq (len .Changes) 1 }}
{{- (index .Changes 0) }}
{{- else -}}
updated multiple resources
{{ range $change := .Changes }}
{{ $change }}
{{- end }}
{{- end }}`

	defaultProposalTitleTmpl string = `Flipt: Update features {{with .Base.Directory}}in {{.}} {{end}}on {{.Base.Ref}}`
	defaultProposalBodyTmpl  string = `This pull request updates Flipt resources {{with .Base.Directory}}in {{.}} {{end}}on branch {{.Base.Ref}}.

🟢 **Source:**{{if .Branch.Directory}}
- Directory: {{.Branch.Directory}}{{end}}
- Branch: {{.Branch.Ref}}

🎯 **Target:**{{if .Base.Directory}}
- Directory: {{.Base.Directory}}{{end}}
- Branch: {{.Base.Ref}}

👀 Please review the changes and merge if everything looks good.`
)

type Config struct {
	Matchers  []glob.Glob
	Templates ConfigTemplates
}

type ConfigTemplates struct {
	CommitMessageTemplate *template.Template
	ProposalTitleTemplate *template.Template
	ProposalBodyTemplate  *template.Template
}

// DefaultFliptConfig returns the default value for the Config struct.
// It used when a flipt.yml cannot be located.
func DefaultFliptConfig() *Config {
	return &Config{
		Matchers: []glob.Glob{
			// must end in either yaml, yml or json
			// must be nested a single directory below the root
			glob.MustCompile("*/features.yaml"),
			glob.MustCompile("*/features.yml"),
			glob.MustCompile("*/features.json"),
		},
		Templates: ConfigTemplates{
			CommitMessageTemplate: template.Must(template.New("commitMessage").Parse(defaultCommitMsgTmpl)),
			ProposalTitleTemplate: template.Must(template.New("proposalTitle").Parse(defaultProposalTitleTmpl)),
			ProposalBodyTemplate:  template.Must(template.New("proposalBody").Parse(defaultProposalBodyTmpl)),
		},
	}
}

// GetConfig supports opening and parsing flipt configuration within a target filesystem.
// It initially attempts to parse the broader flipt.yml configuration file.
// Failing to locate this, it falls back to parsing the .flipt.yml index file.
func GetConfig(logger *zap.Logger, src fs.FS) (*Config, error) {
	fi, err := src.Open(configFileNameYAML)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}

		fi, err = src.Open(configFileNameYML)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			}

			return DefaultFliptConfig(), nil
		}
	}

	defer fi.Close()

	return parseConfig(logger, fi)
}

func (c *Config) List(src fs.FS) (paths []string, err error) {
	err = fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		for _, matcher := range c.Matchers {
			if matcher.Match(path) {
				paths = append(paths, path)
				break
			}
		}

		return nil
	})

	// we ignore not exist errors and treat them as
	// returning an empty result
	if errors.Is(err, fs.ErrNotExist) {
		err = nil
	}

	return
}

type config struct {
	Version   string `yaml:"version"`
	Templates struct {
		CommitMsg     string `yaml:"commit_message"`
		ProposalTitle string `yaml:"proposal_title"`
		ProposalBody  string `yaml:"proposal_body"`
	} `yaml:"templates"`
}

// parseConfig reads the contents of r as yaml and parses
// the configuration with some predefined defaults
func parseConfig(logger *zap.Logger, r io.Reader) (_ *Config, err error) {
	conf := config{Version: defaultConfigVersion()}
	if err := yaml.NewDecoder(r).Decode(&conf); err != nil {
		return nil, err
	}

	v, err := semver.Parse(conf.Version)
	if err != nil {
		return nil, fmt.Errorf("parsing %s version %q: %w", configFileNameYAML, conf.Version, err)
	}

	if v.GT(configVersion) {
		return nil, fmt.Errorf("unsupported flipt config version: %q", v)
	}

	c := DefaultFliptConfig()
	if conf.Templates.CommitMsg != "" {
		tmpl, err := template.New("commitMessage").Parse(conf.Templates.CommitMsg)
		if err != nil {
			return nil, err
		}

		c.Templates.CommitMessageTemplate = tmpl
	}

	if conf.Templates.ProposalTitle != "" {
		tmpl, err := template.New("proposalTitle").Parse(conf.Templates.ProposalTitle)
		if err != nil {
			logger.Warn("failed to parse template", zap.String("template", "proposalTitle"), zap.Error(err))
		} else {
			c.Templates.ProposalTitleTemplate = tmpl
		}
	}

	if conf.Templates.ProposalBody != "" {
		tmpl, err := template.New("proposalBody").Parse(conf.Templates.ProposalBody)
		if err != nil {
			logger.Warn("failed to parse template", zap.String("template", "proposalBody"), zap.Error(err))
		} else {
			c.Templates.ProposalBodyTemplate = tmpl
		}
	}

	return c, nil
}

func defaultConfigVersion() string {
	return fmt.Sprintf("%d.%d", configVersion.Major, configVersion.Minor)
}
