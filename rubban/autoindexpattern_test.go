package rubban

import (
	"testing"

	"github.com/sherifabdlnaby/rubban/config"
	"github.com/sherifabdlnaby/rubban/rubban/kibana"
)

type mockAPI struct {
	indices       []kibana.Index
	indexpatterns []kibana.IndexPattern
}

func (m *mockAPI) Info() (kibana.Info, error) {
	return kibana.Info{}, nil
}

func (m *mockAPI) BulkCreateIndexPattern(indexPattern []kibana.IndexPattern) error {
	return nil
}

func (m *mockAPI) Indices(filter string) ([]kibana.Index, error) {
	return m.indices, nil
}

func (m *mockAPI) IndexPatterns(filter string) ([]kibana.IndexPattern, error) {
	return m.indexpatterns, nil
}

func newMockAPI(indices []kibana.Index, indexpatterns []kibana.IndexPattern) kibana.API {
	return &mockAPI{indices: indices, indexpatterns: indexpatterns}
}

// TestAutoindexPatternMatchers tests how the matchers work.
func TestAutoindexPatternMatchers(t *testing.T) {
	for _, tcase := range []struct {
		generalPattern        string
		indices               []kibana.Index
		indexpatterns         []kibana.IndexPattern
		expectedIndexPatterns []string
		tcaseName             string
	}{
		{
			generalPattern:        "?-*",
			indices:               []kibana.Index{{Name: "foo-bar-2020.02.14"}, {Name: "foo-qux-2020.02.14"}, {Name: "test-2020.02.14"}},
			indexpatterns:         []kibana.IndexPattern{{Title: "test-*", TimeFieldName: "@timestamp"}},
			expectedIndexPatterns: []string{"foo-*"},
			tcaseName:             `"test-*" already exists`,
		},
		{
			generalPattern:        "?-*",
			indices:               []kibana.Index{{Name: "foo-bar-2020.02.14"}, {Name: "foo-qux-2020.02.14"}, {Name: "test-2020.02.14"}},
			indexpatterns:         []kibana.IndexPattern{},
			expectedIndexPatterns: []string{"foo-*", "test-*"},
			tcaseName:             `"test-*" does not exist`,
		},
		{
			generalPattern:        "?-*",
			indices:               []kibana.Index{},
			indexpatterns:         []kibana.IndexPattern{{Title: "*", TimeFieldName: "@timestamp"}},
			expectedIndexPatterns: []string{},
			tcaseName:             `negative test`,
		},
		{
			generalPattern:        "?-*",
			indices:               []kibana.Index{{Name: "-cool-index-"}, {Name: ".kibana"}, {Name: "test----aabcc2020.02.14"}},
			indexpatterns:         []kibana.IndexPattern{},
			expectedIndexPatterns: []string{"test-*", "-*", "*-*"},
			tcaseName:             `random gibberish that should not match most of the time besides two`,
		},
		{
			generalPattern:        "?-?-*",
			indices:               []kibana.Index{{Name: "foo-bar-2020.02.14"}, {Name: "foo-baz-2020.02.14"}},
			indexpatterns:         []kibana.IndexPattern{},
			expectedIndexPatterns: []string{"foo-bar-*", "foo-baz-*"},
			tcaseName:             `multiple matcher test`,
		},
		{
			generalPattern:        "?-?-*",
			indices:               []kibana.Index{{Name: "foo-bar-2020.02.14"}, {Name: "foo-baz-2020.02.14"}},
			indexpatterns:         []kibana.IndexPattern{{Title: "foo-bar-*", TimeFieldName: "@timestamp"}},
			expectedIndexPatterns: []string{"foo-baz-*"},
			tcaseName:             `multiple matcher test but w/ existing index patterns`,
		},
		{
			generalPattern:        "?-?-*",
			indices:               []kibana.Index{{Name: "foo-bar-a-2020.02.14"}, {Name: "foo-baz-b-2020.02.14"}},
			indexpatterns:         []kibana.IndexPattern{},
			expectedIndexPatterns: []string{"foo-baz-*", "foo-bar-*"},
			tcaseName:             `multiple matcher and matching eagerly vs. lazily test`,
		},
	} {
		aip := NewAutoIndexPattern(config.AutoIndexPattern{
			Enabled: true,
			GeneralPatterns: []config.GeneralPattern{{
				Pattern:       tcase.generalPattern,
				TimeFieldName: "@timestamp",
			}},
			Schedule: "* * * * *",
		})

		m := newMockAPI(tcase.indices, tcase.indexpatterns)
		r := rubban{api: m}

		computedIndexPatterns := make(indexPatternMap)
		r.getIndexPattern(aip.GeneralPatterns[0], computedIndexPatterns)

		t.Run(tcase.tcaseName, func(t *testing.T) {
			if len(tcase.expectedIndexPatterns) == 0 && len(computedIndexPatterns) != 0 {
				t.Fatalf("expected zero index patterns but got %d (%v)", len(computedIndexPatterns), computedIndexPatterns)
			} else {
				if len(tcase.expectedIndexPatterns) != len(computedIndexPatterns) {
					t.Fatalf("expected %d index patterns but got %d (%v)", len(tcase.expectedIndexPatterns), len(computedIndexPatterns), computedIndexPatterns)
				}
				for _, e := range tcase.expectedIndexPatterns {
					_, ok := computedIndexPatterns[e]
					if !ok {
						t.Fatalf("failed to find index pattern %s (%v)", e, computedIndexPatterns)
					}
				}
			}
		})

	}
}
