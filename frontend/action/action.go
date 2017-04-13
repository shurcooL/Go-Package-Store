// Package action defines actions that can be applied to the data model in store.
package action

import "github.com/shurcooL/Go-Package-Store/frontend/model"

// Action represents any of the supported actions.
type Action interface{}

// Response represents any of the supported responses.
type Response interface{}

// AppendRP is an action for appending a single update to the end.
type AppendRP struct {
	RP *model.RepoPresentation
}

// SetUpdating is an action for setting an update with RepoRoot to updating state.
type SetUpdating struct {
	RepoRoot string
}

// SetUpdatingAll is an action for setting all available updates to updating state.
type SetUpdatingAll struct{}

// SetUpdatingAllResponse is the response from SetUpdatingAll action,
// listing RepoRoot of all updates that were affected.
type SetUpdatingAllResponse struct {
	RepoRoots []string
}

// SetUpdated is an action for setting an update with RepoRoot to updated state.
type SetUpdated struct {
	RepoRoot string
}

// DoneCheckingUpdates is an action for when the update checking process is completed.
type DoneCheckingUpdates struct{}
