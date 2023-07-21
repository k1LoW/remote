package remote

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/k1LoW/ghfs"
	"github.com/k1LoW/go-github-client/v53/factory"
)

var globalHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

var globalGitHubClient *github.Client

type clientSet struct {
	httpClient   *http.Client
	githubClient *github.Client
}

func newClientSet(opts []Option) *clientSet {
	cs := &clientSet{
		httpClient:   globalHTTPClient,
		githubClient: globalGitHubClient,
	}
	for _, opt := range opts {
		opt(cs)
	}
	return cs
}

type Option func(*clientSet)

// HTTPClient set http.Client.
func HTTPClient(c *http.Client) Option {
	return func(cs *clientSet) {
		cs.httpClient = c
	}
}

// GitHubClient set github.Client.
func GitHubClient(c *github.Client) Option {
	return func(cs *clientSet) {
		cs.githubClient = c
	}
}

// Open remote file.
func Open(raw string, opts ...Option) (io.ReadCloser, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	cs := newClientSet(opts)
	switch u.Scheme {
	case "http", "https":
		return openHTTP(cs.httpClient, u)
	case "github":
		return openGitHub(cs.githubClient, raw)
	default:
		p := strings.TrimPrefix(strings.TrimPrefix(raw, "file://"), "local://")
		return os.Open(p)
	}
}

// ReadAll remote file.
func ReadAll(raw string) ([]byte, error) {
	r, err := Open(raw)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func openHTTP(c *http.Client, u *url.URL) (io.ReadCloser, error) {
	resp, err := c.Get(u.String())
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func openGitHub(c *github.Client, raw string) (io.ReadCloser, error) {
	if c == nil && globalGitHubClient == nil {
		// initialize globalGitHubClient
		c, err := factory.NewGithubClient()
		if err != nil {
			return nil, fmt.Errorf("github client is not initialized: %w", err)
		}
		globalGitHubClient = c
	}

	splitted := strings.Split(strings.TrimPrefix(raw, "github://"), "/")
	if len(splitted) < 3 {
		return nil, fmt.Errorf("invalid github path: %s", raw)
	}
	owner := splitted[0]
	repo := splitted[1]
	name := strings.Join(splitted[2:], "/")
	fsys, err := ghfs.New(owner, repo, ghfs.Client(c))
	if err != nil {
		return nil, err
	}
	return fsys.Open(name)
}
