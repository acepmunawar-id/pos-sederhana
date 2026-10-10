// Package pagination holds list query parameters shared by every list endpoint.
package pagination

import "strings"

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

type Params struct {
	Page   int
	Limit  int
	Search string
	Sort   string
	Order  string // "asc" | "desc"
}

type Meta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// Normalize clamps page/limit into valid ranges and lowercases the order.
func (p *Params) Normalize() {
	if p.Page < 1 {
		p.Page = DefaultPage
	}
	if p.Limit < 1 {
		p.Limit = DefaultLimit
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
	p.Search = strings.TrimSpace(p.Search)
	p.Order = strings.ToLower(p.Order)
	if p.Order != "asc" {
		p.Order = "desc"
	}
}

func (p Params) Offset() int { return (p.Page - 1) * p.Limit }

func (p Params) Meta(total int64) Meta {
	return Meta{Page: p.Page, Limit: p.Limit, Total: total}
}
