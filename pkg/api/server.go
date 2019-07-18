package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/KnutZuidema/godarr/pkg/db"
	"github.com/KnutZuidema/godarr/pkg/model"
)

const (
	idPathParameter    = "id"
	defaultPagingCount = 20
)

type Error struct {
	Message    string `json:"message"`
	Link       string `json:"link,omitempty"`
	StatusCode int    `json:"-"`
}

var (
	ErrInvalidRequestBody = &Error{
		Message:    "Could not decode request body",
		StatusCode: http.StatusBadRequest,
	}
	ErrEncodeResponse = &Error{
		Message:    "Could not encode response body",
		StatusCode: http.StatusInternalServerError,
	}
)

type Server struct {
	router     *mux.Router
	db         db.Database
	logger     log.FieldLogger
	addedItems chan<- model.Item
}

func New(db db.Database, addedItems chan<- model.Item) *Server {
	s := &Server{
		db:         db,
		logger:     log.StandardLogger(),
		addedItems: addedItems,
	}
	s.router = s.setupRouter()
	return s
}

func (s Server) setupRouter() *mux.Router {
	router := mux.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Now().Sub(start)
			s.logger.WithFields(log.Fields{
				"path":     r.URL.Path,
				"method":   r.Method,
				"from":     r.RemoteAddr,
				"duration": duration,
			}).Info("completed request")
		})
	})
	router.HandleFunc("/item/{id}", s.errorHandler(s.getItem)).Methods(http.MethodGet)
	router.HandleFunc("/item", s.errorHandler(s.addItem)).Methods(http.MethodPost)
	return router
}

func (s Server) errorHandler(f func(http.ResponseWriter, *http.Request) *Error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err1 := f(w, r); err1 != nil {
			w.WriteHeader(err1.StatusCode)
			if err2 := json.NewEncoder(w).Encode(err1); err2 != nil {
				s.logger.WithFields(log.Fields{
					"path":   r.URL.Path,
					"status": err1.StatusCode,
					"error":  err1.Message,
				}).Error(err2)
			}
		}
	}
}

func (s Server) getItem(w http.ResponseWriter, r *http.Request) *Error {
	id, ok := mux.Vars(r)[idPathParameter]
	if !ok {
		return &Error{
			Message:    "Could not find ID in path",
			StatusCode: http.StatusBadRequest,
		}
	}
	item, err := s.db.GetItem(id)
	if err != nil {
		return &Error{
			Message:    "Could not find item",
			StatusCode: http.StatusNotFound,
		}
	}
	if err := json.NewEncoder(w).Encode(item); err != nil {
		return ErrEncodeResponse
	}
	return nil
}

func (s Server) addItem(w http.ResponseWriter, r *http.Request) *Error {
	var request model.Item
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return ErrInvalidRequestBody
	}
	if request.ExternalID == "" {
		return &Error{
			Message:    "External ID has to be specified",
			StatusCode: http.StatusBadRequest,
		}
	}
	if item, err := s.db.GetItemByExternalID(request.ExternalID); err != sql.ErrNoRows {
		if err == nil {
			return &Error{
				Message:    "Item already exists",
				Link:       "/item/" + item.ID,
				StatusCode: http.StatusConflict,
			}
		}
		return &Error{
			Message:    "Could not verify existence of item",
			StatusCode: http.StatusInternalServerError,
		}
	}
	item := model.Item{
		ExternalID: request.ExternalID,
		ID:         uuid.NewV4().String(),
		Status:     model.ItemStatusAdded,
	}
	if _, err := s.db.CreateItem(&item); err != nil {
		return &Error{
			Message:    "Could not add item",
			StatusCode: http.StatusInternalServerError,
		}
	}
	timer := time.NewTimer(10 * time.Second)
	select {
	case s.addedItems <- item:
		w.WriteHeader(http.StatusCreated)
		return nil
	case <-timer.C:
		return &Error{
			Message:    "Timed out while trying to add item",
			StatusCode: http.StatusInternalServerError,
		}
	}
}

func (s Server) listItems(w http.ResponseWriter, r *http.Request) *Error {
	var paging model.Paging
	if err := json.NewDecoder(r.Body).Decode(&paging); err != nil {
		return ErrInvalidRequestBody
	}
	if paging.Count == 0 {
		paging.Count = defaultPagingCount
	}
	items, err := s.db.ListItems(paging.Offset, paging.Count)
	if err != nil {
		return &Error{
			Message:    "Could not list items",
			StatusCode: http.StatusInternalServerError,
		}
	}
	if err := json.NewEncoder(w).Encode(items); err != nil {
		return ErrEncodeResponse
	}
	return nil
}
