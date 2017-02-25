package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/shurcooL/Go-Package-Store"
	"github.com/shurcooL/httperror"
)

// updateHandler is a handler for update requests.
type updateHandler struct {
	updateRequests chan updateRequest

	// updater is set based on the source of Go packages. If nil, it means
	// we don't have support to update Go packages from the current source.
	// It's used to update repos in the backend, and if set to nil, to disable
	// the frontend UI for updating packages.
	updater gps.Updater
}

type updateRequest struct {
	root         string
	responseChan chan error
}

// ServeHTTP handles update requests.
func (u *updateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		httperror.HandleMethod(w, httperror.Method{Allowed: []string{"POST"}})
		return
	}

	root := req.PostFormValue("repo_root")

	updateRequest := updateRequest{
		root:         root,
		responseChan: make(chan error),
	}
	u.updateRequests <- updateRequest

	err := <-updateRequest.responseChan
	// TODO: Display error in frontend.
	if err != nil {
		log.Println(err)
	}
}

// Worker is a sequential updater of Go packages. It does not update them in parallel
// to avoid race conditions or other problems.
func (u *updateHandler) Worker() {
	for updateRequest := range u.updateRequests {
		c.pipeline.GoPackageList.Lock()
		repoPresentation, ok := c.pipeline.GoPackageList.List[updateRequest.root]
		c.pipeline.GoPackageList.Unlock()
		if !ok {
			updateRequest.responseChan <- fmt.Errorf("root %q not found", updateRequest.root)
			continue
		}
		if repoPresentation.Updated {
			updateRequest.responseChan <- fmt.Errorf("root %q already updated", updateRequest.root)
			continue
		}

		err := u.updater.Update(repoPresentation.Repo)
		if err == nil {
			// Mark repo as updated.
			c.pipeline.GoPackageList.Lock()
			// Move it down the OrderedList towards all other updated.
			{
				var i, j int
				for ; c.pipeline.GoPackageList.OrderedList[i].Repo.Root != updateRequest.root; i++ { // i is the current package about to be updated.
				}
				for j = len(c.pipeline.GoPackageList.OrderedList) - 1; c.pipeline.GoPackageList.OrderedList[j].Updated; j-- { // j is the last not-updated package.
				}
				c.pipeline.GoPackageList.OrderedList[i], c.pipeline.GoPackageList.OrderedList[j] =
					c.pipeline.GoPackageList.OrderedList[j], c.pipeline.GoPackageList.OrderedList[i]
			}
			c.pipeline.GoPackageList.List[updateRequest.root].Updated = true
			c.pipeline.GoPackageList.Unlock()
		}
		updateRequest.responseChan <- err
		fmt.Println("\nDone.")
	}
}
