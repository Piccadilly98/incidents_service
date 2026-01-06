package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Piccadilly98/incidents_service/internal/error_worker"
	"github.com/Piccadilly98/incidents_service/internal/models/dto"
	"github.com/Piccadilly98/incidents_service/internal/service"
)

type GetPaginationHandler struct {
	serv *service.Service
	ew   *error_worker.ErrorWorker
}

func NewGetPaginationHadler(serv *service.Service,
	ew *error_worker.ErrorWorker) (*GetPaginationHandler, error) {
	if serv == nil {
		return nil, fmt.Errorf("service cannot be nil")
	}
	if ew == nil {
		return nil, fmt.Errorf("error worker cannot be nil")
	}

	return &GetPaginationHandler{
		serv: serv,
		ew:   ew,
	}, nil
}

func (gh *GetPaginationHandler) Handler(w http.ResponseWriter, r *http.Request) {
	params, err := gh.getValidQueryDTO(r)
	if err != nil {
		processingError(w, err, gh.ew)
		return
	}

	res, err := gh.serv.GetPagination(r.Context(), params)
	if err != nil {
		processingError(w, err, gh.ew)
		return
	}
	b, err := json.Marshal(res)
	if err != nil {
		processingError(w, err, gh.ew)
		return
	}
	w.Header().Set(HeaderContentType, HeaderJson)
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (gh *GetPaginationHandler) getValidQueryDTO(r *http.Request) (*dto.PaginationQueryParams, error) {
	res := &dto.PaginationQueryParams{}

	if str := r.URL.Query().Get(QueryParamIncidentID); str != "" {
		res.ID = &str
	}
	if str := r.URL.Query().Get(QueryParamPageNum); str != "" {
		num, err := strconv.Atoi(str)
		if err != nil {
			return nil, fmt.Errorf("invalid page_num: is not integer")
		}
		if num < 1 {
			return nil, fmt.Errorf("page cannot be < 1")
		}
		res.PageNum = &num
	}
	if str := r.URL.Query().Get(QueryParamRadius); str != "" {
		num, err := strconv.Atoi(str)
		if err != nil {
			return nil, fmt.Errorf("invalid radius: is not integer")
		}
		num++
		if num <= 0 {
			return nil, fmt.Errorf("radius cannot be <= 0")
		}
		res.Radius = &num
	}
	if str := r.URL.Query().Get(QueryParamType); str != "" {
		res.Type = str
	}
	if str := r.URL.Query().Get(QueryParamName); str != "" {
		res.Name = str
	}
	if str := r.URL.Query().Get(QueryParamStatus); str != "" {
		res.Status = str
	}

	err := res.Validate()
	if err != nil {
		return nil, err
	}
	return res, nil
}
