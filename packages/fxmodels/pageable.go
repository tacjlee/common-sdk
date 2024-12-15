package fxmodels

type Pageable struct {
	PageNumber int    `json:"pageNumber"`
	PageSize   int    `json:"pageSize"`
	Offset     int64  `json:"offset"`
	Order      string `json:"order"`
}
