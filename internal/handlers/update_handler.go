package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Piccadilly98/incidents_service/internal/error_worker"
	"github.com/Piccadilly98/incidents_service/internal/models/dto"
	"github.com/Piccadilly98/incidents_service/internal/service"
)

type UpdateHandler struct {
	serv *service.Service
	ew   *error_worker.ErrorWorker
}

func NewUpdateHandler(serv *service.Service, ew *error_worker.ErrorWorker) (*UpdateHandler, error) {
	if serv == nil {
		return nil, fmt.Errorf("service cannot be nil")
	}
	if ew == nil {
		return nil, fmt.Errorf("error worker cannot be nil")
	}

	return &UpdateHandler{
		serv: serv,
		ew:   ew,
	}, nil
}

func (u *UpdateHandler) Handler(w http.ResponseWriter, r *http.Request) {
	if val, ok := r.Context().Value(ContextKeyValidApiKey).(bool); ok {
		if val != ContextValueValidApiKey {
			ErrorResponse(w, fmt.Errorf("invalid api-key"), http.StatusForbidden)
			return
		}
	} else {
		ErrorResponse(w, fmt.Errorf("invalid api-key"), http.StatusForbidden)
		return
	}

	id := checkURLParam(w, r, u.ew)
	if id == "" {
		return
	}

	req := &dto.UpdateRequest{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		processingError(w, err, u.ew)
		return
	}

	res, err := u.serv.UpdateIncidentByID(r.Context(), id, req)
	if err != nil {
		processingError(w, err, u.ew)
		return
	}

	b, err := json.Marshal(res)
	if err != nil {
		processingError(w, err, u.ew)
		return
	}
	w.Header().Set(HeaderContentType, HeaderJson)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
