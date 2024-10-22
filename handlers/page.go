package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Page interface {
	Render(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (string, error)
}
