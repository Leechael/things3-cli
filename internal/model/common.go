package model

// PaginatedResponse wraps list results for stable machine parsing.
type PaginatedResponse[T any] struct {
	Count   int `json:"count"`
	Results []T `json:"results"`
}

// URLCommandResult describes a dispatched Things URL command.
type URLCommandResult struct {
	Command    string `json:"command"`
	URL        string `json:"url"`
	Dispatched bool   `json:"dispatched"`
	Message    string `json:"message,omitempty"`
}
