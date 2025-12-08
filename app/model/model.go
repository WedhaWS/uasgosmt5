package model

type WebResponse[T any] struct {
	Status string        `json:"status"`
	Data   T             `json:"data"`
	Paging *PageMetaData `json:"paging,omitempty"`
	Errors string        `json:"errors,omitempty"`
}

type PageMetaData struct {
	Page      int   `json:"page"`
	Size      int   `json:"size"`
	TotalItem int64 `json:"total_item"`
	TotalPage int   `json:"total_page"`
}