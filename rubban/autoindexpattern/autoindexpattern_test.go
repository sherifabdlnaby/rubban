package autoindexpattern

import (
	"context"
	"testing"

	"github.com/sherifabdlnaby/rubban/config"
	"github.com/sherifabdlnaby/rubban/log"
	"github.com/sherifabdlnaby/rubban/rubban/kibana"
)

type mockAPI struct {
	indices       []kibana.Index
	indexPatterns []kibana.IndexPattern
}

func (m *mockAPI) Info(ctx context.Context) (kibana.Info, error) {
	panic("implement me")
}

func (m *mockAPI) Indices(ctx context.Context, filter string) ([]kibana.Index, error) {
	return m.indices, nil
}

func (m *mockAPI) IndexPatterns(ctx context.Context, filter string) ([]kibana.IndexPattern, error) {
	return m.indexPatterns, nil
}

func (m *mockAPI) BulkCreateIndexPattern(ctx context.Context, indexPattern map[string]kibana.IndexPattern) error {
	panic("implement me")
}

func newMockAPI(indices []kibana.Index, indexPatterns []kibana.IndexPattern) kibana.API {
	return &mockAPI{indices: indices, indexPatterns: indexPatterns}
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
		autoIdxPttrn := NewAutoIndexPattern(config.AutoIndexPattern{
			Enabled: true,
			GeneralPatterns: []config.GeneralPattern{{
				Pattern:       tcase.generalPattern,
				TimeFieldName: "@timestamp",
			}},
			Schedule: "* * * * *",
		}, newMockAPI(tcase.indices, tcase.indexpatterns), log.Default())

		///
		result := autoIdxPttrn.getIndexPattern(context.Background(), autoIdxPttrn.GeneralPatterns[0])

		t.Run(tcase.tcaseName, func(t *testing.T) {
			if len(tcase.expectedIndexPatterns) == 0 && len(result) != 0 {
				t.Fatalf("expected zero index patterns but got %d (%v)", len(result), result)
			} else {
				if len(tcase.expectedIndexPatterns) != len(result) {
					t.Fatalf("expected %d index patterns but got %d (%v)", len(tcase.expectedIndexPatterns), len(result), result)
				}
				for _, e := range tcase.expectedIndexPatterns {
					_, ok := result[e]
					if !ok {
						t.Fatalf("failed to find index pattern %s (%v)", e, result)
					}
				}
			}
		})

	}
}
