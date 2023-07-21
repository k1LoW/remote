package remote

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/google/go-github/v53/github"
	"github.com/jszwec/s3fs"
	"github.com/k1LoW/ghfs"
	"github.com/k1LoW/go-github-client/v53/factory"
	"github.com/mauri870/gcsfs"
	"google.golang.org/api/option"
)

var globalHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

var globalGitHubClient *github.Client

var globalS3Client s3iface.S3API

var globalGCSClient *storage.Client

type clientSet struct {
	httpClient   *http.Client
	githubClient *github.Client
	s3Client     s3iface.S3API
	gcsClient    *storage.Client
}

func newClientSet(opts []Option) *clientSet {
	cs := &clientSet{
		httpClient:   globalHTTPClient,
		githubClient: globalGitHubClient,
		s3Client:     globalS3Client,
		gcsClient:    globalGCSClient,
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

// S3Client set github.Client.
func S3Client(c s3iface.S3API) Option {
	return func(cs *clientSet) {
		cs.s3Client = c
	}
}

// GCSClient set storage.Client.
func GCSClient(c *storage.Client) Option {
	return func(cs *clientSet) {
		cs.gcsClient = c
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
	case "s3":
		return openS3(cs.s3Client, raw)
	case "gs", "gcs":
		return openGCS(cs.gcsClient, raw)
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

func openS3(c s3iface.S3API, raw string) (io.ReadCloser, error) {
	if c == nil && globalS3Client == nil {
		sess, err := session.NewSession()
		if err != nil {
			return nil, err
		}
		c = s3.New(sess)
		globalS3Client = c
	}
	splitted := strings.Split(strings.TrimPrefix(raw, "s3://"), "/")
	if len(splitted) < 2 {
		return nil, fmt.Errorf("invalid s3 path: %s", raw)
	}
	bucket := splitted[0]
	name := strings.Join(splitted[1:], "/")
	fsys := s3fs.New(c, bucket)
	return fsys.Open(name)
}

func openGCS(c *storage.Client, raw string) (io.ReadCloser, error) {
	if c == nil && globalGCSClient == nil {
		var err error
		ctx := context.Background()
		if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON") != "" {
			c, err = storage.NewClient(ctx, option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON"))))
		} else {
			c, err = storage.NewClient(ctx)
		}
		if err != nil {
			return nil, err
		}
		globalGCSClient = c
	}
	splitted := strings.Split(strings.TrimPrefix(strings.TrimPrefix(raw, "gs://"), "gcs://"), "/")
	if len(splitted) < 2 {
		return nil, fmt.Errorf("invalid gcs path: %s", raw)
	}
	bucket := splitted[0]
	name := strings.Join(splitted[1:], "/")
	fsys := gcsfs.NewWithClient(c, bucket)
	return fsys.Open(name)
}
