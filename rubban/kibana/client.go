package kibana

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/sherifabdlnaby/rubban/config"
	"github.com/sherifabdlnaby/rubban/log"
	"github.com/sherifabdlnaby/rubban/version"
)

//Client is a HTTP API Request wrapper.
type Client struct {
	baseURL  *url.URL
	username string
	password string
	http     *http.Client
	logger   log.Logger
}

//NewKibanaClient Constructor
func NewKibanaClient(config config.Kibana, logger log.Logger) (*Client, error) {
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
			/* #nosec */
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
			Timeout:   10 * time.Second,
		},
		logger: logger,
	}, nil
}

func (c *Client) getURLFromPath(path string) string {
	return c.baseURL.String() + path
}

func (c *Client) newRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.getURLFromPath(path), body)
	if err != nil {
		return nil, err
	}

	// Set Headers
	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Accept", "application/json")
	req.Header.Set("kbn-xsrf", "true")
	req.Header.Set("User-Agent", "Rubban/"+version.Version)

	// Set Auth
	req.SetBasicAuth(c.username, c.password)

	return req, nil
}

//Get Perform a GET Request to Kibana
func (c *Client) Get(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(ctx, "GET", path, body)
	if err != nil {
		return nil, err
	}
	return c.http.Do(req)
}

//Post Perform a POST Request to Kibana
func (c *Client) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(ctx, "POST", path, body)
	if err != nil {
		return nil, err
	}
	return c.http.Do(req)
}

//Put Perform a PUT Request to Kibana
func (c *Client) Put(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(ctx, "PUT", path, body)
	if err != nil {
		return nil, err
	}
	return c.http.Do(req)
}

//Validate Validate connection to Kibana by pinging /status api.
func (c *Client) Validate(ctx context.Context, retry int, waitTime time.Duration) error {
	var err error
	var resp *http.Response
	var pingPath = "/api/status"

	c.logger.Infof("Testing connection to Kibana API at %s", c.getURLFromPath(pingPath))

	for i := 0; i < retry+1; i++ {
		if i != 0 {
			c.logger.Infof("Retrying in %g seconds...  (%d/%d)", waitTime.Seconds(), i, retry)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):

			}
		}

		resp, err = c.Get(ctx, pingPath, nil)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				c.logger.Infof("Successfully connected to Kibana API %s", c.getURLFromPath(pingPath))
				return nil
			}
			err = fmt.Errorf("%s", resp.Status)
		}

		c.logger.Warnw(fmt.Sprintf("Could not connect to Kibana API %s", c.getURLFromPath(pingPath)), "error", err.Error())
	}
	return err
}

//GuessVersion Get Kibana Version (Will use different methods to determine API version)
func (c *Client) GuessVersion(ctx context.Context) (semver.Version, error) {

	// 1
	resp, err := c.Get(ctx, "/api/status", nil)
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
