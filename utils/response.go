package utils

// APIResponse adalah struktur standar output JSON
type APIResponse struct {
	Meta Meta        `json:"meta"`
	Data interface{} `json:"data"`
}

type Meta struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Status  string `json:"status"`
}

// SuccessResponse untuk format sukses (Code 200-299)
func SuccessResponse(message string, code int, data interface{}) APIResponse {
	meta := Meta{
		Message: message,
		Code:    code,
		Status:  "success",
	}

	return APIResponse{
		Meta: meta,
		Data: data,
	}
}

// ErrorResponse untuk format gagal (Code 400-500)
func ErrorResponse(message string, code int, data interface{}) APIResponse {
	meta := Meta{
		Message: message,
		Code:    code,
		Status:  "error",
	}

	return APIResponse{
		Meta: meta,
		Data: data,
	}
}	