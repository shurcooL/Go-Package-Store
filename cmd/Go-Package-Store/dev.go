// +build dev

package main

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/shurcooL/Go-Package-Store"
	gpscomponent "github.com/shurcooL/Go-Package-Store/component"
	"github.com/shurcooL/Go-Package-Store/workspace"
	"github.com/shurcooL/htmlg"
	"github.com/shurcooL/httperror"
)

import _ "net/http/pprof"

const production = false

func init() {
	http.Handle("/mock.html", errorHandler(mockHandler))
	http.Handle("/component.html", errorHandler(componentHandler))
	http.Handle("/vcomponent.html", errorHandler(vComponentHandler))
}

func mockHandler(w http.ResponseWriter, req *http.Request) error {
	if req.Method != "GET" {
		return httperror.Method{Allowed: []string{"GET"}}
	}

	// Reset the pipeline and populate it with mock repo presentations,
	// complete with artificial delays (to simulate processing time).
	c.pipeline = workspace.NewPipeline(wd)
	go func() {
		for _, rp := range mockWorkspaceRPs {
			time.Sleep(5 * time.Second)
			rp := rp
			c.pipeline.AddPresented(&rp)
		}
		time.Sleep(5 * time.Second)
		c.pipeline.Done()
	}()

	return indexHandler(w, req)
}

func vComponentHandler(w http.ResponseWriter, req *http.Request) error {
	if req.Method != "GET" {
		return httperror.Method{Allowed: []string{"GET"}}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, err := io.WriteString(w, `<html>
	<head>
		<title>Go Package Store</title>
		<link href="/assets/style.css" rel="stylesheet" type="text/css" />
		<script async src="/assets/vcomponent/vcomponent.js" type="text/javascript"></script>
	</head>
	<body></body></html>`)
	return err
}

func componentHandler(w http.ResponseWriter, req *http.Request) error {
	if req.Method != "GET" {
		return httperror.Method{Allowed: []string{"GET"}}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	_, err := io.WriteString(w, `<html>
	<head>
		<title>Go Package Store</title>
		<link href="/assets/style.css" rel="stylesheet" type="text/css" />
	</head>
	<body>`)
	if err != nil {
		return err
	}

	err = htmlg.RenderComponents(w, gpscomponent.Header{})
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, `<div class="center-max-width"><div class="content">`)
	if err != nil {
		return err
	}

	err = htmlg.RenderComponents(w, gpscomponent.UpdatesHeader{
		RPs:             nil,
		CheckingUpdates: false,
	})
	if err != nil {
		return err
	}

	err = htmlg.RenderComponents(w, gpscomponent.UpdatesHeader{
		RPs:             mockComponentRPs,
		CheckingUpdates: true,
	})
	if err != nil {
		return err
	}

	err = htmlg.RenderComponents(w,
		mockComponentRPs[0],
		mockComponentRPs[1],
		mockComponentRPs[2],
	)
	if err != nil {
		return err
	}

	err = htmlg.RenderComponents(w, gpscomponent.InstalledUpdates)
	if err != nil {
		return err
	}

	err = htmlg.RenderComponents(w, mockComponentRPs[3])
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, `</div></div>`)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, `</body></html>`)
	return err
}

