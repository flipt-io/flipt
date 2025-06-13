package info

import (
	"encoding/json"
	"net/http"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/product"
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
		if f.Build == nil {
			f.Build = &Build{}
		}

		f.Build.Commit = commit
		f.Build.BuildDate = date
		f.Build.GoVersion = goVersion
		f.Build.IsRelease = isRelease
		f.Build.Version = version
	}
}

func WithLatestRelease(releaseInfo release.Info) Option {
	return func(f *Flipt) {
		if f.Build == nil {
			f.Build = &Build{}
		}

		f.Build.LatestVersion = releaseInfo.LatestVersion
		f.Build.LatestVersionURL = releaseInfo.LatestVersionURL
		f.Build.UpdateAvailable = releaseInfo.UpdateAvailable
	}
}

func WithConfig(cfg *config.Config) Option {
	return func(f *Flipt) {
		f.Authentication = &Authentication{Required: cfg.Authentication.Required}
		f.Analytics = &Analytics{Enabled: cfg.Analytics.Enabled()}
		f.UI = &UI{Theme: cfg.UI.DefaultTheme, TopbarColor: cfg.UI.Topbar.Color}
	}
}

func WithLicenseManager(licenseManager interface{ Product() product.Product }) Option {
	return func(f *Flipt) {
		f.licenseManager = licenseManager
	}
}

type Option func(f *Flipt)

type Build struct {
	Version          string `json:"version,omitempty"`
	LatestVersion    string `json:"latestVersion,omitempty"`
	LatestVersionURL string `json:"latestVersionURL,omitempty"`
	Commit           string `json:"commit,omitempty"`
	BuildDate        string `json:"buildDate,omitempty"`
	GoVersion        string `json:"goVersion,omitempty"`
	UpdateAvailable  bool   `json:"updateAvailable"`
	IsRelease        bool   `json:"isRelease"`
}

type Authentication struct {
	Required bool `json:"required"`
}

type Analytics struct {
	Enabled bool `json:"enabled,omitempty"`
}

type UI struct {
	Theme       config.UITheme `json:"theme,omitempty"`
	TopbarColor string         `json:"topbarColor,omitempty"`
}

type Flipt struct {
	licenseManager interface{ Product() product.Product }
	Build          *Build          `json:"build,omitempty"`
	Authentication *Authentication `json:"authentication,omitempty"`
	Analytics      *Analytics      `json:"analytics,omitempty"`
	UI             *UI             `json:"ui,omitempty"`
}

func (f Flipt) IsDevelopment() bool {
	return f.Build.Version == "dev"
}

func (f Flipt) Product() product.Product {
	return f.licenseManager.Product()
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

// MarshalJSON implements custom JSON marshaling for Flipt to include the dynamic product field.
func (f Flipt) MarshalJSON() ([]byte, error) {
	type Alias Flipt // Prevent recursion
	aux := struct {
		Alias
		Product product.Product `json:"product"`
	}{
		Alias:   (Alias)(f),
		Product: f.Product(),
	}
	return json.Marshal(aux)
}
