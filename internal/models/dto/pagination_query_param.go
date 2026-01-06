package dto

import (
	"fmt"

	"github.com/google/uuid"
)

type PaginationQueryParams struct {
	PageNum *int
	Status  string
	ID      *string
	Name    string
	Type    string
	Radius  *int
}

func (p *PaginationQueryParams) Validate() error {
	if p.ID != nil {
		if *p.ID == "" {
			return fmt.Errorf("id cannot be empty")
		}
		_, err := uuid.Parse(*p.ID)
		if err != nil {
			return fmt.Errorf("%s: is not uuid\n", *p.ID)
		}
	}
	return nil
}
