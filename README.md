Go Package Store
================

[![Build Status](https://travis-ci.org/shurcooL/Go-Package-Store.svg?branch=master)](https://travis-ci.org/shurcooL/Go-Package-Store) [![GoDoc](https://godoc.org/github.com/shurcooL/Go-Package-Store?status.svg)](https://godoc.org/github.com/shurcooL/Go-Package-Store)

Go Package Store displays updates for the Go packages in your GOPATH.

![](Screenshot.png)

Installation
------------

```bash
go get -u github.com/shurcooL/Go-Package-Store/cmd/Go-Package-Store
```

That will create a binary at `$GOPATH/bin/Go-Package-Store`. You should run it from a terminal where your `$GOPATH` is set.

Usage
-----

```
Usage: Go-Package-Store [flags]
       [newline separated packages] | Go-Package-Store -stdin [flags]
  -dep string
    	Determine the list of Go packages from the specified Gopkg.toml file.
  -git-subrepo string
    	Look for Go packages vendored using git-subrepo in the specified vendor directory.
  -godeps string
    	Read the list of Go packages from the specified Godeps.json file.
  -http string
    	Listen for HTTP connections on this address. (default "localhost:7043")
  -stdin
    	Read the list of newline separated Go packages from stdin.

GitHub Access Token:
  To display updates for private repositories on GitHub, or when
  you've exceeded the unauthenticated rate limit, you can provide
  a GitHub access token for Go Package Store to use via the
  GO_PACKAGE_STORE_GITHUB_TOKEN environment variable.

Examples:
  # Check for updates for all Go packages in GOPATH.
  Go-Package-Store

  # Show updates for all golang.org/x/... packages.
  go list golang.org/x/... | Go-Package-Store -stdin

  # Show updates for all dependencies within Gopkg.toml constraints.
  Go-Package-Store -dep=/path/to/repo/Gopkg.toml

  # Show updates for all Go packages vendored using git-subrepo
  # in the specified vendor directory.
  Go-Package-Store -git-subrepo=/path/to/repo/vendor
```

Development
-----------

This package relies on `go generate` directives to process and statically embed assets. For development only, you may need extra dependencies. You can build and run the package in development mode, where all assets are always read and processed from disk:

```bash
go build -tags=dev github.com/shurcooL/Go-Package-Store/cmd/Go-Package-Store
```

When you're done with development, you should run `go generate` before committing:

```bash
go generate github.com/shurcooL/Go-Package-Store/...
```

Alternatives
------------

-	[GoFresh](https://github.com/divan/gofresh) - Console tool for checking and updating package dependencies (imports).

License
-------

-	[MIT License](https://opensource.org/licenses/mit-license.php)
