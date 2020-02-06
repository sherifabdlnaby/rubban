package kibana

import (
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

func (c *Client) get(uri string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.baseUrl.String()+uri, nil)
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
	req, err := http.NewRequest("POST", c.baseUrl.String()+uri, nil)
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

func (c *Client) Validate(retry int, waitTime time.Duration) bool {
	var err error
	var resp *http.Response
	for i := 0; i < retry; i++ {
		resp, err = c.get("/api/status")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				c.logger.Info("successfully connected to Kibana")
				return true
			}
			err = fmt.Errorf("couldn't connect to kibana: %d %s", resp.StatusCode, resp.Status)
		}

		c.logger.Infow("could not connect to Kibana", "error", err.Error())
		c.logger.Infof("retrying in %g seconds...  (%d/%d)", waitTime.Seconds(), i+1, retry)
		time.Sleep(waitTime)
	}
	return false
}

func (c *Client) GuessVersion() (*semver.Version, error) {

	// 1
	resp, err := c.get("/api/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	info := Info{}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		err := json.NewDecoder(resp.Body).Decode(&info)
		if err != nil {
			return nil, err
		}
		return info.GetSemVar()
	}

	// 2
	// Will add more ways to guess version has above API was changed in other Kibana versions.

	return nil, nil
}
