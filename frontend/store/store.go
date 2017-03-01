package store

import (
	"fmt"

	"github.com/shurcooL/Go-Package-Store/frontend/action"
	gpscomponent "github.com/shurcooL/Go-Package-Store/vcomponent"
)

//var Store = struct {
//	//rpsMu sync.Mutex // TODO: Move towards a channel-based unified state manipulator.
//	RPs   []*gpscomponent.RepoPresentation
//
//	CheckingUpdates bool
//}{CheckingUpdates: true}

var (
	//rpsMu sync.Mutex // TODO: Move towards a channel-based unified state manipulator.
	rps []*gpscomponent.RepoPresentation

	checkingUpdates = true
)

func RPs() []*gpscomponent.RepoPresentation { return rps }
func CheckingUpdates() bool                 { return checkingUpdates }

// Apply applies action a to the store.
func Apply(a action.Action) action.Response {
	switch a := a.(type) {
	case *action.AppendRP:
		rps = append(rps, a.RP)
		moveUp(rps, a.RP)
		return nil

	case *action.SetUpdating:
		for _, rp := range rps {
			if rp.RepoRoot == a.RepoRoot {
				rp.UpdateState = gpscomponent.Updating
				return nil
			}
		}
		panic(fmt.Errorf("RepoRoot %q was not found in store", a.RepoRoot))

	case *action.SetUpdatingAll:
		var repoRoots []string
		for _, rp := range rps {
			if rp.UpdateState == gpscomponent.Available {
				repoRoots = append(repoRoots, rp.RepoRoot)
				rp.UpdateState = gpscomponent.Updating
			}
		}
		// TODO: Instead of response, look into async-action-creators:
		//       -	http://redux.js.org/docs/advanced/AsyncActions.html#async-action-creators
		//       -	https://gophers.slack.com/archives/D02LBN6UW/p1488335043280451
		return &action.SetUpdatingAllResponse{RepoRoots: repoRoots}

	case *action.SetUpdated:
		moveDown(rps, a.RepoRoot)
		for _, rp := range rps {
			if rp.RepoRoot == a.RepoRoot {
				rp.UpdateState = gpscomponent.Updated
				return nil
			}
		}
		panic(fmt.Errorf("RepoRoot %q was not found in store", a.RepoRoot))

	case *action.DoneCheckingUpdates:
		checkingUpdates = false
		return nil

	default:
		panic(fmt.Errorf("%v (type %T) is not a valid action", a, a))
	}
}

// TODO: Both moveDown and moveUp can be inlined and simplified.

// moveDown moves root down the rps towards all other updated.
func moveDown(rps []*gpscomponent.RepoPresentation, root string) {
	var i int
	for ; rps[i].RepoRoot != root; i++ { // i is the current package about to be updated.
	}
	for ; i+1 < len(rps) && rps[i+1].UpdateState != gpscomponent.Updated; i++ {
		rps[i], rps[i+1] = rps[i+1], rps[i] // Swap the two.
	}
}

// moveUp moves last entry up the rps above all other updated entries, unless rp is already updated.
func moveUp(rps []*gpscomponent.RepoPresentation, rp *gpscomponent.RepoPresentation) {
	// TODO: The "unless rp is already updated" part might not be needed if more strict about possible cases.
	if rp.UpdateState == gpscomponent.Updated {
		return
	}
	for i := len(rps) - 1; i-1 >= 0 && rps[i-1].UpdateState == gpscomponent.Updated; i-- {
		rps[i], rps[i-1] = rps[i-1], rps[i] // Swap the two.
	}
}
