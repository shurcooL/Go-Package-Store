package main

import (
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	gpscomponent "github.com/shurcooL/Go-Package-Store/component"
	"github.com/shurcooL/Go-Package-Store/frontend/model"
)

func main() {
	vecty.RenderBody(&UpdatesBody{})
}

type UpdatesBody struct {
	vecty.Core
}

func (*UpdatesBody) Render() vecty.ComponentOrHTML {
	return elem.Body(
		gpscomponent.UpdatesContent(
			mockComponentRPs,
			true,
		)...,
	)
}

var mockComponentRPs = []*model.RepoPresentation{
	{
		RepoRoot:          "github.com/gopherjs/gopherjs",
		ImportPathPattern: "github.com/gopherjs/gopherjs/...",
		LocalRevision:     "",
		RemoteRevision:    "",
		HomeURL:           "https://github.com/gopherjs/gopherjs",
		ImageURL:          "https://avatars.githubusercontent.com/u/6654647?v=3",
		Changes: []model.Change{
			{
				Message: "improved reflect support for blocking functions",
				URL:     "https://github.com/gopherjs/gopherjs/commit/87bf7e405aa3df6df0dcbb9385713f997408d7b9",
				Comments: model.Comments{
					Count: 0,
					URL:   "",
				},
			},
			{
				Message: "small cleanup",
				URL:     "https://github.com/gopherjs/gopherjs/commit/77a838f965881a888416bae38f790f76bb1f64bd",
				Comments: model.Comments{
					Count: 1,
					URL:   "https://www.example.com/",
				},
			},
			{
				Message: "replaced js.This and js.Arguments by js.MakeFunc",
				URL:     "https://github.com/gopherjs/gopherjs/commit/29dd054a0753760fe6e826ded0982a1bf69f702a",
				Comments: model.Comments{
					Count: 0,
					URL:   "",
				},
			},
		},
		Error:           "",
		UpdateState:     model.Available,
		UpdateSupported: true,
	},
	{
		RepoRoot:          "golang.org/x/image",
		ImportPathPattern: "golang.org/x/image/...",
		LocalRevision:     "",
		RemoteRevision:    "",
		HomeURL:           "http://golang.org/x/image",
		ImageURL:          "https://avatars.githubusercontent.com/u/4314092?v=3",
		Changes: []model.Change{
			{
				Message: "draw: generate code paths for image.Gray sources.",
				URL:     "https://github.com/golang/image/commit/f510ad81a1256ee96a2870647b74fa144a30c249",
				Comments: model.Comments{
					Count: 0,
					URL:   "",
				},
			},
		},
		Error:           "",
		UpdateState:     model.Updating,
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
		UpdateState:       model.Available,
		UpdateSupported:   true,
	},
	{
		RepoRoot:          "golang.org/x/image",
		ImportPathPattern: "golang.org/x/image/...",
		LocalRevision:     "",
		RemoteRevision:    "",
		HomeURL:           "http://golang.org/x/image",
		ImageURL:          "https://avatars.githubusercontent.com/u/4314092?v=3",
		Changes: []model.Change{
			{
				Message: "draw: generate code paths for image.Gray sources.",
				URL:     "https://github.com/golang/image/commit/f510ad81a1256ee96a2870647b74fa144a30c249",
				Comments: model.Comments{
					Count: 0,
					URL:   "",
				},
			},
		},
		Error:           "",
		UpdateState:     model.Updated,
		UpdateSupported: true,
	},
}
