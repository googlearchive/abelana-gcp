package main

import (
	"net/http"

	"github.com/go-martini/martini"
)

func init() {
	m := martini.Classic()
	m.Get("/", func() string {
		return "Hello world!"
	})
	http.Handle("/", m)
}
