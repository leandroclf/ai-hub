package files

// Reference identifies immutable content; access URLs are generated separately.
type Reference struct {
	ID          string `json:"file_id"`
	Version     string `json:"version"`
	SHA256      string `json:"sha256"`
	Size        int64  `json:"size_bytes"`
	ContentType string `json:"content_type"`
	Purpose     string `json:"purpose"`
	State       string `json:"state"`
}
