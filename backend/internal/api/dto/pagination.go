package dto

type PageRequest struct {
	Page  int64 `query:"page"`
	Limit int64 `query:"limit"`
}

func (p *PageRequest) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
}

func (p PageRequest) Skip() int64 {
	return (p.Page - 1) * p.Limit
}

type PageMeta struct {
	Page       int64 `json:"page"`
	Limit      int64 `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

func NewPageMeta(page, limit, total int64) PageMeta {
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}
	return PageMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

type PaginatedResponse[T any] struct {
	Items []T      `json:"items"`
	Meta  PageMeta `json:"meta"`
}
