package kibana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/sherifabdlnaby/rubban/config"
	"github.com/sherifabdlnaby/rubban/log"
)

//APIVer7 Implements API Calls compatible with Kibana 7^
type APIVer7 struct {
	client *Client
	log    log.Logger
}

//NewAPIVer7 Constructor
func NewAPIVer7(config config.Kibana, log log.Logger) (*APIVer7, error) {
	client, err := NewKibanaClient(config, log.Extend("Client"))
	if err != nil {
		return &APIVer7{}, err
	}

	return &APIVer7{
		client: client,
		log:    log,
	}, nil
}

//Info Return Kibana Info
func (a *APIVer7) Info(ctx context.Context) (Info, error) {
	resp, err := a.client.Get(ctx, "/api/status", nil)
	if err != nil {
		return Info{}, err
	}
	defer resp.Body.Close()

	info := Info{}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		err := json.NewDecoder(resp.Body).Decode(&info)
		if err != nil {
			return Info{}, err
		}
	}

	return info, err
}

//Indices Get Indices match supported filter (support wildcards)
func (a *APIVer7) Indices(ctx context.Context, filter string) ([]Index, error) {
	indices := make([]Index, 0)
	resp, err := a.client.Post(ctx, fmt.Sprintf("/api/console/proxy?path=_cat/indices/%s?format=json&h=index&method=GET", filter), nil)
	if err != nil {
		return indices, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		err := json.NewDecoder(resp.Body).Decode(&indices)
		if err != nil {
			return nil, err
		}
	}
	return indices, err
}

//IndexPatterns Get IndexPatterns from kibana matching the supplied filter (support wildcards)
func (a *APIVer7) IndexPatterns(ctx context.Context, filter string) ([]IndexPattern, error) {

	page := 1
	count := 0
	aggPatterns := make([]IndexPattern, 0)
	for {
		patterns, total, err := a.indexPatternPage(ctx, filter, page)
		if err != nil {
			return nil, err
		}

		aggPatterns = append(aggPatterns, patterns...)
		count += len(patterns)
		page++

		if count >= total {
			break
		}
	}

	return aggPatterns, nil
}

//BulkCreateIndexPattern Add Index Patterns to Kibana
func (a *APIVer7) BulkCreateIndexPattern(ctx context.Context, indexPattern map[string]IndexPattern) error {
	if len(indexPattern) == 0 {
		return nil
	}

	// Prepare Requests
	bulkRequest := make([]BulkIndexPattern, 0)
	for _, pattern := range indexPattern {
		bulkRequest = append(bulkRequest, BulkIndexPattern{
			Type: "index-pattern",
			Attributes: IndexPattern{
				Title:         pattern.Title,
				TimeFieldName: pattern.TimeFieldName,
			},
		})
	}
	// Json Marshaling
	buff, err := json.Marshal(bulkRequest)
	if err != nil {
		return fmt.Errorf("failed to JSON marshaling bulk create index pattern")
	}

	// Send Request
	resp, err := a.client.Post(ctx, "/api/saved_objects/_bulk_create", bytes.NewReader(buff))
	if err != nil {
		return fmt.Errorf("failed to bulk create saved objects, error: %s", err.Error())
	}

	_ = resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to bulk create saved objects, error: %s", resp.Status)
	}

	return nil
}

func (a *APIVer7) indexPatternPage(ctx context.Context, filter string, page int) ([]IndexPattern, int, error) {

	indexPatterns := make([]IndexPattern, 0)
	indexPatternPage := IndexPatternPage{}
	resp, err := a.client.Get(ctx, fmt.Sprintf("/api/saved_objects/_find?fields=title&fields=timeFieldName&per_page=1&search=\"%s\"&search_fields=title&type=index-pattern&page=%d", filter, page), nil)
	if err != nil {
		return indexPatterns, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		err := json.NewDecoder(resp.Body).Decode(&indexPatternPage)
		if err != nil {
			return nil, 0, err
		}
	}

	for i := 0; i < len(indexPatternPage.SavedObjects); i++ {
		indexPatterns = append(indexPatterns, indexPatternPage.SavedObjects[i].Attributes)
	}

	return indexPatterns, indexPatternPage.Total, err
}
