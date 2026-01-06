package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Piccadilly98/incidents_service/internal/error_worker"
	"github.com/Piccadilly98/incidents_service/internal/service"
)

type DeactivateHandler struct {
	serv *service.Service
	ew   *error_worker.ErrorWorker
}

func NewDeactivateHandler(serv *service.Service, ew *error_worker.ErrorWorker) (*DeactivateHandler, error) {
	if serv == nil {
		return nil, fmt.Errorf("service cannot be nil")
	}
	if ew == nil {
		return nil, fmt.Errorf("error worker cannot be nil")
	}

	return &DeactivateHandler{
		serv: serv,
		ew:   ew,
	}, nil
}

func (d *DeactivateHandler) Handler(w http.ResponseWriter, r *http.Request) {
	id := checkURLParam(w, r, d.ew)
	if id == "" {
		return
	}
	mod := r.Header.Get(HeaderDeactivateMode)
	if mod == HeaderDeactivateForce {
		err := d.serv.DeleteIncidentByID(r.Context(), id)
		if err != nil {
			processingError(w, err, d.ew)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	res, err := d.serv.DeactivateIncidentByID(r.Context(), id)
	if err != nil {
		processingError(w, err, d.ew)
		return
	}

	b, err := json.Marshal(res)
	if err != nil {
		processingError(w, err, d.ew)
		return
	}
	w.Header().Set(HeaderContentType, HeaderJson)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
