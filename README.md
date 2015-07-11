Go Package Store [![Build Status](https://travis-ci.org/shurcooL/Go-Package-Store.svg?branch=master)](https://travis-ci.org/shurcooL/Go-Package-Store)
================

An app that displays updates for the Go packages in your GOPATH.

Installation
------------

```bash
go get -u github.com/shurcooL/Go-Package-Store
```

That will create a binary at `$GOPATH/bin/Go-Package-Store`. You should run it from a terminal where your `$GOPATH` is set.

Screenshot
----------

![](Screenshot.png)

Development
-----------

This package relies on `go generate` directives to process and statically embed assets. For development only, you'll need extra dependencies:

```bash
go get -u -d -tags=generate github.com/shurcooL/Go-Package-Store/...
go get -u -d -tags=js github.com/shurcooL/Go-Package-Store/...
```

Afterwards, you can build and run the package in development mode, where all assets are always read and processed from disk:

```bash
go build -tags=dev github.com/shurcooL/Go-Package-Store
```

When you're done with development, you should run `go generate` before committing:

```bash
go generate github.com/shurcooL/Go-Package-Store/...
```

License
-------

-	[MIT License](http://opensource.org/licenses/mit-license.php)
