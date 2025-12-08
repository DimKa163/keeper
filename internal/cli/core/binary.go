package core

type Binary struct {
	Name      string `json:"name"`
	MIMEType  string `json:"mime_type"`
	SizeBytes int64  `json:"size"`
	Content   []byte `json:"content"`
}
