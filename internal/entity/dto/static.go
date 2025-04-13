package dto

type UploadStaticResponse struct {
	ID        int    `json:"id"`
	Path      string `json:"path"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
