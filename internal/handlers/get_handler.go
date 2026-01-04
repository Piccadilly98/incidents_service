package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Piccadilly98/incidents_service/internal/error_worker"
	"github.com/Piccadilly98/incidents_service/internal/service"
)

type GetHandler struct {
	serv *service.Service
	ew   *error_worker.ErrorWorker
}

func NewGetHandler(serv *service.Service, ew *error_worker.ErrorWorker) (*GetHandler, error) {
	if serv == nil {
		return nil, fmt.Errorf("service cannot be nil")
	}
	if ew == nil {
		return nil, fmt.Errorf("error worker cannot be nil")
	}

	return &GetHandler{
		serv: serv,
		ew:   ew,
	}, nil
}

func (g *GetHandler) Handler(w http.ResponseWriter, r *http.Request) {
	if val, ok := r.Context().Value(ContextKeyValidApiKey).(bool); ok {
		if val != ContextValueValidApiKey {
			ErrorResponse(w, fmt.Errorf("invalid api-key"), http.StatusForbidden)
			return
		}
	} else {
		ErrorResponse(w, fmt.Errorf("invalid api-key"), http.StatusForbidden)
		return
	}

	id := checkURLParam(w, r, g.ew)
	if id == "" {
		return
	}

	res, err := g.serv.GetIncidentInfoByID(r.Context(), id)
	if err != nil {
		processingError(w, err, g.ew)
		return
	}
	b, err := json.Marshal(res)
	if err != nil {
		processingError(w, err, g.ew)
		return
	}
	w.Header().Set(HeaderContentType, HeaderJson)
	w.WriteHeader(http.StatusOK)
	w.Write(b)

}
