package helper

// SuccessResponse adalah struktur standar untuk respons API yang berhasil.
// `json:"..."` adalah struct tag yang menentukan nama field saat di-encode ke JSON.
type SuccessResponse struct {
	Success bool        `json:"success"`          // Menandakan status permintaan (selalu true).
	Message string      `json:"message"`          // Pesan yang menjelaskan hasil operasi.
	Data    interface{} `json:"data,omitempty"` // Data utama dari respons (misalnya, satu objek atau array objek). `omitempty` berarti field ini tidak akan muncul di JSON jika nilainya kosong/null.
}

// ErrorResponse adalah struktur standar untuk respons API yang gagal.
type ErrorResponse struct {
	Success bool   `json:"success"` // Menandakan status permintaan (selalu false).
	Message string `json:"message"` // Pesan yang menjelaskan error yang terjadi.
}

