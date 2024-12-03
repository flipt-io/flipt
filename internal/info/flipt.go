package info

import (
	"encoding/json"
	"net/http"

	"go.flipt.io/flipt/internal/config"
)

type Flipt struct {
	Version          string             `json:"version,omitempty"`
	LatestVersion    string             `json:"latestVersion,omitempty"`
	LatestVersionURL string             `json:"latestVersionURL,omitempty"`
	Commit           string             `json:"commit,omitempty"`
	BuildDate        string             `json:"buildDate,omitempty"`
	GoVersion        string             `json:"goVersion,omitempty"`
	UpdateAvailable  bool               `json:"updateAvailable"`
	IsRelease        bool               `json:"isRelease"`
	OS               string             `json:"os,omitempty"`
	Arch             string             `json:"arch,omitempty"`
	Authentication   bool               `json:"authentication"`
	Storage          config.StorageType `json:"storage,omitempty"`
	StorageInfo      map[string]string  `json:"storageMetadata,omitempty"`
	AnalyticsEnabled bool               `json:"analyticsEnabled,omitempty"`
	UITheme          string             `json:"uiTheme,omitempty"`
	UITopbarColor    string             `json:"uiTopbarColor,omitempty"`
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
