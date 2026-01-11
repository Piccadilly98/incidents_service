package dto

import "time"

type WebhookTask struct {
	Dto        LocationCheckResponse
	CountReTry int    `json:"count_retry"`
	Method     string `json:"method"`
	Url        string `json:"url"`
}

type ResultWebhookRequestDTO struct {
	Dto  LocationCheckResponse
	Date time.Time `json:"date_request"`
}

func (wt *WebhookTask) ToResultWebhookDto() *ResultWebhookRequestDTO {
	return &ResultWebhookRequestDTO{
		Dto:  wt.Dto,
		Date: time.Now().UTC(),
	}
}
