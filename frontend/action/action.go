package action

import "github.com/shurcooL/Go-Package-Store/frontend/model"

// Action represents any of the supported actions.
type Action interface{}

// Response represents any of the supported responses.
type Response interface{}

type AppendRP struct {
	RP *model.RepoPresentation
}

type SetUpdating struct {
	RepoRoot string
}

type SetUpdatingAll struct{}

type SetUpdatingAllResponse struct {
	RepoRoots []string
}

type SetUpdated struct {
	RepoRoot string
}

type DoneCheckingUpdates struct{}
