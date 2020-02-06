package kibana

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sherifabdlnaby/bosun/config"
)

type Client struct {
	BaseURL  *url.URL
	Username string
	Password string
	http     *http.Client
}

func NewKibanaClient(config config.Kibana) (*Client, error) {

	// Create Base URL

	//// Add Scheme if doesn't exist (default to HTTP)
	rawUrl := config.Host
	if !strings.HasPrefix(rawUrl, "http://") && !strings.HasPrefix(rawUrl, "https://") {
		rawUrl = "http://" + rawUrl
	}

	//// Create BaseUrl
	baseURL, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	return &Client{
		BaseURL:  baseURL,
		Username: config.User,
		Password: config.Password,
		http: &http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
			Timeout:   5 * time.Second,
		},
	}, nil
}

func (k *Client) get(uri string) (*http.Response, error) {
	req, err := http.NewRequest("GET", k.BaseURL.String()+uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("kbn-xsrf", "true")
	req.SetBasicAuth(k.Username, k.Password)
	resp, err := k.http.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (k *Client) post(uri string) (*http.Response, error) {
	req, err := http.NewRequest("POST", k.BaseURL.String()+uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("kbn-xsrf", "true")
	req.SetBasicAuth(k.Username, k.Password)
	resp, err := k.http.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (k *Client) Info() (Info, error) {
	resp, err := k.get("/api/status")
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

func (k *Client) Indices(filter string) ([]Index, error) {
	resp, err := k.post(fmt.Sprintf("/api/console/proxy?path=_cat/indices/%s?format=json&h=index&method=GET", filter))
	defer resp.Body.Close()

	indices := make([]Index, 0)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		err := json.NewDecoder(resp.Body).Decode(&indices)
		if err != nil {
			return indices, err
		}
	}
	return indices, err
}

func (k *Client) IndexPatterns(filter string) ([]Index, error) {
	resp, err := k.post(fmt.Sprintf("/api/console/proxy?path=_cat/indices/%s?format=json&h=index&method=GET", filter))
	defer resp.Body.Close()

	indices := make([]Index, 0)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		err := json.NewDecoder(resp.Body).Decode(&indices)
		if err != nil {
			return indices, err
		}
	}
	return indices, err
}
