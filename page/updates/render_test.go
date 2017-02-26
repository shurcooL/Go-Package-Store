package updates_test

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	gpscomponent "github.com/shurcooL/Go-Package-Store/component"
	"github.com/shurcooL/Go-Package-Store/page/updates"
)

func BenchmarkRenderBodyInnerHTML(b *testing.B) {
	rps, err := loadRPs()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := updates.RenderBodyInnerHTML(ioutil.Discard, rps, false)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func loadRPs() ([]*gpscomponent.RepoPresentation, error) {
	f, err := os.Open(filepath.Join("testdata", "updates.json"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var rps []*gpscomponent.RepoPresentation
	for dec := json.NewDecoder(f); ; {
		var rp gpscomponent.RepoPresentation
		err := dec.Decode(&rp)
		if err == io.EOF {
			break
		} else if err != nil {
			return rps, err
		}
		rps = append(rps, &rp)
	}
	return rps, nil
}
