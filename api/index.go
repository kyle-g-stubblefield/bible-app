package handler

import (
	"mr-stubblefield/bible-app/pkg/server"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	h := server.NewServer().Handler
	h.ServeHTTP(w, r)
}
