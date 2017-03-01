package action

import gpscomponent "github.com/shurcooL/Go-Package-Store/vcomponent"

// Action represents any of the supported actions.
type Action interface{}

// Response represents any of the supported responses.
type Response interface{}

type AppendRP struct {
	RP *gpscomponent.RepoPresentation
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
