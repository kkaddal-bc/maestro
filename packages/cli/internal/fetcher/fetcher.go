package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/kkaddal-bc/maestro/packages/cli/internal/manifest"
)

const (
	defaultBaseURL = "https://api.github.com"
	owner          = "kkaddal-bc"
	repo           = "maestro"
	userAgent      = "maestro-cli"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	UserAgent  string
	owner      string
	repo       string
}

func New() *Client {
	return &Client{
		BaseURL:    defaultBaseURL,
		HTTPClient: http.DefaultClient,
		UserAgent:  userAgent,
		owner:      owner,
		repo:       repo,
	}
}

func NewWithBaseURL(baseURL string) *Client {
	client := New()
	client.BaseURL = baseURL
	return client
}

func (c *Client) FetchManifest() (*manifest.Manifest, error) {
	release, err := c.fetchRelease("latest")
	if err != nil {
		return nil, err
	}

	asset, err := release.assetByName("manifest.json")
	if err != nil {
		return nil, err
	}

	reader, err := c.downloadAsset(asset.ID)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return manifest.Parse(reader)
}

func (c *Client) FetchSkillsArchive(version string) (io.ReadCloser, error) {
	release, err := c.fetchRelease("tags/" + url.PathEscape(version))
	if err != nil {
		return nil, err
	}

	asset, err := release.assetByName("skills.tar.gz")
	if err != nil {
		return nil, err
	}

	return c.downloadAsset(asset.ID)
}

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (r *release) assetByName(name string) (*asset, error) {
	for i := range r.Assets {
		if r.Assets[i].Name == name {
			return &r.Assets[i], nil
		}
	}
	return nil, fmt.Errorf("asset %q not found in release %s", name, r.TagName)
}

func (c *Client) fetchRelease(ref string) (*release, error) {
	var rel release
	if err := c.getJSON(path.Join("repos", c.owner, c.repo, "releases", ref), &rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func (c *Client) getJSON(endpoint string, dst any) error {
	req, err := c.newRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("request %s: %s: %s", endpoint, resp.Status, strings.TrimSpace(string(snippet)))
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode %s: %w", endpoint, err)
	}

	return nil
}

func (c *Client) downloadAsset(id int64) (io.ReadCloser, error) {
	req, err := c.newRequest(http.MethodGet, path.Join("repos", c.owner, c.repo, "releases", "assets", fmt.Sprint(id)), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("download asset %d: %s: %s", id, resp.Status, strings.TrimSpace(string(snippet)))
	}

	return resp.Body, nil
}

func (c *Client) newRequest(method, endpoint string, body io.Reader) (*http.Request, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	ref, err := url.Parse("/" + strings.TrimPrefix(endpoint, "/"))
	if err != nil {
		return nil, err
	}
	full := base.ResolveReference(ref)
	req, err := http.NewRequest(method, full.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent())
	return req, nil
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func (c *Client) userAgent() string {
	if c.UserAgent != "" {
		return c.UserAgent
	}
	return userAgent
}
