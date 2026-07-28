package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type ApiarySummary struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Role            string     `json:"role"`
	HiveCount       int        `json:"hive_count"`
	LastInspectedAt *time.Time `json:"last_inspected_at"`
}

// ListApiaries returns every apiary userID belongs to, including ones with
// zero hives — unlike list_hives, which only surfaces apiaries through
// their hives and so silently omits empty ones.
func (t *HiveTools) ListApiaries(ctx context.Context, userID int64) ([]ApiarySummary, error) {
	memberships, err := t.apiaries.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list apiaries: %w", err)
	}

	summaries := make([]ApiarySummary, len(memberships))
	for i, m := range memberships {
		summaries[i] = ApiarySummary{
			ID:              m.Apiary.ID,
			Name:            m.Apiary.Name,
			Role:            m.UserRole,
			HiveCount:       m.HiveCount,
			LastInspectedAt: m.LastInspectedAt,
		}
	}
	return summaries, nil
}

func (t *HiveTools) ListApiariesTool() Tool {
	return Tool{
		Name:        "list_apiaries",
		Description: "List every apiary the caller belongs to, with its name, role, hive count, and last inspection date. Includes apiaries with zero hives, unlike list_hives.",
		InputSchema: InputSchema{},
		Handler: func(ctx context.Context, userID int64, _ json.RawMessage) (any, error) {
			return t.ListApiaries(ctx, userID)
		},
	}
}
