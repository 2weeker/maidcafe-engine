package models

const maxPage = 10000

type Pager struct {
	Page    int
	PerPage int
	HasNext bool
}

func NewPager(page, perPage int) Pager {
	return Pager{Page: min(max(page, 1), maxPage), PerPage: perPage}
}

func (p Pager) Limit() int    { return p.PerPage + 1 }
func (p Pager) Offset() int   { return (p.Page - 1) * p.PerPage }
func (p Pager) HasPrev() bool { return p.Page > 1 }

func Trim[T any](p *Pager, items []T) []T {
	if len(items) > p.PerPage {
		p.HasNext = true
		return items[:p.PerPage]
	}
	return items
}

func PageOf(n, perPage int) int {
	return (max(n, 1)-1)/perPage + 1
}
