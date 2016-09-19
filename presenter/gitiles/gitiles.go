// Package gitiles provides a Gitiles API-powered presenter. It supports repositories that are on code.googlesource.com.
package gitiles

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/shurcooL/Go-Package-Store"
)

// SetClient sets a custom HTTP client for accessing the Gitiles API by this presenter.
// By default, http.DefaultClient is used.
//
// It should not be called while the presenter is in use.
func SetClient(httpClient *http.Client) {
	client = httpClient
}

// client is the HTTP client used by this presenter.
var client = http.DefaultClient

func init() {
	gps.RegisterProvider(func(repo *gps.Repo) gps.Presenter {
		switch {
		case strings.HasPrefix(repo.Remote.RepoURL, "https://code.googlesource.com/"):
			return newGitilesPresenter(repo)
		default:
			return nil
		}
	})
}

type gitilesPresenter struct {
	repo *gps.Repo
	log  log
	err  error
}

func newGitilesPresenter(repo *gps.Repo) gps.Presenter {
	p := &gitilesPresenter{repo: repo}

	// This might take a while.
	p.log, p.err = fetchLog(client, repo.Remote.RepoURL+"/+log?format=JSON")

	return p
}

// fetchLog fetches a Gitiles log at a given url, using client.
func fetchLog(client *http.Client, url string) (log, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return log{}, err
	}
	req.Header.Set("User-Agent", "github.com/shurcooL/Go-Package-Store/presenter/gitiles")
	resp, err := client.Do(req)
	if err != nil {
		return log{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return log{}, fmt.Errorf("non-200 status code: %v", resp.StatusCode)
	}

	// Consume and verify header.
	buf := make([]byte, len(header))
	if _, err := io.ReadFull(resp.Body, buf); err != nil {
		return log{}, err
	}
	if !bytes.Equal(buf, []byte(header)) {
		return log{}, fmt.Errorf("header %q doesn't match expected %q", string(buf), header)
	}

	var l log
	err = json.NewDecoder(resp.Body).Decode(&l)
	return l, err
}

// Note, that JSON format has a ")]}'" line at the top, to prevent cross-site scripting.
// When parsing, assert that the first line has ")]}'", strip it, and parse the rest of
// JSON normally.
//
// Source: http://www.chromium.org/developers/change-logs
const header = `)]}'` + "\n"

func (g gitilesPresenter) Home() *template.URL {
	url := template.URL("https://" + g.repo.Root)
	return &url
}

func (gitilesPresenter) Image() template.URL {
	return "https://ssl.gstatic.com/codesite/ph/images/defaultlogo.png"
}

func (g gitilesPresenter) Changes() <-chan gps.Change {
	// Verify/find Repo.Remote.Revision.
	log := g.log.Log
	for len(log) > 0 && log[0].Commit != g.repo.Remote.Revision {
		log = log[1:]
	}

	out := make(chan gps.Change)
	go func() {
		for _, commit := range log {
			if commit.Commit == g.repo.Local.Revision {
				break
			}
			out <- gps.Change{
				Message: gps.FirstParagraph(commit.Message),
				URL:     template.URL(g.repo.Remote.RepoURL + "/+/" + commit.Commit + "%5e%21"),
			}
		}
		close(out)
	}()
	return out
}

func (g gitilesPresenter) Error() error { return g.err }

type log struct {
	Log  []commit `json:"log"`
	Next string   `json:"next"` // TODO: Use or remove.
}

type commit struct {
	Commit  string `json:"commit"`
	Message string `json:"message"`
}
