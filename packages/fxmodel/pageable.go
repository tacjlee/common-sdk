package fxmodel

type Pageable struct {
	PageNumber int                    `json:"pageNumber"`
	PageSize   int                    `json:"pageSize"`
	Offset     int64                  `json:"offset"`
	Order      string                 `json:"order"`
	Filter     map[string]interface{} `json:"filter"`
}
