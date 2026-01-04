package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Piccadilly98/incidents_service/internal/error_worker"
	"github.com/Piccadilly98/incidents_service/internal/models/dto"
	"github.com/Piccadilly98/incidents_service/internal/service"
)

type RegistrationHandler struct {
	serv *service.Service
	ew   *error_worker.ErrorWorker
}

func NewRegistrationHandler(serv *service.Service, ew *error_worker.ErrorWorker) (*RegistrationHandler, error) {
	if serv == nil {
		return nil, fmt.Errorf("service cannot be nil")
	}
	if ew == nil {
		return nil, fmt.Errorf("error worker cannot be nil")
	}

	return &RegistrationHandler{
		serv: serv,
		ew:   ew,
	}, nil
}

func (rh *RegistrationHandler) Handler(w http.ResponseWriter, r *http.Request) {
	if val, ok := r.Context().Value(ContextKeyValidApiKey).(bool); ok {
		if val != ContextValueValidApiKey {
			ErrorResponse(w, fmt.Errorf("invalid api-key"), http.StatusForbidden)
			return
		}
	} else {
		ErrorResponse(w, fmt.Errorf("invalid api-key"), http.StatusForbidden)
		return
	}

	if !checkHeaderJson(w, r) {
		return
	}
	model := &dto.RegistrationIncidentRequest{}
	err := json.NewDecoder(r.Body).Decode(model)
	if err != nil {
		processingError(w, err, rh.ew)
		return
	}

	res, err := rh.serv.RegistrationIncident(r.Context(), model)
	if err != nil {
		processingError(w, err, rh.ew)
		return
	}

	b, err := json.Marshal(res)
	if err != nil {
		processingError(w, err, rh.ew)
		return
	}
	w.Header().Set(HeaderContentType, HeaderJson)
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
}
