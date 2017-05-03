package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/shurcooL/httperror"
)

// errorHandler factors error handling out of the HTTP handler.
type errorHandler func(w http.ResponseWriter, req *http.Request) error

func (h errorHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	err := h(w, req)
	if err == nil {
		// Do nothing.
		return
	}
	if err, ok := httperror.IsMethod(err); ok {
		httperror.HandleMethod(w, err)
		return
	}
	if err, ok := httperror.IsRedirect(err); ok {
		http.Redirect(w, req, err.URL, http.StatusSeeOther)
		return
	}
	if err, ok := httperror.IsBadRequest(err); ok {
		httperror.HandleBadRequest(w, err)
		return
	}
	if err, ok := httperror.IsHTTP(err); ok {
		code := err.Code
		error := fmt.Sprintf("%d %s\n\n%v", code, http.StatusText(code), err)
		http.Error(w, error, code)
		return
	}
	if os.IsNotExist(err) {
		log.Println(err)
		http.Error(w, "404 Not Found\n\n"+err.Error(), http.StatusNotFound)
		return
	}
	if os.IsPermission(err) {
		log.Println(err)
		http.Error(w, "403 Forbidden\n\n"+err.Error(), http.StatusForbidden)
		return
	}

	log.Println(err)
	http.Error(w, "500 Internal Server Error\n\n"+err.Error(), http.StatusInternalServerError)
}
