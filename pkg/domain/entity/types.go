package entity

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type SortOptions struct {
	Field      string
	Ascending  bool
	Descending bool
}
