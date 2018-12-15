package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/shurcooL/Go-Package-Store"
	"github.com/shurcooL/Go-Package-Store/workspace"
	"github.com/shurcooL/httperror"
)

func newUpdateWorker(updater gps.Updater) updateWorker {
	return updateWorker{
		updater:        updater,
		updateRequests: make(chan updateRequest),
	}
}

type updateWorker struct {
	updater        gps.Updater
	updateRequests chan updateRequest
}

type updateRequest struct {
	Root         string
	ResponseChan chan error
}

// Handler for update endpoint.
func (u updateWorker) Handler(w http.ResponseWriter, req *http.Request) error {
	if req.Method != "POST" {
		return httperror.Method{Allowed: []string{"POST"}}
	}

	ur := updateRequest{
		Root:         req.PostFormValue("RepoRoot"),
		ResponseChan: make(chan error),
	}
	u.updateRequests <- ur

	err := <-ur.ResponseChan
	// TODO: Display error in frontend.
	if err != nil {
		log.Println("update error:", err)
	}

	return nil
}

// Start performing sequential updates of Go packages. It does not update
// in parallel to avoid race conditions.
func (u updateWorker) Start() {
	go u.run()
}

func (u updateWorker) run() {
	for ur := range u.updateRequests {
		c.pipeline.Packages.Lock()
		rp, ok := c.pipeline.Packages.ByRoot[ur.Root]
		c.pipeline.Packages.Unlock()
		if !ok {
			ur.ResponseChan <- fmt.Errorf("root %q not found", ur.Root)
			continue
		}
		if rp.UpdateState != workspace.Available {
			ur.ResponseChan <- fmt.Errorf("root %q not available for update: %v", ur.Root, rp.UpdateState)
			continue
		}

		// Mark repo as updating.
		c.pipeline.Packages.Lock()
		c.pipeline.Packages.ByRoot[ur.Root].UpdateState = workspace.Updating
		c.pipeline.Packages.Unlock()

		updateError := u.updater.Update(rp.Repo)

		if updateError == nil {
			c.pipeline.Packages.Lock()
			for i, rp := range c.pipeline.Packages.Active {
				if rp.Repo.Root == ur.Root {
					// Remove from active.
					copy(c.pipeline.Packages.Active[i:], c.pipeline.Packages.Active[i+1:])
					c.pipeline.Packages.Active = c.pipeline.Packages.Active[:len(c.pipeline.Packages.Active)-1]

					// Mark repo as updated.
					rp.UpdateState = workspace.Updated

					// Append to history.
					c.pipeline.Packages.History = append(c.pipeline.Packages.History, rp)

					break
				}
			}
			c.pipeline.Packages.Unlock()
		}

		ur.ResponseChan <- updateError
		fmt.Println("\nDone.")
	}
}

// TODO: Currently lots of logic (for manipulating repo presentations as they
//       get updated, etc.) haphazardly present both in backend and frontend,
//       need to think about that. Probably want to unify workspace.RepoPresentation
//       and component.RepoPresentation types, maybe. Try it.
