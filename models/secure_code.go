package models

type SecureCode struct {
	Id       string `json:"id"`
	Metadata any    `json:"metadata"`
	Secret   string `json:"-"`
}
