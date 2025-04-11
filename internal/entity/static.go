package entity

type Static struct {
	ID        int    `json:"id"`
	FilePath  string `json:"file_path"`
	FileName  string `json:"file_name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
