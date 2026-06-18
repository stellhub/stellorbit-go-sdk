package governance

type Logger interface {
	Printf(format string, args ...any)
}

type SnapshotParser struct {
	parser Parser
	logger Logger
}

func NewSnapshotParser(parser Parser, logger Logger) SnapshotParser {
	return SnapshotParser{parser: parser, logger: logger}
}

func (p SnapshotParser) Parse(snapshot Snapshot, previous Registry) Registry {
	if len(snapshot.Entries) == 0 {
		return NewRegistry(snapshot.Revision, snapshot.Checksum, nil)
	}
	previousRules := previous.Rules
	parsed := make([]Rule, 0, len(snapshot.Entries))
	hasDeletedEntry := false
	hasInvalidEntry := false
	hasNonDeletedEntry := false
	hasSuccessfulEntry := false
	for _, entry := range snapshot.Entries {
		identity := entryIdentity(entry)
		if entry.Deleted {
			hasDeletedEntry = true
			previousRules, _ = removePreviousRules(previousRules, identity)
			continue
		}
		hasNonDeletedEntry = true
		rules, err := p.parser.ParseAll(entry, snapshot.Checksum)
		if err != nil {
			hasInvalidEntry = true
			var fallback []Rule
			previousRules, fallback = removePreviousRules(previousRules, identity)
			if len(fallback) > 0 {
				parsed = append(parsed, fallback...)
			} else {
				p.logf("stellorbit: skip invalid governance config %s: %v", identity, err)
			}
			continue
		}
		hasSuccessfulEntry = true
		previousRules, _ = removePreviousRules(previousRules, identity)
		parsed = append(parsed, rules...)
	}
	if !hasSuccessfulEntry && hasInvalidEntry && hasNonDeletedEntry && !hasDeletedEntry && len(previous.Rules) > 0 {
		p.logf("stellorbit: all governance rules failed to parse, keep last-known-good registry")
		return previous.Clone()
	}
	return NewRegistry(snapshot.Revision, snapshot.Checksum, parsed)
}

func removePreviousRules(previousRules []Rule, configID string) ([]Rule, []Rule) {
	remaining := make([]Rule, 0, len(previousRules))
	removed := make([]Rule, 0)
	for _, rule := range previousRules {
		if rule.FromConfig(configID) {
			removed = append(removed, rule)
			continue
		}
		remaining = append(remaining, rule)
	}
	return remaining, removed
}

func entryIdentity(entry Entry) string {
	if entry.ConfigID != "" {
		return entry.ConfigID
	}
	return entry.ConfigKey
}

func (p SnapshotParser) logf(format string, args ...any) {
	if p.logger != nil {
		p.logger.Printf(format, args...)
	}
}
