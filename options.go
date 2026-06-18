package stellorbit

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	stellnula "github.com/stellhub/stellnula-go-sdk"
	"github.com/stellhub/stellorbit-go-sdk/governance"
)

const (
	defaultTimeout           = 10 * time.Second
	defaultAppID             = "stellorbit-go-sdk"
	defaultEnv               = "dev"
	defaultScopeValue        = "default"
	defaultRuleNamespace     = "governance"
	defaultRuleGroup         = "service-governance"
	apiKeyHeader             = "X-Stellorbit-Api-Key"
	contentTypeHeader        = "Content-Type"
	acceptHeader             = "Accept"
	userAgentHeader          = "User-Agent"
	applicationJSON          = "application/json"
	defaultUserAgent         = "stellorbit-go-sdk"
	stellorbitClientIDPrefix = "stellorbit-"
)

var (
	ErrEndpointRequired    = errors.New("stellorbit: endpoint or stellnula endpoint is required")
	ErrServiceNameRequired = governance.ErrServiceNameRequired
)

type Logger = governance.Logger

type Options struct {
	Endpoint                 string
	APIKey                   string
	Timeout                  time.Duration
	HTTPClient               *http.Client
	StellnulaEndpoint        string
	StellnulaGRPCEndpoint    string
	StellnulaGRPCPlaintext   *bool
	StellnulaAPIToken        string
	AppID                    string
	ClientID                 string
	Env                      string
	Region                   string
	Zone                     string
	Cluster                  string
	RuleNamespace            string
	RuleGroup                string
	WatchEnabled             *bool
	FailFastOnBootstrap      bool
	SnapshotDirectory        string
	Labels                   map[string]string
	AcceptLargeFileReference bool
	Logger                   Logger
}

type Option func(*clientSettings)

type clientSettings struct {
	ruleSource             governance.Source
	stellnulaClientOptions []stellnula.ClientOption
}

func WithRuleSource(source governance.Source) Option {
	return func(settings *clientSettings) {
		settings.ruleSource = source
	}
}

func WithStellnulaClientOptions(options ...stellnula.ClientOption) Option {
	return func(settings *clientSettings) {
		settings.stellnulaClientOptions = append(settings.stellnulaClientOptions, options...)
	}
}

func Bool(value bool) *bool {
	return &value
}

func (o Options) normalize(allowNoEndpoint bool) (Options, error) {
	o.Endpoint = strings.TrimRight(strings.TrimSpace(o.Endpoint), "/")
	o.StellnulaEndpoint = strings.TrimRight(strings.TrimSpace(o.StellnulaEndpoint), "/")
	o.StellnulaGRPCEndpoint = strings.TrimSpace(o.StellnulaGRPCEndpoint)
	if o.Endpoint == "" && o.StellnulaEndpoint == "" && !allowNoEndpoint {
		return Options{}, ErrEndpointRequired
	}
	if err := validateEndpoint("endpoint", o.Endpoint); err != nil {
		return Options{}, err
	}
	if err := validateEndpoint("stellnula endpoint", o.StellnulaEndpoint); err != nil {
		return Options{}, err
	}
	if o.Timeout <= 0 {
		o.Timeout = defaultTimeout
	}
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{Timeout: o.Timeout}
	}
	o.AppID = defaultText(o.AppID, defaultAppID)
	o.ClientID = defaultText(o.ClientID, stellorbitClientIDPrefix+time.Now().UTC().Format("20060102150405.000000000"))
	o.Env = defaultText(o.Env, defaultEnv)
	o.Region = defaultText(o.Region, defaultScopeValue)
	o.Zone = defaultText(o.Zone, defaultScopeValue)
	o.Cluster = defaultText(o.Cluster, defaultScopeValue)
	o.RuleNamespace = defaultText(o.RuleNamespace, defaultRuleNamespace)
	o.RuleGroup = defaultText(o.RuleGroup, defaultRuleGroup)
	if o.WatchEnabled == nil {
		o.WatchEnabled = Bool(true)
	}
	if o.StellnulaGRPCPlaintext == nil {
		o.StellnulaGRPCPlaintext = Bool(true)
	}
	o.Labels = copyStringMap(o.Labels)
	return o, nil
}

func validateEndpoint(name string, endpoint string) error {
	if endpoint == "" {
		return nil
	}
	if _, err := url.ParseRequestURI(endpoint); err != nil {
		return fmt.Errorf("stellorbit: invalid %s: %w", name, err)
	}
	return nil
}

func defaultText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func copyStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	copied := make(map[string]string, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}
