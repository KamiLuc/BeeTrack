package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

// ReportHandler serves apiary activity report exports.
type ReportHandler struct {
	reports *service.ReportService
}

// NewReportHandler creates a ReportHandler backed by the given service.
func NewReportHandler(reports *service.ReportService) *ReportHandler {
	return &ReportHandler{reports: reports}
}

type reportPDFRequest struct {
	HiveIDs    []int64  `json:"hive_ids"`
	Categories []string `json:"categories"`
	From       string   `json:"from"`
	To         string   `json:"to"`
}

// PDF handles POST /api/v1/apiaries/{id}/report/pdf — generates a PDF
// covering the requested hives, record categories, and date range, and
// streams it back as a file download.
func (h *ReportHandler) PDF(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary id")
		return
	}

	var req reportPDFRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	from, err := time.Parse("2006-01-02", req.From)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_DATE", "from must be a YYYY-MM-DD date")
		return
	}
	to, err := time.Parse("2006-01-02", req.To)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_DATE", "to must be a YYYY-MM-DD date")
		return
	}

	categories := make([]service.ReportCategory, 0, len(req.Categories))
	for _, raw := range req.Categories {
		c, err := service.ParseReportCategory(raw)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, "INVALID_CATEGORY", fmt.Sprintf("invalid category %q", raw))
			return
		}
		categories = append(categories, c)
	}

	report, err := h.reports.Generate(r.Context(), userID, apiaryID, service.ReportFilter{
		HiveIDs:    req.HiveIDs,
		Categories: categories,
		From:       from,
		To:         to,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrApiaryNotFound):
			respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
		case errors.Is(err, service.ErrHiveIDsRequired):
			respond.Error(w, http.StatusBadRequest, "HIVE_IDS_REQUIRED", err.Error())
		case errors.Is(err, service.ErrReportCategoriesRequired):
			respond.Error(w, http.StatusBadRequest, "CATEGORIES_REQUIRED", err.Error())
		case errors.Is(err, service.ErrInvalidDateRange):
			respond.Error(w, http.StatusBadRequest, "INVALID_DATE_RANGE", err.Error())
		default:
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	pdfBytes, err := renderReportPDF(report)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to render report")
		return
	}

	filename := fmt.Sprintf("raport-%s-%s.pdf", slugify(report.Apiary.Name), time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Write(pdfBytes)
}

func slugify(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			out = append(out, r)
		case r >= 'A' && r <= 'Z':
			out = append(out, r+32)
		case r == ' ' || r == '-' || r == '_':
			out = append(out, '-')
		}
	}
	return string(out)
}
