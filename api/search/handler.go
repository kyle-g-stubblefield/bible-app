package search

import (
	"kyle-g-stubblefield/bible-app/pkg/server"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
    server.SearchRequestHandler(w, r)
}