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
		c.pipeline.GoPackageList.Lock()
		rp, ok := c.pipeline.GoPackageList.ByRoot[ur.Root]
		c.pipeline.GoPackageList.Unlock()
		if !ok {
			ur.ResponseChan <- fmt.Errorf("root %q not found", ur.Root)
			continue
		}
		if rp.UpdateState != workspace.Available {
			ur.ResponseChan <- fmt.Errorf("root %q not available for update: %v", ur.Root, rp.UpdateState)
			continue
		}

		// Mark repo as updating.
		c.pipeline.GoPackageList.Lock()
		c.pipeline.GoPackageList.ByRoot[ur.Root].UpdateState = workspace.Updating
		c.pipeline.GoPackageList.Unlock()

		updateError := u.updater.Update(rp.Repo)

		if updateError == nil {
			c.pipeline.GoPackageList.Lock()
			for i, rp := range c.pipeline.GoPackageList.Active {
				if rp.Repo.Root == ur.Root {
					// Remove from active.
					copy(c.pipeline.GoPackageList.Active[i:], c.pipeline.GoPackageList.Active[i+1:])
					c.pipeline.GoPackageList.Active = c.pipeline.GoPackageList.Active[:len(c.pipeline.GoPackageList.Active)-1]

					// Mark repo as updated.
					rp.UpdateState = workspace.Updated

					// Append to history.
					c.pipeline.GoPackageList.History = append(c.pipeline.GoPackageList.History, rp)

					break
				}
			}
			c.pipeline.GoPackageList.Unlock()
		}

		ur.ResponseChan <- updateError
		fmt.Println("\nDone.")
	}
}

// TODO: Currently lots of logic (for manipulating repo presentations as they
//       get updated, etc.) haphazardly present both in backend and frontend,
//       need to think about that. Probably want to unify workspace.RepoPresentation
//       and component.RepoPresentation types, maybe. Try it.
