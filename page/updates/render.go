package updates

import (
	"fmt"
	"io"
	"time"

	gpscomponent "github.com/shurcooL/Go-Package-Store/component"
	"github.com/shurcooL/htmlg"
)

func RenderBodyInnerHTML(w io.Writer, rps []*gpscomponent.RepoPresentation, checkingUpdates bool) error {
	started := time.Now()
	defer func() { fmt.Println("RenderBodyInnerHTML:", time.Since(started), len(rps)) }()

	err := htmlg.RenderComponents(w, gpscomponent.Header{})
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, `<div class="center-max-width"><div class="content">`)
	if err != nil {
		return err
	}

	err = htmlg.RenderComponents(w, gpscomponent.UpdatesHeader{
		RPs:             rps,
		CheckingUpdates: checkingUpdates,
	})
	if err != nil {
		return err
	}

	wroteInstalledUpdates := false
	for _, rp := range rps {
		if rp.UpdateState == gpscomponent.Updated && !wroteInstalledUpdates {
			err = htmlg.RenderComponents(w, gpscomponent.InstalledUpdates)
			if err != nil {
				return err
			}
			wroteInstalledUpdates = true
		}

		err := htmlg.RenderComponents(w, rp)
		if err != nil {
			return err
		}
	}

	_, err = io.WriteString(w, `</div></div>`)
	return err
}
