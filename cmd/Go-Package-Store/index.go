package main

import (
	"html/template"
	"io"
	"net/http"

	"github.com/shurcooL/httperror"
)

var headHTML = template.Must(template.New("").Parse(`<html>
	<head>
		<title>Go Package Store</title>
		<link href="/assets/style.css" rel="stylesheet" type="text/css" />
		<script async src="/frontend.js" type="text/javascript"></script>
		{{if .Production}}<script type="text/javascript">
		  (function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){
		  (i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),
		  m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)
		  })(window,document,'script','https://www.google-analytics.com/analytics.js','ga');

		  ga('create', 'UA-56541369-2', 'auto');
		  ga('send', 'pageview');

		</script>{{end}}
	</head>
	<body>`))

func indexHandler(w http.ResponseWriter, req *http.Request) error {
	if req.Method != "GET" {
		return httperror.Method{Allowed: []string{"GET"}}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := struct{ Production bool }{production}
	err := headHTML.Execute(w, data)
	if err != nil {
		return err
	}

	err = renderInitialBody(w)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, `</body></html>`)
	return err
}

func renderInitialBody(w io.Writer) error {
	_, err := io.WriteString(w, `<div style="width: 100%; text-align: center; background-color: rgb(225, 235, 245);"><span style="background-color: rgb(209, 225, 240); padding: 15px; display: inline-block;">Updates</span></div><div class="center-max-width"><div class="content"><h2 style="text-align: center;">Checking for updates...</h2></div></div>`)
	return err
}
