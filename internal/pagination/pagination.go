package pagination

type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type PaginatedResponse[T any] struct {
	Data       []T `json:"data"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func ValidatePaginationParams(params *PaginationParams) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 10
	}
}

func CalculateOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}

func CalculateTotalPages(total, pageSize int) int {
	return (total + pageSize - 1) / pageSize
}
