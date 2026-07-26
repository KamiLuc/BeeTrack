package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

var (
	ErrApiaryNotFound = errors.New("apiary not found")
	ErrHiveNotFound   = errors.New("hive not found")
)

// authorizeHive fetches hiveID and confirms userID belongs to its apiary,
// mapping either lookup's miss to ErrHiveNotFound.
func authorizeHive(ctx context.Context, apiaries ApiaryLister, hives HiveLister, userID, hiveID int64) (*model.Hive, error) {
	hive, err := hives.GetByID(ctx, hiveID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHiveNotFound
		}
		return nil, fmt.Errorf("get hive: %w", err)
	}
	if _, _, err := apiaries.GetMembership(ctx, hive.ApiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHiveNotFound
		}
		return nil, fmt.Errorf("get apiary membership: %w", err)
	}
	return hive, nil
}
