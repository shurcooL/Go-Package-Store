package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/shurcooL/Go-Package-Store"
	"github.com/shurcooL/Go-Package-Store/workspace"
	"github.com/shurcooL/httperror"
)

func NewUpdateWorker(updater gps.Updater) updateWorker {
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
		rp, ok := c.pipeline.GoPackageList.List[ur.Root]
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
		c.pipeline.GoPackageList.List[ur.Root].UpdateState = workspace.Updating
		c.pipeline.GoPackageList.Unlock()

		updateError := u.updater.Update(rp.Repo)

		if updateError == nil {
			// Move down and mark repo as updated.
			c.pipeline.GoPackageList.Lock()
			moveDown(c.pipeline.GoPackageList.OrderedList, ur.Root)
			c.pipeline.GoPackageList.List[ur.Root].UpdateState = workspace.Updated
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
//       Also probably want to try separating available updates from completed updates.
//       That should simplify some logic, and will make it easier to maintain history
//       of updates in the future.

// moveDown moves root down the orderedList towards all other updated.
func moveDown(orderedList []*workspace.RepoPresentation, root string) {
	var i int
	for ; orderedList[i].Repo.Root != root; i++ { // i is the current package about to be updated.
	}
	for ; i+1 < len(orderedList) && orderedList[i+1].UpdateState != workspace.Updated; i++ {
		orderedList[i], orderedList[i+1] = orderedList[i+1], orderedList[i] // Swap the two.
	}
}
