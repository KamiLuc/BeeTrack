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
// mapping either lookup's miss to ErrHiveNotFound. Also returns the apiary's
// name, so callers that authorize hives one at a time (rather than via
// resolveHives) can still label results by apiary.
func authorizeHive(ctx context.Context, apiaries ApiaryLister, hives HiveLister, userID, hiveID int64) (*model.Hive, string, error) {
	hive, err := hives.GetByID(ctx, hiveID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrHiveNotFound
		}
		return nil, "", fmt.Errorf("get hive: %w", err)
	}
	apiary, _, err := apiaries.GetMembership(ctx, hive.ApiaryID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrHiveNotFound
		}
		return nil, "", fmt.Errorf("get apiary membership: %w", err)
	}
	return hive, apiary.Name, nil
}
