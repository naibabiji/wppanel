package models

type WPExtension struct {
	ID      int    `json:"id"`
	EType   string `json:"etype"`
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}
