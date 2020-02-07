package kibana

import (
	"encoding/json"
	"fmt"
)

type ApiVer7 struct {
	client *Client
}

func NewApiVer7(client *Client) *ApiVer7 {
	return &ApiVer7{client: client}
}

func (a *ApiVer7) Info() (Info, error) {
	resp, err := a.client.get("/api/status")
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

func (a *ApiVer7) Indices(filter string) ([]Index, error) {
	indices := make([]Index, 0)
	resp, err := a.client.post(fmt.Sprintf("/api/console/proxy?path=_cat/indices/%s?format=json&h=index&method=GET", filter))
	if err != nil {
		return indices, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		err := json.NewDecoder(resp.Body).Decode(&indices)
		if err != nil {
			return indices, err
		}
	}
	return indices, err
}

func (a *ApiVer7) IndexPatternFields(filter string) ([]IndexPattern, error) {

	var err error
	page := 1
	count := 0
	total := 0
	aggPatterns := make([]IndexPattern, 0)
	patterns := make([]IndexPattern, 0)
	for {
		patterns, total, err = a.indexPatternPage(filter, page)
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

func (a *ApiVer7) IndexPatterns(filter string) ([]IndexPattern, error) {

	var err error
	page := 1
	count := 0
	total := 0
	aggPatterns := make([]IndexPattern, 0)
	patterns := make([]IndexPattern, 0)
	for {
		patterns, total, err = a.indexPatternPage(filter, page)
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

func (a *ApiVer7) BulkCreateIndexPattern(indexPattern []IndexPattern) error {
	if len(indexPattern) == 0 {
		return nil
	}

	// Prepare Requests
	bulkRequest := make([]BulkIndexPattern, 0)
	for _, pattern := range indexPattern {
		bulkRequest = append(bulkRequest, BulkIndexPattern{
			Type: "index-pattern",
			Attributes: BulkIndexPatterAttributes{
				Title:         pattern.Title,
				TimeFieldName: pattern.TimeFieldName,
			},
		})
	}
	// Json Marshalling
	buff, err := json.Marshal(bulkRequest)
	if err != nil {
		return fmt.Errorf("failed to JSON marshalling bulk create index pattern")
	}

	// Send Request
	resp, err := a.client.postWithJson("/api/saved_objects/_bulk_create", buff)
	if err != nil {
		return fmt.Errorf("failed to bulk create saved objects, error: %s", err.Error())
	}

	_ = resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to bulk create saved objects, error: %s", resp.Status)
	}

	return nil
}

func (a *ApiVer7) indexPatternPage(filter string, page int) ([]IndexPattern, int, error) {

	indexPatterns := make([]IndexPattern, 0)
	indexPatternPage := IndexPatternPage{}
	resp, err := a.client.get(fmt.Sprintf("/api/saved_objects/_find?fields=title&fields=timeFieldName&per_page=1&search=\"%s\"&search_fields=title&type=index-pattern&page=%d", filter, page))
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
