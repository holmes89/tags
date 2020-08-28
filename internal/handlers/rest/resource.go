package rest

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/holmes89/tags/internal"
	"github.com/holmes89/tags/internal/database"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

var decoder = schema.NewDecoder()

type resourceHandler struct {
	repo database.Repository
}

func NewResourceHandler(mr *mux.Router, repo database.Repository) http.Handler {
	r := mr.PathPrefix("/resource").Subrouter()

	h := &resourceHandler{
		repo: repo,
	}

	r.HandleFunc("/", h.FindAll).Methods("GET")
	r.HandleFunc("/{id}", h.FindByID).Methods("GET")
	r.HandleFunc("/", h.Create).Methods("POST")

	return r
}

func (h *resourceHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	var params *internal.ResourceParams
	if len(r.URL.Query()) > 0 {
		p := internal.ResourceParams{}
		if err := decoder.Decode(&p, r.URL.Query()); err != nil {
			logrus.WithError(err).Error("unable to parse params")
			EncodeError(w, http.StatusBadRequest, "resources", "unable to parse params", "find all")
			return
		}
		params = &p
	}
	resp, err := h.repo.FindAllResources(params)
	if err != nil {
		EncodeError(w, http.StatusInternalServerError, "resources", "unable to find resources", "find all")
		return
	}
	EncodeJSONResponse(r.Context(), w, resp)
}

func (h *resourceHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := vars["id"]

	resp, err := h.repo.FindResourceByID(id)
	if err != nil {
		EncodeError(w, http.StatusInternalServerError, "resources", "unable to find resource", "find by id")
		return
	}

	EncodeJSONResponse(r.Context(), w, resp)
}

func (h *resourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	var resource internal.Resource
	if err := json.Unmarshal(b, &resource); err != nil {
		EncodeError(w, http.StatusBadRequest, "resources", "Bad Request from unmarshalling", "create")
		return
	}
	resp, err := h.repo.CreateResource(resource)
	switch err {
	case nil:
		EncodeJSONResponse(r.Context(), w, resp)
	case internal.ErrInvalid:
		EncodeError(w, http.StatusBadRequest, "resources", "missing fields", "create")
	case internal.ErrConflict:
		EncodeError(w, http.StatusConflict, "resources", "entity exists", "create")
	default:
		EncodeError(w, http.StatusInternalServerError, "resources", "failed to create resource", "create")
	}
}
