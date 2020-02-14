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
	aip := NewAutoIndexPattern(config.AutoIndexPattern{
		Enabled: true,
		GeneralPatterns: []config.GeneralPattern{{
			Pattern:       "?-*",
			TimeFieldName: "@timestamp",
		}},
		Schedule: "* * * * *",
	})

	for i, tcase := range []struct {
		indices               []kibana.Index
		indexpatterns         []kibana.IndexPattern
		expectedIndexPatterns []string
	}{
		// "test-*" already exists.
		{
			indices:               []kibana.Index{{Name: "foo-bar-2020.02.14"}, {Name: "foo-qux-2020.02.14"}, {Name: "test-2020.02.14"}},
			indexpatterns:         []kibana.IndexPattern{{Title: "test-*", TimeFieldName: "@timestamp"}},
			expectedIndexPatterns: []string{"foo-*"},
		},
		// "test-*" does not exist.
		{
			indices:               []kibana.Index{{Name: "foo-bar-2020.02.14"}, {Name: "foo-qux-2020.02.14"}, {Name: "test-2020.02.14"}},
			indexpatterns:         []kibana.IndexPattern{},
			expectedIndexPatterns: []string{"foo-*", "test-*"},
		},
		// negative test.
		{
			indices:               []kibana.Index{},
			indexpatterns:         []kibana.IndexPattern{{Title: "*", TimeFieldName: "@timestamp"}},
			expectedIndexPatterns: []string{},
		},
		// random gibberish that should not match most of the time besides the last one.
		{
			indices:               []kibana.Index{{Name: "fooaa2020.02.14"}, {Name: "-cool-index-"}, {Name: ".kibana"}, {Name: "test----aabcc2020.02.14"}},
			indexpatterns:         []kibana.IndexPattern{},
			expectedIndexPatterns: []string{"test-*"},
		},
	} {
		m := newMockAPI(tcase.indices, tcase.indexpatterns)
		r := rubban{api: m}

		computedIndexPatterns := make(indexPatternMap)
		r.getIndexPattern(aip.GeneralPatterns[0], computedIndexPatterns)

		if len(tcase.expectedIndexPatterns) == 0 && len(computedIndexPatterns) != 0 {
			t.Fatalf("(%d) expected zero index patterns but got %d (%v)", i, len(computedIndexPatterns), computedIndexPatterns)
		} else {
			for _, e := range tcase.expectedIndexPatterns {
				_, ok := computedIndexPatterns[e]
				if !ok {
					t.Fatalf("(%d) failed to find index pattern %s", i, e)
				}
			}
		}
	}
}
