package kibana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

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
func (a *APIVer7) IndexPatterns(ctx context.Context, filter string, fields []string) ([]IndexPattern, error) {

	page := 1
	count := 0
	aggPatterns := make([]IndexPattern, 0)
	for {
		patterns, total, err := a.indexPatternPage(ctx, filter, fields, page)
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

//IndexPatterns Get IndexPatterns from kibana matching the supplied filter (support wildcards)
func (a *APIVer7) IndexPatternFields(ctx context.Context, pattern string) (*IndexPatternFields, error) {

	urlQuery := make(url.Values, 0)
	urlQuery.Add("meta_fields", "_score")
	urlQuery.Add("meta_fields", "_index")
	urlQuery.Add("meta_fields", "_type")
	urlQuery.Add("meta_fields", "_id")
	urlQuery.Add("meta_fields", "_source")
	urlQuery.Add("pattern", pattern)

	// Manually Encode Filter because for some reason Kibana needs a unescaped and quoted filter string.
	resp, err := a.client.Get(ctx, "/api/index_patterns/_fields_for_wildcard?"+urlQuery.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := IndexPatternFields{}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		err := json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return nil, err
		}
	}

	return &response, nil

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

func (a *APIVer7) indexPatternPage(ctx context.Context, filter string, fields []string, page int) ([]IndexPattern, int, error) {

	indexPatterns := make([]IndexPattern, 0)
	indexPatternPage := IndexPatternPage{}
	urlQuery := make(url.Values, 0)
	urlQuery.Add("per_page", "20")
	urlQuery.Add("page", strconv.Itoa(page))
	urlQuery.Add("search_fields", "title")
	urlQuery.Add("type", "index-pattern")
	urlQuery.Add("fields", "title")
	urlQuery.Add("fields", "timeFieldName")
	for _, field := range fields {
		urlQuery.Add("fields", field)
	}
	query := urlQuery.Encode()

	// Manually Encode Filter because for some reason Kibana needs a unescaped and quoted filter string.
	resp, err := a.client.Get(ctx, "/api/saved_objects/_find?search=\""+filter+"\""+"&"+query, nil)
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
		indexPattern := indexPatternPage.SavedObjects[i].Attributes
		indexPattern.ID = indexPatternPage.SavedObjects[i].ID
		indexPattern.Version = indexPatternPage.SavedObjects[i].Version
		indexPatterns = append(indexPatterns, indexPattern)
	}

	return indexPatterns, indexPatternPage.Total, err
}

type PutIndexPatternAttr struct {
	Title  string `json:"title"`
	Fields string `json:"fields,squash"`
}

type PutIndexPatternBody struct {
	Attributes PutIndexPatternAttr `json:"attributes"`
	Version    string              `json:"version"`
}

func (a *APIVer7) PutIndexPattern(ctx context.Context, indexPattern IndexPattern) error {
	// Prepare Requests

	request := PutIndexPatternBody{
		Attributes: PutIndexPatternAttr{
			Title:  indexPattern.Title,
			Fields: string(indexPattern.IndexPatternFields.Fields),
		},
		Version: indexPattern.Version,
	}

	// Json Marshaling
	buff, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to JSON marshaling PUT index pattern")
	}

	// Send Request
	resp, err := a.client.Put(ctx, "/api/saved_objects/index-pattern/"+indexPattern.ID, bytes.NewReader(buff))
	if err != nil {
		return fmt.Errorf("failed to PUT index pattern, error: %s", err.Error())
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to PUT index pattern, error: %s", resp.Status)
	}

	return err
}
