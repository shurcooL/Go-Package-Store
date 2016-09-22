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

// NewPresenter returns a Gitiles API-powered presenter.
// httpClient is the HTTP client to be used by the presenter for accessing the Gitiles API.
// If httpClient is nil, then http.DefaultClient is used.
func NewPresenter(httpClient *http.Client) gps.Presenter {
	return func(repo *gps.Repo) *gps.Presentation {
		switch {
		case strings.HasPrefix(repo.Remote.RepoURL, "https://code.googlesource.com/"):
			return presentGitilesRepo(httpClient, repo)
		default:
			return nil
		}
	}
}

func presentGitilesRepo(client *http.Client, repo *gps.Repo) *gps.Presentation {
	// This might take a while.
	log, err := fetchLog(client, repo.Remote.RepoURL+"/+log?format=JSON")
	if err != nil {
		return &gps.Presentation{Error: err}
	}

	home := template.URL("https://" + repo.Root)
	return &gps.Presentation{
		Home:    &home,
		Image:   "https://ssl.gstatic.com/codesite/ph/images/defaultlogo.png",
		Changes: extractChanges(repo, log),
	}
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

type log struct {
	Log  []commit `json:"log"`
	Next string   `json:"next"` // TODO: Use or remove.
}

type commit struct {
	Commit  string `json:"commit"`
	Message string `json:"message"`
}

func extractChanges(repo *gps.Repo, l log) []gps.Change {
	// Verify/find Repo.Remote.Revision.
	log := l.Log
	for len(log) > 0 && log[0].Commit != repo.Remote.Revision {
		log = log[1:]
	}

	var cs []gps.Change
	for _, commit := range log {
		if commit.Commit == repo.Local.Revision {
			break
		}
		cs = append(cs, gps.Change{
			Message: firstParagraph(commit.Message),
			URL:     template.URL(repo.Remote.RepoURL + "/+/" + commit.Commit + "%5e%21"),
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
