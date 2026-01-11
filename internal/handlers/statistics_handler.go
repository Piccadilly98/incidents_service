package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Piccadilly98/incidents_service/internal/error_worker"
	"github.com/Piccadilly98/incidents_service/internal/service"
)

type StatisticHandler struct {
	serv *service.Service
	ew   *error_worker.ErrorWorker
}

func NewStatisticHandler(serv *service.Service, ew *error_worker.ErrorWorker) (*StatisticHandler, error) {
	if serv == nil {
		return nil, fmt.Errorf("service cannot be nil")
	}
	if ew == nil {
		return nil, fmt.Errorf("error worker cannot be nil")
	}

	return &StatisticHandler{
		serv: serv,
		ew:   ew,
	}, nil
}

func (sh *StatisticHandler) Handler(w http.ResponseWriter, r *http.Request) {
	result, err := sh.serv.GetChecksStatistics(r.Context())
	if err != nil {
		processingError(w, err, sh.ew)
		return
	}

	b, err := json.Marshal(result)
	if err != nil {
		processingError(w, err, sh.ew)
		return
	}

	w.Header().Set(HeaderContentType, HeaderJson)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
