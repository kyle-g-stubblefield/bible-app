package api

import (
	"kyle-g-stubblefield/bible-app/pkg/server"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
    server.ApiRequestHandler(w, r)
}