var mockComponentRPs = []*gpscomponent.RepoPresentation{
	{
		RepoRoot:          "github.com/gopherjs/gopherjs",
		ImportPathPattern: "github.com/gopherjs/gopherjs/...",
		LocalRevision:     "",
		RemoteRevision:    "",
		HomeURL:           "https://github.com/gopherjs/gopherjs",
		ImageURL:          "https://avatars.githubusercontent.com/u/6654647?v=3",
		Changes: []gpscomponent.Change{
			{
				Message: "improved reflect support for blocking functions",
				URL:     "https://github.com/gopherjs/gopherjs/commit/87bf7e405aa3df6df0dcbb9385713f997408d7b9",
				Comments: gpscomponent.Comments{
					Count: 0,
					URL:   "",
				},
			},
			{
				Message: "small cleanup",
				URL:     "https://github.com/gopherjs/gopherjs/commit/77a838f965881a888416bae38f790f76bb1f64bd",
				Comments: gpscomponent.Comments{
					Count: 1,
					URL:   "https://www.example.com/",
				},
			},
			{
				Message: "replaced js.This and js.Arguments by js.MakeFunc",
				URL:     "https://github.com/gopherjs/gopherjs/commit/29dd054a0753760fe6e826ded0982a1bf69f702a",
				Comments: gpscomponent.Comments{
					Count: 0,
					URL:   "",
				},
			},
		},
		Error:           "",
		UpdateState:     gpscomponent.Available,
		UpdateSupported: true,
	},
	{
		RepoRoot:          "golang.org/x/image",
		ImportPathPattern: "golang.org/x/image/...",
		LocalRevision:     "",
		RemoteRevision:    "",
		HomeURL:           "http://golang.org/x/image",
		ImageURL:          "https://avatars.githubusercontent.com/u/4314092?v=3",
		Changes: []gpscomponent.Change{
			{
				Message: "draw: generate code paths for image.Gray sources.",
				URL:     "https://github.com/golang/image/commit/f510ad81a1256ee96a2870647b74fa144a30c249",
				Comments: gpscomponent.Comments{
					Count: 0,
					URL:   "",
				},
			},
		},
		Error:           "",
		UpdateState:     gpscomponent.Updating,
		UpdateSupported: true,
	},
	{
		RepoRoot:          "unknown.com/package",
		ImportPathPattern: "unknown.com/package/...",
		LocalRevision:     "abcdef0123456789000000000000000000000000",
		RemoteRevision:    "d34db33f01010101010101010101010101010101",
		HomeURL:           "https://unknown.com/package",
		ImageURL:          "https://github.com/images/gravatars/gravatar-user-420.png",
		Changes:           nil,
		Error:             "",
		UpdateState:       gpscomponent.Available,
		UpdateSupported:   true,
	},
	{
		RepoRoot:          "golang.org/x/image",
		ImportPathPattern: "golang.org/x/image/...",
		LocalRevision:     "",
		RemoteRevision:    "",
		HomeURL:           "http://golang.org/x/image",
		ImageURL:          "https://avatars.githubusercontent.com/u/4314092?v=3",
		Changes: []gpscomponent.Change{
			{
				Message: "draw: generate code paths for image.Gray sources.",
				URL:     "https://github.com/golang/image/commit/f510ad81a1256ee96a2870647b74fa144a30c249",
				Comments: gpscomponent.Comments{
					Count: 0,
					URL:   "",
				},
			},
		},
		Error:           "",
		UpdateState:     gpscomponent.Updated,
		UpdateSupported: true,
	},
}

