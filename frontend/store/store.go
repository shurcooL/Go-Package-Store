// Package store is a store for updates.
// Its contents can only be modified by appling actions.
package store

import (
	"fmt"

	"github.com/shurcooL/Go-Package-Store/frontend/action"
	"github.com/shurcooL/Go-Package-Store/frontend/model"
)

var (
	active          []*model.RepoPresentation // Latest at the end.
	history         []*model.RepoPresentation // Latest at the end.
	checkingUpdates = true
)

// Active returns the active repo presentations in store.
// Most recently added ones are last.
func Active() []*model.RepoPresentation { return active }

// History returns the historical repo presentations in store.
// Most recently added ones are last.
func History() []*model.RepoPresentation { return history }

// CheckingUpdates reports whether the process of checking for updates is still running.
func CheckingUpdates() bool { return checkingUpdates }

// Apply applies action a to the store.
func Apply(a action.Action) action.Response {
	switch a := a.(type) {
	case *action.AppendRP:
		switch a.RP.UpdateState {
		case model.Available, model.Updating:
			active = append(active, a.RP)
		case model.Updated:
			history = append(history, a.RP)
		}
		return nil

	case *action.SetUpdating:
		for _, rp := range active {
			if rp.RepoRoot == a.RepoRoot {
				rp.UpdateState = model.Updating
				return nil
			}
		}
		panic(fmt.Errorf("RepoRoot %q was not found in store", a.RepoRoot))

	case *action.SetUpdatingAll:
		var repoRoots []string
		for _, rp := range active {
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
		for i, rp := range active {
			if rp.RepoRoot == a.RepoRoot {
				// Remove from active.
				copy(active[i:], active[i+1:])
				active = active[:len(active)-1]

				// Set UpdateState.
				rp.UpdateState = model.Updated

				// Append to history.
				history = append(history, rp)

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
