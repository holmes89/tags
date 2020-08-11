package rest

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/holmes89/tags/internal"
	"github.com/holmes89/tags/internal/database"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type tagHandler struct {
	repo database.Repository
}

func NewTagHandler(mr *mux.Router, repo database.Repository) http.Handler {
	r := mr.PathPrefix("/tag").Subrouter()

	h := &tagHandler{
		repo: repo,
	}

	r.HandleFunc("/", h.FindAll).Methods("GET")
	r.HandleFunc("/{id}", h.FindByID).Methods("GET")
	r.HandleFunc("/{id}/resources/", h.FindResourcesByTag).Methods("GET")
	r.HandleFunc("/", h.Create).Methods("POST")

	return r
}

func (h *tagHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	var params internal.TagParams
	if err := decoder.Decode(&params, r.URL.Query()); err != nil {
		logrus.WithError(err).Error("unable to parse params")
		EncodeError(w, http.StatusBadRequest, "tags", "unable to parse params", "find all")
		return
	}
	resp, err := h.repo.FindAllTags(params)
	if err != nil {
		EncodeError(w, http.StatusInternalServerError, "tags", "unable to find tags", "find all")
		return
	}
	EncodeJSONResponse(r.Context(), w, resp)
}

func (h *tagHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := vars["id"]
	resp, err := h.repo.FindTagByName(id)
	if err != nil {
		EncodeError(w, http.StatusInternalServerError, "tags", "unable to find tags", "find all")
		return
	}
	EncodeJSONResponse(r.Context(), w, resp)
}

func (h *tagHandler) Create(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	var tag internal.Tag
	if err := json.Unmarshal(b, &tag); err != nil {
		EncodeError(w, http.StatusBadRequest, "tag", "Bad Request from unmarshalling", "create")
		return
	}
	t, err := h.repo.CreateTag(tag)
	if err != nil {
		EncodeError(w, http.StatusInternalServerError, "tag", "failed to create tag", "create")
		return
	}
	EncodeJSONResponse(r.Context(), w, t)
}

func (h *tagHandler) FindResourcesByTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	params := internal.ResourceParams{Tag: id}
	resp, err := h.repo.FindAll(params)
	if err != nil {
		EncodeError(w, http.StatusInternalServerError, "tags", "unable to find tags", "find all resources")
		return
	}
	EncodeJSONResponse(r.Context(), w, resp)
}