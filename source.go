package stellorbit

import (
	stellnula "github.com/stellhub/stellnula-go-sdk"
	stellnulasource "github.com/stellhub/stellorbit-go-sdk/internal/source/stellnula"
)

func NewStellnulaGovernanceRuleSource(options Options, clientOptions ...stellnula.ClientOption) (GovernanceRuleSource, error) {
	normalized, err := options.normalize(false)
	if err != nil {
		return nil, err
	}
	return newStellnulaSource(normalized, clientOptions...)
}

func NewStellnulaGovernanceRuleSourceWithClient(client *stellnula.Client, failFast bool, logger Logger) GovernanceRuleSource {
	return stellnulasource.NewWithClient(client, failFast, logger)
}
