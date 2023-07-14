package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/uuid"
	"github.com/xo/dburl"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/info"
	"go.uber.org/zap"
	"gopkg.in/segmentio/analytics-go.v3"
)

const (
	filename = "telemetry.json"
	version  = "1.2"
	event    = "flipt.ping"
)

type ping struct {
	Version string `json:"version"`
	UUID    string `json:"uuid"`
	Flipt   flipt  `json:"flipt"`
}

type storage struct {
	Type     string `json:"type,omitempty"`
	Database string `json:"database,omitempty"`
	Cache    string `json:"cache,omitempty"`
}

type authentication struct {
	Methods []string `json:"methods,omitempty"`
}

type flipt struct {
	Version        string                    `json:"version"`
	Storage        *storage                  `json:"storage,omitempty"`
	Authentication *authentication           `json:"authentication,omitempty"`
	Experimental   config.ExperimentalConfig `json:"experimental,omitempty"`
}

type state struct {
	Version       string `json:"version"`
	UUID          string `json:"uuid"`
	LastTimestamp string `json:"lastTimestamp"`
}

type Reporter struct {
	cfg      config.Config
	logger   *zap.Logger
	client   analytics.Client
	info     info.Flipt
	shutdown chan struct{}
}

func NewReporter(cfg config.Config, logger *zap.Logger, analyticsKey string, info info.Flipt) (*Reporter, error) {
	// don't log from analytics package
	analyticsLogger := func() analytics.Logger {
		stdLogger := log.Default()
		stdLogger.SetOutput(io.Discard)
		return analytics.StdLogger(stdLogger)
	}

	client, err := analytics.NewWithConfig(analyticsKey, analytics.Config{
		BatchSize: 1,
		Logger:    analyticsLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("initializing telemetry client %w", err)
	}

	return &Reporter{
		cfg:      cfg,
		logger:   logger,
		client:   client,
		info:     info,
		shutdown: make(chan struct{}),
	}, nil
}

func (r *Reporter) Run(ctx context.Context) {
	var (
		reportInterval = 4 * time.Hour
		ticker         = time.NewTicker(reportInterval)
		failures       = 0
	)

	const maxFailures = 3

	defer ticker.Stop()

	r.logger.Debug("starting telemetry reporter")
	if err := r.report(ctx); err != nil {
		r.logger.Debug("reporting telemetry", zap.Error(err))
	}

	for {
		select {
		case <-ticker.C:
			if err := r.report(ctx); err != nil {
				r.logger.Debug("reporting telemetry", zap.Error(err))

				if failures++; failures >= maxFailures {
					r.logger.Debug("telemetry reporting failure threshold reached, shutting down")
					return
				}
			} else {
				failures = 0
			}
		case <-r.shutdown:
			ticker.Stop()
			return
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (r *Reporter) Shutdown() error {
	close(r.shutdown)
	return r.client.Close()
}

type file interface {
	io.ReadWriteSeeker
	Truncate(int64) error
}

// report sends a ping event to the analytics service.
func (r *Reporter) report(ctx context.Context) (err error) {
	f, err := os.OpenFile(filepath.Join(r.cfg.Meta.StateDirectory, filename), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("opening state file: %w", err)
	}
	defer f.Close()

	return r.ping(ctx, f)
}

// ping sends a ping event to the analytics service.
// visible for testing
func (r *Reporter) ping(_ context.Context, f file) error {
	if !r.cfg.Meta.TelemetryEnabled {
		return nil
	}

	var (
		info = r.info
		s    state
	)

	if err := json.NewDecoder(f).Decode(&s); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("reading state: %w", err)
	}

	// if s is empty we need to create a new state
	if s.UUID == "" {
		s = newState()
		r.logger.Debug("initialized new state")
	} else {
		t, _ := time.Parse(time.RFC3339, s.LastTimestamp)
		r.logger.Debug("last report", zap.Time("when", t), zap.Duration("elapsed", time.Since(t)))
	}

	var (
		props = analytics.NewProperties()
		flipt = flipt{
			Version:      info.Version,
			Experimental: r.cfg.Experimental,
		}
	)

	var dbProtocol = r.cfg.Database.Protocol.String()

	if dbProtocol == "" && r.cfg.Database.URL != "" {
		dbProtocol = "unknown"

		url, err := dburl.Parse(r.cfg.Database.URL)
		if err == nil {
			// just swallow the error, we don't want to fail telemetry reporting
			dbProtocol = url.Scheme
		}
	}

	flipt.Storage = &storage{
		Type:     string(r.cfg.Storage.Type),
		Database: dbProtocol,
	}

	// only report cache if enabled
	if r.cfg.Cache.Enabled {
		flipt.Storage.Cache = r.cfg.Cache.Backend.String()
	}

	// authentication
	methods := make([]string, 0, len(r.cfg.Authentication.Methods.EnabledMethods()))

	for _, m := range r.cfg.Authentication.Methods.EnabledMethods() {
		methods = append(methods, m.Name())
	}

	// only report authentications if enabled
	if len(methods) > 0 {
		flipt.Authentication = &authentication{
			Methods: methods,
		}
	}

	p := ping{
		Version: version,
		UUID:    s.UUID,
		Flipt:   flipt,
	}

	// marshal as json first so we can get the correct case field names in the analytics service
	out, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshaling ping: %w", err)
	}

	if err := json.Unmarshal(out, &props); err != nil {
		return fmt.Errorf("unmarshaling ping: %w", err)
	}

	if err := r.client.Enqueue(analytics.Track{
		AnonymousId: s.UUID,
		Event:       event,
		Properties:  props,
	}); err != nil {
		return fmt.Errorf("tracking ping: %w", err)
	}

	s.Version = version
	s.LastTimestamp = time.Now().UTC().Format(time.RFC3339)

	// reset the state file
	if err := f.Truncate(0); err != nil {
		return fmt.Errorf("truncating state file: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("resetting state file: %w", err)
	}

	if err := json.NewEncoder(f).Encode(s); err != nil {
		return fmt.Errorf("writing state: %w", err)
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
