package model

// WebhookPayload is the top-level structure for the incoming webhook.
type WebhookPayload struct {
	Username string  `json:"username"`
	Content  string  `json:"content"`
	Embeds   []Embed `json:"embeds"`
}

// Embed represents the rich content block in the notification.
type Embed struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Description string  `json:"description"`
	Color       int     `json:"color"`
	Fields      []Field `json:"fields"`
	Footer      Footer  `json:"footer"`
}

// Field represents a key-value pair within an embed.
type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

// Footer contains the footer text of an embed.
type Footer struct {
	Text string `json:"text"`
}
