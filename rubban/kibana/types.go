package kibana

import (
	"context"

	"github.com/Masterminds/semver/v3"
)

//API is an interface for supporting multiple Kibana APIs
type API interface {
	Info(ctx context.Context) (Info, error)

	Indices(ctx context.Context, filter string) ([]Index, error)

	IndexPatterns(ctx context.Context, filter string) ([]IndexPattern, error)

	BulkCreateIndexPattern(ctx context.Context, indexPattern map[string]IndexPattern) error
}

//Info for Json Unmarshalling API Response
type Info struct {
	Name    string `json:"name"`
	UUID    string `json:"uuid"`
	Version struct {
		Number        string `json:"number"`
		BuildHash     string `json:"build_hash"`
		BuildNumber   int    `json:"build_number"`
		BuildSnapshot bool   `json:"build_snapshot"`
	} `json:"version"`
}

//GetSemVer Get's Kibana Semantic Version
func (i Info) GetSemVer() (*semver.Version, error) {
	return semver.NewVersion(i.Version.Number)
}

//Index for Json Unmarshalling API Response
type Index struct {
	Name string `json:"index"`
}

//IndexPattern for Json Unmarshalling API Response
type IndexPattern struct {
	Title         string `json:"title"`
	TimeFieldName string `json:"timeFieldName"`
}

//BulkIndexPattern for Json Unmarshalling API Response
type BulkIndexPattern struct {
	Type       string       `json:"type"`
	Attributes IndexPattern `json:"attributes,omitempty"`
}

//IndexPatternPage for Json Unmarshalling API Response
type IndexPatternPage struct {
	Page         int `json:"page"`
	PerPage      int `json:"per_page"`
	Total        int `json:"total"`
	SavedObjects []struct {
		Type       string       `json:"type"`
		ID         string       `json:"id"`
		Attributes IndexPattern `json:"attributes"`
	} `json:"saved_objects"`
}
