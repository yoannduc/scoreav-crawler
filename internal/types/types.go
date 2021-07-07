package types

type Element struct {
	Pk     string `json:"pk"`
	Sk     string `json:"sk"`
	ID     string `json:"id"`
	Link   string `json:"link"`
	Title  string `json:"title"`
	SDesc  string `json:"short_description"`
	LDesc  string `json:"long_description,omitempty"`
	Date   string `json:"date"`
	Source string `json:"source"`
	Type   string `json:"type"`
}

type Event struct {
	Source string `json:"source"`
	Type   string `json:"type"`
}
