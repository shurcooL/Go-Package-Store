package store

import (
	"fmt"

	"github.com/shurcooL/Go-Package-Store/frontend/action"
	"github.com/shurcooL/Go-Package-Store/frontend/model"
)

//var Store = struct {
//	//rpsMu sync.Mutex // TODO: Move towards a channel-based unified state manipulator.
//	RPs   []*model.RepoPresentation
//
//	CheckingUpdates bool
//}{CheckingUpdates: true}

var (
	//rpsMu sync.Mutex // TODO: Move towards a channel-based unified state manipulator.
	rps []*model.RepoPresentation

	checkingUpdates = true
)

func RPs() []*model.RepoPresentation { return rps }
func CheckingUpdates() bool          { return checkingUpdates }

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
				rp.UpdateState = model.Updating
				return nil
			}
		}
		panic(fmt.Errorf("RepoRoot %q was not found in store", a.RepoRoot))

	case *action.SetUpdatingAll:
		var repoRoots []string
		for _, rp := range rps {
			if rp.UpdateState == model.Available {
				repoRoots = append(repoRoots, rp.RepoRoot)
				rp.UpdateState = model.Updating
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
				rp.UpdateState = model.Updated
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
func moveDown(rps []*model.RepoPresentation, root string) {
	var i int
	for ; rps[i].RepoRoot != root; i++ { // i is the current package about to be updated.
	}
	for ; i+1 < len(rps) && rps[i+1].UpdateState != model.Updated; i++ {
		rps[i], rps[i+1] = rps[i+1], rps[i] // Swap the two.
	}
}

// moveUp moves last entry up the rps above all other updated entries, unless rp is already updated.
func moveUp(rps []*model.RepoPresentation, rp *model.RepoPresentation) {
	// TODO: The "unless rp is already updated" part might not be needed if more strict about possible cases.
	if rp.UpdateState == model.Updated {
		return
	}
	for i := len(rps) - 1; i-1 >= 0 && rps[i-1].UpdateState == model.Updated; i-- {
		rps[i], rps[i-1] = rps[i-1], rps[i] // Swap the two.
	}
}
