package kibana

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/sherifabdlnaby/bosun/config"
	"github.com/sherifabdlnaby/bosun/log"
)

type Client struct {
	baseUrl  *url.URL
	username string
	password string
	http     *http.Client
	logger   log.Logger
}

type API interface {
	Info() (Info, error)

	Indices(filter string) ([]Index, error)

	IndexPatternFields(filter string) ([]IndexPattern, error)

	IndexPatterns(filter string) ([]IndexPattern, error)

	BulkCreateIndexPattern(indexPattern []IndexPattern) error
}

func NewKibanaClient(config config.Kibana, logger log.Logger) (*Client, error) {

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
		baseUrl:  baseURL,
		username: config.User,
		password: config.Password,
		http: &http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
			Timeout:   5 * time.Second,
		},
		logger: logger,
	}, nil
}

func (c *Client) getFullUrl(path string) string {
	return c.baseUrl.String() + path
}
func (c *Client) get(uri string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.getFullUrl(uri), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("kbn-xsrf", "true")
	req.SetBasicAuth(c.username, c.password)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) post(uri string) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.getFullUrl(uri), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("kbn-xsrf", "true")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) postWithJson(uri string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.getFullUrl(uri), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("kbn-xsrf", "true")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) Validate(retry int, waitTime time.Duration) bool {
	var err error
	var resp *http.Response
	var pingApiUrl = "/api/status"

	c.logger.Infof("Testing connection to Kibana API at %s", c.getFullUrl(pingApiUrl))

	for i := 0; i < retry+1; i++ {

		if i != 0 {
			c.logger.Infof("Retrying in %g seconds...  (%d/%d)", waitTime.Seconds(), i, retry)
			time.Sleep(waitTime)
		}

		resp, err = c.get(pingApiUrl)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				c.logger.Infof("Successfully connected to Kibana API %s", c.getFullUrl(pingApiUrl))
				return true
			}
			err = fmt.Errorf("%s", resp.Status)
		}

		c.logger.Warnw(fmt.Sprintf("Could not connect to Kibana API %s", c.getFullUrl(pingApiUrl)), "error", err.Error())
	}
	return false
}

func (c *Client) GuessVersion() (semver.Version, error) {

	// 1
	resp, err := c.get("/api/status")
	if err != nil {
		return semver.Version{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		info := Info{}
		err := json.NewDecoder(resp.Body).Decode(&info)
		if err != nil {
			return semver.Version{}, err
		}
		semVer, err := info.GetSemVer()
		if err != nil {
			return semver.Version{}, err
		}
		return *semVer, nil
	}

	// 2
	// Will add more ways to guess version has above API was changed in other Kibana versions.

	return semver.Version{}, nil
}
