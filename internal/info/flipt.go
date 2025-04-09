package info

import (
	"encoding/json"
	"net/http"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/release"
)

func New(opts ...Option) Flipt {
	f := Flipt{}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func WithBuild(commit, date, goVersion, version string, isRelease bool) Option {
	return func(f *Flipt) {
		f.Commit = commit
		f.BuildDate = date
		f.GoVersion = goVersion
		f.IsRelease = isRelease
		f.Version = version
	}
}

func WithLatestRelease(releaseInfo release.Info) Option {
	return func(f *Flipt) {
		f.LatestVersion = releaseInfo.LatestVersion
		f.LatestVersionURL = releaseInfo.LatestVersionURL
		f.UpdateAvailable = releaseInfo.UpdateAvailable
	}
}

func WithOS(os, arch string) Option {
	return func(f *Flipt) {
		f.OS = os
		f.Arch = arch
	}
}

func WithConfig(cfg *config.Config) Option {
	return func(f *Flipt) {
		f.Authentication = authentication{Required: cfg.Authentication.Required}
		f.Storage = storage{Type: cfg.Storage.Type, ReadOnly: cfg.Storage.ReadOnly != nil && *cfg.Storage.ReadOnly, Metadata: cfg.Storage.Info()}
		f.Analytics = &analytics{Enabled: cfg.Analytics.Enabled()}
		f.UI = &ui{Theme: cfg.UI.DefaultTheme, TopbarColor: cfg.UI.Topbar.Color}
	}
}

type Option func(f *Flipt)

type authentication struct {
	Required bool `json:"required"`
}

type analytics struct {
	Enabled bool `json:"enabled,omitempty"`
}

type storage struct {
	Type     config.StorageType `json:"type"`
	ReadOnly bool               `json:"readOnly,omitempty"`
	Metadata map[string]string  `json:"metadata,omitempty"`
}

type ui struct {
	Theme       config.UITheme `json:"theme,omitempty"`
	TopbarColor string         `json:"topbarColor,omitempty"`
}

type Flipt struct {
	Version          string         `json:"version,omitempty"`
	LatestVersion    string         `json:"latestVersion,omitempty"`
	LatestVersionURL string         `json:"latestVersionURL,omitempty"`
	Commit           string         `json:"commit,omitempty"`
	BuildDate        string         `json:"buildDate,omitempty"`
	GoVersion        string         `json:"goVersion,omitempty"`
	UpdateAvailable  bool           `json:"updateAvailable"`
	IsRelease        bool           `json:"isRelease"`
	OS               string         `json:"os,omitempty"`
	Arch             string         `json:"arch,omitempty"`
	Authentication   authentication `json:"authentication"`
	Storage          storage        `json:"storage"`
	Analytics        *analytics     `json:"analytics,omitempty"`
	UI               *ui            `json:"ui,omitempty"`
}

func (f Flipt) IsDevelopment() bool {
	return f.Version == "dev"
}

func (f Flipt) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		out []byte
		err error
	)

	if r.Header.Get("Accept") == "application/json+pretty" {
		out, err = json.MarshalIndent(f, "", "  ")
	} else {
		out, err = json.Marshal(f)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(out); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
