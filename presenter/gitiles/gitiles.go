// Package gitiles provides a Gitiles API-powered presenter. It supports repositories that are on code.googlesource.com.
package gitiles

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/shurcooL/Go-Package-Store/presenter"
)

// NewPresenter returns a Gitiles API-powered presenter.
// httpClient is the HTTP client to be used by the presenter for accessing the Gitiles API.
// If httpClient is nil, then http.DefaultClient is used.
func NewPresenter(httpClient *http.Client) presenter.Presenter {
	return func(ctx context.Context, repo presenter.Repo) *presenter.Presentation {
		switch {
		case strings.HasPrefix(repo.RepoURL, "https://code.googlesource.com/"):
			return presentGitilesRepo(ctx, httpClient, repo)
		default:
			return nil
		}
	}
}

func presentGitilesRepo(ctx context.Context, client *http.Client, repo presenter.Repo) *presenter.Presentation {
	// This might take a while.
	log, err := fetchLog(ctx, client, repo.RepoURL+"/+log?format=JSON")
	if err != nil {
		return &presenter.Presentation{Error: err}
	}

	return &presenter.Presentation{
		HomeURL:  "https://" + repo.Root,
		ImageURL: "https://ssl.gstatic.com/codesite/ph/images/defaultlogo.png",
		Changes:  extractChanges(repo, log),
	}
}

// fetchLog fetches a Gitiles log at a given url, using client.
func fetchLog(ctx context.Context, client *http.Client, url string) (log, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return log{}, err
	}
	req.Header.Set("User-Agent", "github.com/shurcooL/Go-Package-Store/presenter/gitiles")
	resp, err := client.Do(req.WithContext(ctx))
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
// Source: https://www.chromium.org/developers/change-logs.
const header = `)]}'` + "\n"

type log struct {
	Log  []commit `json:"log"`
	Next string   `json:"next"`
}

type commit struct {
	Commit  string `json:"commit"`
	Message string `json:"message"`
}

func extractChanges(repo presenter.Repo, l log) []presenter.Change {
	// Verify/find Repo.RemoteRevision.
	log := l.Log
	for len(log) > 0 && log[0].Commit != repo.RemoteRevision {
		log = log[1:]
	}

	var cs []presenter.Change
	for _, commit := range log {
		if commit.Commit == repo.LocalRevision {
			break
		}
		cs = append(cs, presenter.Change{
			Message: firstParagraph(commit.Message),
			URL:     repo.RepoURL + "/+/" + commit.Commit + "%5e%21",
		})
	}
	return cs
}

// firstParagraph returns the first paragraph of text s.
func firstParagraph(s string) string {
	i := strings.Index(s, "\n\n")
	if i == -1 {
		return s
	}
	return s[:i]
}
