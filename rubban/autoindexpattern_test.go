package rubban

import (
	"github.com/sherifabdlnaby/rubban/config"
	"github.com/sherifabdlnaby/rubban/rubban/kibana"
	"testing"
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
		GeneralPatterns: []config.GeneralPattern{config.GeneralPattern{
			Pattern:       "*-?",
			TimeFieldName: "@timestamp",
		}},
		Schedule: "* * * * *",
	})

	for _, tcase := range []struct {
		indices       []kibana.Index
		indexpatterns []kibana.IndexPattern
	}{
		{
			indices:       []kibana.Index{kibana.Index{Name: "foo-bar-aa-2020.02.14"}, kibana.Index{Name: "foo-qux-aa-2020.02.14"}},
			indexpatterns: []kibana.IndexPattern{kibana.IndexPattern{Title: "test-*", TimeFieldName: "@timestamp"}},
		},
		{
			indices:       []kibana.Index{kibana.Index{Name: "foo-bar-2020.02.14"}, kibana.Index{Name: "foo-qux-2020.02.14"}},
			indexpatterns: []kibana.IndexPattern{kibana.IndexPattern{Title: "test-*", TimeFieldName: "@timestamp"}},
		},
	} {
		m := newMockAPI(tcase.indices, tcase.indexpatterns)
		r := rubban{api: m}
		computedIndexPatterns := make(indexPatternMap)
		r.getIndexPattern(aip.GeneralPatterns[0], computedIndexPatterns)

		_, ok := computedIndexPatterns["foo-*"]
		if !ok {
			t.Fatal("failed to find index pattern foo-*")
		}
	}
}
