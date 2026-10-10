// Package handler contains the HTTP layer: parse/validate input, call a service, write the response.
package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/posumkm/backend/internal/middleware"
	"github.com/posumkm/backend/internal/service"
	"github.com/posumkm/backend/pkg/apperror"
	"github.com/posumkm/backend/pkg/pagination"
	"github.com/posumkm/backend/pkg/response"
	"github.com/posumkm/backend/pkg/validator"
)

// bindJSON decodes and validates the request body. On failure it writes the error and returns false.
func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		response.Error(c, apperror.BadRequest("Format JSON tidak valid"))
		return false
	}
	if err := validator.Struct(dst); err != nil {
		response.Error(c, err)
		return false
	}
	return true
}

func pathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, apperror.BadRequest("ID tidak valid"))
		return 0, false
	}
	return id, true
}

func actor(c *gin.Context) service.Actor {
	return service.Actor{UserID: middleware.UserID(c), Role: middleware.Role(c)}
}

// listParams reads page, limit, search, sort and order.
func listParams(c *gin.Context) pagination.Params {
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	p := pagination.Params{Page: page, Limit: limit, Search: c.Query("search"), Sort: c.Query("sort"), Order: c.Query("order")}
	p.Normalize()
	return p
}

// queryInt64 parses an optional integer query parameter.
func queryInt64(c *gin.Context, key string) (*int64, error) {
	v := c.Query(key)
	if v == "" {
		return nil, nil
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return nil, apperror.Validation([]apperror.FieldError{{Field: key, Message: "harus berupa angka"}})
	}
	return &n, nil
}

func queryBool(c *gin.Context, key string) (*bool, error) {
	v := c.Query(key)
	if v == "" {
		return nil, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return nil, apperror.Validation([]apperror.FieldError{{Field: key, Message: "harus true atau false"}})
	}
	return &b, nil
}

// queryEnum returns the upper-cased value if it is one of allowed (or empty).
func queryEnum(c *gin.Context, key string, allowed ...string) (string, error) {
	v := strings.ToUpper(strings.TrimSpace(c.Query(key)))
	if v == "" {
		return "", nil
	}
	for _, a := range allowed {
		if v == a {
			return v, nil
		}
	}
	return "", apperror.Validation([]apperror.FieldError{{Field: key, Message: "harus salah satu dari: " + strings.Join(allowed, ", ")}})
}

// dateFilter parses optional inclusive date_from/date_to (YYYY-MM-DD) into [from, to+1d).
func dateFilter(c *gin.Context, loc *time.Location) (*time.Time, *time.Time, error) {
	var from, to *time.Time
	if v := c.Query("date_from"); v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, loc)
		if err != nil {
			return nil, nil, apperror.Validation([]apperror.FieldError{{Field: "date_from", Message: "format tanggal harus YYYY-MM-DD"}})
		}
		from = &t
	}
	if v := c.Query("date_to"); v != "" {
		t, err := time.ParseInLocation("2006-01-02", v, loc)
		if err != nil {
			return nil, nil, apperror.Validation([]apperror.FieldError{{Field: "date_to", Message: "format tanggal harus YYYY-MM-DD"}})
		}
		next := t.AddDate(0, 0, 1)
		to = &next
	}
	return from, to, nil
}
