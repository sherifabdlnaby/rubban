package kibana

import (
	"bytes"
	"context"
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
	baseURL  *url.URL
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
	rawURL := config.Host
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	//// Create BaseUrl
	baseURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		baseURL:  baseURL,
		username: config.User,
		password: config.Password,
		http: &http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
			Timeout:   10 * time.Second,
		},
		logger: logger,
	}, nil
}

func (c *Client) getFullURL(path string) string {
	return c.baseURL.String() + path
}
func (c *Client) get(ctx context.Context, uri string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.getFullURL(uri), nil)
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
	req, err := http.NewRequest("POST", c.getFullURL(uri), nil)
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

func (c *Client) postWithJSON(uri string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.getFullURL(uri), bytes.NewBuffer(body))
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

func (c *Client) Validate(ctx context.Context, retry int, waitTime time.Duration) error {
	var err error
	var resp *http.Response
	var pingAPIURL = "/api/status"

	c.logger.Infof("Testing connection to Kibana API at %s", c.getFullURL(pingAPIURL))

	for i := 0; i < retry+1; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if i != 0 {
				c.logger.Infof("Retrying in %g seconds...  (%d/%d)", waitTime.Seconds(), i, retry)
				time.Sleep(waitTime)
			}

			resp, err = c.get(ctx, pingAPIURL)
			if err == nil {
				_ = resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					c.logger.Infof("Successfully connected to Kibana API %s", c.getFullURL(pingAPIURL))
					return nil
				}
				err = fmt.Errorf("%s", resp.Status)
			}

			c.logger.Warnw(fmt.Sprintf("Could not connect to Kibana API %s", c.getFullURL(pingAPIURL)), "error", err.Error())
		}
	}
	return err
}

func (c *Client) GuessVersion() (semver.Version, error) {

	// 1
	resp, err := c.get(context.TODO(), "/api/status")
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
