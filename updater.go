package gps

// Updater is able to update Go packages contained in repositories.
type Updater interface {
	// Update specified repository to latest version.
	Update(repo *Repo) error
}
