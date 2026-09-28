package profile

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
)

//go:embed cohorts/*.json
var cohortFS embed.FS

// Store is the set of cohorts compiled into the daemon.
type Store struct {
	byID   map[string]Profile
	active string
}

// LoadEmbedded reads and validates every shipped cohort.
func LoadEmbedded() (*Store, error) {
	entries, err := fs.ReadDir(cohortFS, "cohorts")
	if err != nil {
		return nil, err
	}
	byID := make(map[string]Profile, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := cohortFS.ReadFile("cohorts/" + entry.Name())
		if err != nil {
			return nil, err
		}
		p, err := Parse(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", entry.Name(), err)
		}
		if _, exists := byID[p.ID]; exists {
			return nil, fmt.Errorf("duplicate profile id %s", p.ID)
		}
		byID[p.ID] = p
	}
	if _, ok := byID[DefaultID]; !ok {
		return nil, fmt.Errorf("default profile %s is not shipped", DefaultID)
	}
	return &Store{byID: byID, active: DefaultID}, nil
}

// Active returns the cohort the daemon serves when no id is requested.
func (s *Store) Active() Profile {
	return s.byID[s.active]
}

// Get returns a cohort by id. An empty id returns the active cohort.
func (s *Store) Get(id string) (Profile, error) {
	if id == "" {
		id = s.active
	}
	p, ok := s.byID[id]
	if !ok {
		return Profile{}, fmt.Errorf("unknown profile")
	}
	return p, nil
}

// IDs returns cohort ids in lexical order.
func (s *Store) IDs() []string {
	ids := make([]string, 0, len(s.byID))
	for id := range s.byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