var mockWorkspaceRPs = []workspace.RepoPresentation{
	{
		Repo: &gps.Repo{
			Root: (string)("github.com/gopherjs/gopherjs"),
		},
		Presentation: &gps.Presentation{
			HomeURL:  (string)("https://github.com/gopherjs/gopherjs"),
			ImageURL: (string)("https://avatars.githubusercontent.com/u/6654647?v=3"),
			Changes: ([]gps.Change)([]gps.Change{
				(gps.Change)(gps.Change{
					Message: (string)("improved reflect support for blocking functions"),
					URL:     (string)("https://github.com/gopherjs/gopherjs/commit/87bf7e405aa3df6df0dcbb9385713f997408d7b9"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("small cleanup"),
					URL:     (string)("https://github.com/gopherjs/gopherjs/commit/77a838f965881a888416bae38f790f76bb1f64bd"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(1),
						URL:   (string)("https://www.example.com/"),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("replaced js.This and js.Arguments by js.MakeFunc"),
					URL:     (string)("https://github.com/gopherjs/gopherjs/commit/29dd054a0753760fe6e826ded0982a1bf69f702a"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
			}),
		},
	},

	{
		Repo: &gps.Repo{
			Root: (string)("golang.org/x/image"),
		},
		Presentation: &gps.Presentation{
			HomeURL:  (string)("http://golang.org/x/image"),
			ImageURL: (string)("https://avatars.githubusercontent.com/u/4314092?v=3"),
			Changes: ([]gps.Change)([]gps.Change{
				(gps.Change)(gps.Change{
					Message: (string)("draw: generate code paths for image.Gray sources."),
					URL:     (string)("https://github.com/golang/image/commit/f510ad81a1256ee96a2870647b74fa144a30c249"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
			}),
		},
	},

	{
		Repo: &gps.Repo{
			Root: (string)("unknown.com/package"),
			Local: struct {
				RemoteURL string
				Revision  string
			}{Revision: "abcdef0123456789000000000000000000000000"},
			Remote: struct {
				RepoURL  string
				Branch   string
				Revision string
			}{Revision: "d34db33f01010101010101010101010101010101"},
		},
		Presentation: &gps.Presentation{
			HomeURL:  (string)("https://unknown.com/package"),
			ImageURL: (string)("https://github.com/images/gravatars/gravatar-user-420.png"),
		},
	},

	{
		Repo: &gps.Repo{
			Root: (string)("golang.org/x/foobar"),
		},
		Presentation: &gps.Presentation{
			HomeURL:  (string)("http://golang.org/x/foobar"),
			ImageURL: (string)("https://avatars.githubusercontent.com/u/4314092?v=3"),
			Changes:  ([]gps.Change)(nil),
			Error:    (error)(errors.New("something went wrong\n\nnew lines are kept -    spaces are too.")),
		},
	},

	{
		UpdateState: workspace.Updated,

		Repo: &gps.Repo{
			Root: (string)("github.com/influxdb/influxdb"),
		},
		Presentation: &gps.Presentation{
			HomeURL:  (string)("https://github.com/influxdb/influxdb"),
			ImageURL: (string)("https://avatars.githubusercontent.com/u/5713248?v=3"),
			Changes: ([]gps.Change)([]gps.Change{
				(gps.Change)(gps.Change{
					Message: (string)("Add link to \"How to Report Bugs Effectively\""),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/6f398c1daf88fe34faede69f4404a334202acae8"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Update CONTRIBUTING.md"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/37fa6056009dd4e84e9852ec50ce747e22375a99"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Update CONTRIBUTING.md"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/87a6a8f15a13c5bf0ac60608edc1be570e7b023e"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Add note about requiring distro details"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/901f91dc9559bebddf9b49607eac4ffd5caa4158"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(4),
						URL:   (string)("https://www.example.com/"),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Correct typo in change log"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/8eefdba0d3ef3ab5a408073ae275d495b67c9535"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Correct markdown for URL"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/41688ea6af78d45d051c7f6ac24a6468d36b9fad"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Update with PR1744"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/db09b20d199c973a209e181c9e2f890969bd0b57"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Merge pull request #1770 from kylezh/dev"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/a7c0d71d9ccadde17e7aa5cbba538b4a99670633"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Merge pull request #1787 from influxdb/measurement_batch_in_series"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/40479784e2bd690b9021ec730287c426124230dd"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Store Measurement commands in batches"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/a5749bebfb40239b8fd7b25d2ab1aa234c31c6b2"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Merge pull request #1786 from influxdb/remove-syslog"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/2facd6158620e86262407ae3c4c131860f6953c5"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Merge pull request #1785 from influxdb/1784"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/4a5fdcc9ea3bf6dc178f45758332b871e45b93eb"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Fix urlgen to work on Ubuntu"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/666d09367690627f9c3212c1c25c566416c645da"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Remove unused syslog.go"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/06bfd9c496becacff404e6768e7c0fd8ce9603c2"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Fix timezone abbreviation."),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/06eac99c230dcc24bee9c3e1c1ef01725ce017ad"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Merge pull request #1782 from influxdb/more_contains_unit_tests"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/fffbcf3fbe953e03e69ac1d22c142ecd6b3aba3b"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("More shard \"contains\" unit tests"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/ec93341f3fddd294f404fd1469fb651d4ba16e4c"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Update changelog for rc6 release"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/65b4d1a060883a5901bd7c40492a3345d2eabc77"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Merge pull request #1781 from influxdb/single_shard_data"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/5889b12832b2e43424951c92089db03f31df1078"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Refactor shard group time bound checking"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/05d630bfb8041362c89249e3e6fabe6261cecc66"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
				(gps.Change)(gps.Change{
					Message: (string)("Fix error when alter retention policy"),
					URL:     (string)("https://github.com/influxdb/influxdb/commit/9f8639ded8778a270cc99cf2d9ee1a09f635d67d"),
					Comments: (gps.Comments)(gps.Comments{
						Count: (int)(0),
						URL:   (string)(""),
					}),
				}),
			}),
		},
	},
}
