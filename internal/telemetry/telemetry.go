package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/uuid"
	"github.com/markphelps/flipt/config"
	"github.com/markphelps/flipt/internal/info"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"gopkg.in/segmentio/analytics-go.v3"
)

const (
	file    = "telemetry.json"
	version = "1.0"
)

type ping struct {
	Version string `json:"version"`
	UUID    string `json:"uuid"`
	Flipt   flipt  `json:"flipt"`
}

type flipt struct {
	Version string `json:"version"`
}

type state struct {
	Version       string `json:"version"`
	UUID          string `json:"uuid"`
	LastTimestamp string `json:"lastTimestamp"`
}

type Reporter struct {
	cfg    config.Config
	logger logrus.FieldLogger
	client analytics.Client
}

func NewReporter(cfg config.Config, logger logrus.FieldLogger, analytics analytics.Client) *Reporter {
	return &Reporter{
		cfg:    cfg,
		logger: logger,
		client: analytics,
	}
}

func (r *Reporter) Report(ctx context.Context, info info.Flipt) (err error) {
	s, err := r.readState(ctx)
	if err != nil {
		return fmt.Errorf("reading state: %w", err)
	}

	t, _ := time.Parse(time.RFC3339, s.LastTimestamp)

	r.logger.Debugf("last report was: %v ago", time.Since(t))

	var (
		props = analytics.NewProperties()
		p     = ping{
			Version: s.Version,
			UUID:    s.UUID,
			Flipt: flipt{
				Version: info.Version,
			},
		}
	)

	if err := mapstructure.Decode(p, &props); err != nil {
		return fmt.Errorf("decoding properties: %w", err)
	}

	if err := r.client.Enqueue(analytics.Track{
		AnonymousId: s.UUID,
		Event:       "flipt.ping",
		Properties:  props,
	}); err != nil {
		return fmt.Errorf("tracking ping: %w", err)
	}

	s.LastTimestamp = time.Now().UTC().Format(time.RFC3339)

	if err := r.writeState(ctx, s); err != nil {
		return fmt.Errorf("writing state: %w", err)
	}

	return nil
}

func (r *Reporter) Close() error {
	return r.client.Close()
}

func (r *Reporter) readState(_ context.Context) (state, error) {
	f, err := os.Open(filepath.Join(r.cfg.Meta.StateDirectory, file))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return newState(), nil
		}

		return state{}, fmt.Errorf("opening state file: %w", err)
	}

	defer f.Close()

	var s state
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return state{}, fmt.Errorf("decoding state file: %w", err)
	}

	return s, nil
}

func (r *Reporter) writeState(_ context.Context, s state) error {
	b, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	//nolint:gosec
	if err := os.WriteFile(filepath.Join(r.cfg.Meta.StateDirectory, file), b, 0644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}
	return nil
}

func newState() state {
	var uid string

	u, err := uuid.NewV4()
	if err != nil {
		uid = "unknown"
	} else {
		uid = u.String()
	}

	return state{
		Version: version,
		UUID:    uid,
	}
}
