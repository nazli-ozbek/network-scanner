package model

type IPRange struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Range string `json:"range"`
}
