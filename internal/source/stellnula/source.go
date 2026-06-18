package stellnula

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	stn "github.com/stellhub/stellnula-go-sdk"
	"github.com/stellhub/stellorbit-go-sdk/governance"
)

type Source struct {
	client       *stn.Client
	parser       governance.SnapshotParser
	registry     atomic.Value
	registration stn.Registration
	failFast     bool
	logger       governance.Logger
}

type Options struct {
	Endpoint                 string
	GRPCEndpoint             string
	GRPCPlaintext            *bool
	APIToken                 string
	AppID                    string
	ClientID                 string
	Env                      string
	Region                   string
	Zone                     string
	Cluster                  string
	Namespace                string
	Group                    string
	WatchEnabled             *bool
	FailFastOnBootstrap      bool
	SnapshotDirectory        string
	Labels                   map[string]string
	AcceptLargeFileReference bool
	Logger                   governance.Logger
	HTTPClient               *http.Client
}

func New(options Options, clientOptions ...stn.ClientOption) (*Source, error) {
	if options.Endpoint == "" {
		return nil, fmt.Errorf("stellorbit: stellnula endpoint is required for governance rule source")
	}
	client, err := stn.NewClient(stn.Options{
		Endpoint:                 options.Endpoint,
		GRPCEndpoint:             options.GRPCEndpoint,
		GRPCPlaintext:            options.GRPCPlaintext,
		APIToken:                 options.APIToken,
		AppID:                    options.AppID,
		ClientID:                 options.ClientID,
		Env:                      options.Env,
		Region:                   options.Region,
		Zone:                     options.Zone,
		Cluster:                  options.Cluster,
		Namespace:                options.Namespace,
		Group:                    options.Group,
		WatchEnabled:             options.WatchEnabled,
		FailFastOnBootstrap:      options.FailFastOnBootstrap,
		SnapshotDirectory:        options.SnapshotDirectory,
		Labels:                   options.Labels,
		Logger:                   options.Logger,
		HTTPClient:               options.HTTPClient,
		AcceptLargeFileReference: options.AcceptLargeFileReference,
	}, clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("create stellnula governance rule client: %w", err)
	}
	return NewWithClient(client, options.FailFastOnBootstrap, options.Logger), nil
}

func NewWithClient(client *stn.Client, failFast bool, logger governance.Logger) *Source {
	if client == nil {
		panic("stellorbit: stellnula client must not be nil")
	}
	source := &Source{
		client:   client,
		parser:   governance.NewSnapshotParser(governance.NewParser(), logger),
		failFast: failFast,
		logger:   logger,
	}
	source.registry.Store(governance.EmptyRegistry())
	return source
}

func (s *Source) Start(ctx context.Context) error {
	if s.client == nil {
		return fmt.Errorf("stellorbit: stellnula client must not be nil")
	}
	s.registration = s.client.Listen(func(snapshot stn.Snapshot, _ []stn.ConfigChange) {
		s.replaceRegistry(snapshot)
	})
	if err := s.client.Start(ctx); err != nil {
		s.replaceRegistry(s.client.Snapshot())
		if s.failFast {
			return fmt.Errorf("start stellnula governance rule source: %w", err)
		}
		s.logf("stellorbit: start stellnula governance rule source failed, keep current registry: %v", err)
		return nil
	}
	s.replaceRegistry(s.client.Snapshot())
	return nil
}

func (s *Source) Registry() governance.Registry {
	return governance.LoadRegistry(&s.registry)
}

func (s *Source) Close() error {
	var closeErr error
	if s.registration != nil {
		closeErr = s.registration.Close()
	}
	if s.client == nil {
		return closeErr
	}
	if err := s.client.Close(); err != nil && closeErr == nil {
		closeErr = err
	}
	return closeErr
}

func (s *Source) replaceRegistry(snapshot stn.Snapshot) {
	previous := governance.LoadRegistry(&s.registry)
	if snapshot.Revision < previous.Revision {
		s.logf(
			"stellorbit: ignore stale governance snapshot revision %d, current revision is %d",
			snapshot.Revision,
			previous.Revision,
		)
		return
	}
	next := s.parser.Parse(toGovernanceSnapshot(snapshot), previous)
	s.registry.Store(next)
}

func (s *Source) logf(format string, args ...any) {
	if s.logger != nil {
		s.logger.Printf(format, args...)
	}
}

func toGovernanceSnapshot(snapshot stn.Snapshot) governance.Snapshot {
	entries := make([]governance.Entry, 0, len(snapshot.Entries))
	for _, entry := range snapshot.Entries {
		entries = append(entries, governance.Entry{
			ConfigID:    entry.ConfigID,
			ConfigKey:   entry.ConfigKey,
			ContentType: entry.ContentType,
			Value:       entry.ConfigValue(),
			Revision:    entry.Revision,
			Deleted:     entry.Deleted,
		})
	}
	return governance.Snapshot{
		Revision: snapshot.Revision,
		Checksum: snapshot.Checksum,
		Entries:  entries,
	}
}
