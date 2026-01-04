package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Piccadilly98/incidents_service/internal/error_worker"
	"github.com/Piccadilly98/incidents_service/internal/models/dto"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func processingError(w http.ResponseWriter, err error, ew *error_worker.ErrorWorker) {
	code, strErr := ew.ProcessError(err)
	if code == -1 {
		return
	}

	ErrorResponse(w, strErr, code)
}

func ErrorResponse(w http.ResponseWriter, err error, code int) {
	resp := dto.NewErrorDto(err)
	b, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set(HeaderContentType, HeaderJson)
	w.WriteHeader(code)
	w.Write(b)
}

func checkHeaderJson(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get(HeaderContentType) != HeaderJson {
		ErrorResponse(w, fmt.Errorf("invalid header content-type"), http.StatusBadRequest)
		return false
	}
	return true
}

func checkURLParam(w http.ResponseWriter, r *http.Request, ew *error_worker.ErrorWorker) string {
	id := chi.URLParam(r, URLParam)
	if id == "" {
		processingError(w, fmt.Errorf("invalid type incident_id: empty"), ew)
		return ""
	}
	if _, err := uuid.Parse(id); err != nil {
		processingError(w, fmt.Errorf("invalid type incident_id: not uuid"), ew)
		return ""
	}
	return id
}
