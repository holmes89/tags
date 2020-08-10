package rest

import (
	"github.com/gorilla/mux"
	"net/http"
)

type tagHandler struct {
}

func NewTagHandler(mr *mux.Router) http.Handler {
	r := mr.PathPrefix("/tag").Subrouter()

	h := &tagHandler{
	}

	r.HandleFunc("/", h.FindAll).Methods("GET")
	r.HandleFunc("/{name}", h.Create).Methods("POST")
	r.HandleFunc("/{name}", h.Delete).Methods("DELETE")


	return r
}

func (h *tagHandler) FindAll(w http.ResponseWriter, r *http.Request) {

}

func (h *tagHandler) Create(w http.ResponseWriter, r *http.Request) {
}

func (h *tagHandler) Delete(w http.ResponseWriter, r *http.Request) {
}