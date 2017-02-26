package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	gpscomponent "github.com/shurcooL/Go-Package-Store/component"
	"github.com/shurcooL/httperror"
)

func updatesHandler(w http.ResponseWriter, req *http.Request) error {
	if req.Method != "GET" {
		return httperror.Method{Allowed: []string{"GET"}}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, req, "/Users/Dmitri/Dropbox/Work/2013/GoLand/src/github.com/shurcooL/Go-Package-Store/page/updates/testdata/updates.json")
	return nil
	jw := json.NewEncoder(w)
	jw.SetIndent("", "\t")
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("ResponseWriter %v is not a Flusher", w)
	}
	for rp := range c.pipeline.RepoPresentations() {
		var cs []gpscomponent.Change
		for _, c := range rp.Presentation.Changes {
			cs = append(cs, gpscomponent.Change{
				Message:  c.Message,
				URL:      c.URL,
				Comments: gpscomponent.Comments{Count: c.Comments.Count, URL: c.Comments.URL},
			})
		}
		repoPresentation := gpscomponent.RepoPresentation{
			RepoRoot:          rp.Repo.Root,
			ImportPathPattern: rp.Repo.ImportPathPattern(),
			LocalRevision:     rp.Repo.Local.Revision,
			RemoteRevision:    rp.Repo.Remote.Revision,
			HomeURL:           rp.Presentation.HomeURL,
			ImageURL:          rp.Presentation.ImageURL,
			Changes:           cs,
			UpdateState:       gpscomponent.UpdateState(rp.UpdateState),
			UpdateSupported:   c.updater != nil,
		}
		if err := rp.Presentation.Error; err != nil {
			repoPresentation.Error = err.Error()
		}
		err := jw.Encode(repoPresentation)
		if err != nil {
			return fmt.Errorf("error encoding repoPresentation: %v", err)
		}
		flusher.Flush()
	}
	return nil
}
