package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Piccadilly98/incidents_service/internal/error_worker"
	"github.com/Piccadilly98/incidents_service/internal/models/dto"
	"github.com/Piccadilly98/incidents_service/internal/service"
)

type LocationCheckHandler struct {
	serv *service.Service
	ew   *error_worker.ErrorWorker
}

func NewLocationCheckHandler(
	serv *service.Service,
	ew *error_worker.ErrorWorker,
) (*LocationCheckHandler, error) {
	if serv == nil {
		return nil, fmt.Errorf("service cannot be nil")
	}
	if ew == nil {
		return nil, fmt.Errorf("error worker cannot be nil")
	}

	return &LocationCheckHandler{
		serv: serv,
		ew:   ew,
	}, nil
}

func (lc *LocationCheckHandler) Handler(w http.ResponseWriter, r *http.Request) {
	if !checkHeaderJson(w, r) {
		return
	}

	req := &dto.LocationCheckRequest{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		processingError(w, err, lc.ew)
		return
	}

	res, err := lc.serv.LocationCheck(r.Context(), req)
	if err != nil {
		processingError(w, err, lc.ew)
		return
	}

	b, err := json.Marshal(res)
	if err != nil {
		processingError(w, err, lc.ew)
		return
	}
	w.Header().Set(HeaderContentType, HeaderJson)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
