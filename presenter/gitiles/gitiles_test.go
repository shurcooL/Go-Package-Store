package gitiles

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestFetchLog(t *testing.T) {
	logFile, err := os.Open(filepath.Join("testdata", "+log.json"))
	if err != nil {
		panic(err)
	}
	// logFile will be closed by defer resp.Body.Close().
	client := &http.Client{
		Transport: mockTripper(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       logFile,
			}, nil
		}),
	}

	log, err := fetchLog(client, "")
	if err != nil {
		t.Fatalf("fetchLog: %v", err)
	}

	if got, want := len(log.Log), 100; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got, want := log.Log[0], (commit{
		Commit:  "7a1b48a03240285fcdb91f7890f647cd90358f84",
		Message: "google-api-go-client: update all APIs\n\nChange-Id: If682f3b0bcf992351f82763cd76561c7c30466a5\nReviewed-on: https://code-review.googlesource.com/5051\nReviewed-by: Brad Fitzpatrick \u003cbradfitz@golang.org\u003e\n",
	}); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := log.Next, "0bacdc65dfd3dae28a124e21ecebb6f3f3c22087"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

type mockTripper func(*http.Request) (*http.Response, error)

func (t mockTripper) RoundTrip(req *http.Request) (*http.Response, error) { return t(req) }
