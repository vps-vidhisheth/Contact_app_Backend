package web

import (
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

func Paginate(w http.ResponseWriter, r *http.Request, db *gorm.DB, out interface{}, baseQuery *gorm.DB) {
	page := 1
	limit := 5

	if p := r.URL.Query().Get("page"); p != "" {
		if pi, err := strconv.Atoi(p); err == nil && pi > 0 {
			page = pi
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if li, err := strconv.Atoi(l); err == nil && li > 0 {
			limit = li
		}
	}
	offset := (page - 1) * limit

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		RespondError(w, err)
		return
	}

	if err := baseQuery.Limit(limit).Offset(offset).Find(out).Error; err != nil {
		RespondError(w, err)
		return
	}

	resp := map[string]interface{}{
		"data":  out,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	RespondJSON(w, http.StatusOK, resp)
}
