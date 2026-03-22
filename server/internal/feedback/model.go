package feedback

import "time"

type Feedback struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"session_id"`
	PageURL    string    `json:"page_url"`
	ElementCSS string    `json:"element_css"`
	XPercent   float64   `json:"x_percent"`
	YPercent   float64   `json:"y_percent"`
	Comment    string    `json:"comment"`
	AuthorName string    `json:"author_name"`
	Resolved   bool      `json:"resolved"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateRequest struct {
	PageURL    string  `json:"page_url" binding:"required"`
	ElementCSS string  `json:"element_css"`
	XPercent   float64 `json:"x_percent"`
	YPercent   float64 `json:"y_percent"`
	Comment    string  `json:"comment" binding:"required"`
	AuthorName string  `json:"author_name"`
}
