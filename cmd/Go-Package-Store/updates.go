package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shurcooL/Go-Package-Store/frontend/model"
	"github.com/shurcooL/httperror"
)

func updatesHandler(w http.ResponseWriter, req *http.Request) error {
	if req.Method != "GET" {
		return httperror.Method{Allowed: []string{"GET"}}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if !production {
		// TODO, XXX: Clean this up.
		http.ServeFile(w, req, "/Users/Dmitri/Dropbox/Needs Processing/GPS bits/page/updates/testdata/updates.json")
		return nil
	}
	jw := json.NewEncoder(w)
	jw.SetIndent("", "\t")
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("ResponseWriter %v is not a Flusher", w)
	}
	for rp := range c.pipeline.RepoPresentations() {
		var cs []model.Change
		for _, c := range rp.Presentation.Changes {
			cs = append(cs, model.Change{
				Message:  c.Message,
				URL:      c.URL,
				Comments: model.Comments{Count: c.Comments.Count, URL: c.Comments.URL},
			})
		}
		repoPresentation := model.RepoPresentation{
			RepoRoot:          rp.Repo.Root,
			ImportPathPattern: rp.Repo.ImportPathPattern(),
			LocalRevision:     rp.Repo.Local.Revision,
			RemoteRevision:    rp.Repo.Remote.Revision,
			HomeURL:           rp.Presentation.HomeURL,
			ImageURL:          rp.Presentation.ImageURL,
			Changes:           cs,
			UpdateState:       model.UpdateState(rp.UpdateState),
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
