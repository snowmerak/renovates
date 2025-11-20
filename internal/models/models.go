package models

type Repository struct {
	Name     string `json:"repository"`
	Platform string `json:"platform"`
}

type RenovateResult struct {
	Repository string
	Success    bool
	Output     string
	Error      error
}
