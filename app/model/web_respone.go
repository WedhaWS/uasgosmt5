package model

// Sesuai Modul 6: Struktur Response Standar dengan Meta
type WebResponse struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *MetaInfo   `json:"meta,omitempty"` // Pointer agar bisa nil jika tidak ada pagination
}

type MetaInfo struct {
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
	TotalData int64  `json:"totalData"`
	TotalPage int    `json:"totalPage"`
	SortBy    string `json:"sortBy"`
	Order     string `json:"order"` // asc / desc
	Search    string `json:"search"`
}

// Struct untuk menampung parameter query (DTO)
type PaginationParam struct {
	Page   int
	Limit  int
	SortBy string
	Order  string
	Search string
}