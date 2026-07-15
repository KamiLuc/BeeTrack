package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/beetrack/backend/internal/model"
)

func TestListingJSON_IncludesApiaryFields(t *testing.T) {
	lat, lng := 52.2297, 21.0122
	l := &model.Listing{
		ID:              5,
		ApiaryID:        ptrInt64(1),
		ApiaryName:      "Home apiary",
		ApiaryLat:       &lat,
		ApiaryLng:       &lng,
		ApiaryHiveCount: 3,
	}

	got := listingJSON(l)

	if got["apiary_lat"] != &lat {
		t.Errorf("expected apiary_lat %v, got %v", &lat, got["apiary_lat"])
	}
	if got["apiary_lng"] != &lng {
		t.Errorf("expected apiary_lng %v, got %v", &lng, got["apiary_lng"])
	}
	if got["apiary_hive_count"] != 3 {
		t.Errorf("expected apiary_hive_count 3, got %v", got["apiary_hive_count"])
	}
}

func TestListingJSON_NilApiaryGPS(t *testing.T) {
	l := &model.Listing{ID: 5}

	got := listingJSON(l)

	if got["apiary_lat"] != (*float64)(nil) {
		t.Errorf("expected apiary_lat nil, got %v", got["apiary_lat"])
	}
	if got["apiary_lng"] != (*float64)(nil) {
		t.Errorf("expected apiary_lng nil, got %v", got["apiary_lng"])
	}
	if got["apiary_hive_count"] != 0 {
		t.Errorf("expected apiary_hive_count 0, got %v", got["apiary_hive_count"])
	}
}

func ptrInt64(v int64) *int64 {
	return &v
}

func TestParseListingFilter_HasApiary(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "true sets HasApiary", url: "/listings?has_apiary=true", want: true},
		{name: "absent leaves HasApiary false", url: "/listings", want: false},
		{name: "other value leaves HasApiary false", url: "/listings?has_apiary=false", want: false},
		{name: "malformed value leaves HasApiary false", url: "/listings?has_apiary=1", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.url, nil)

			f := parseListingFilter(r)

			if f.HasApiary != tt.want {
				t.Errorf("expected HasApiary %v, got %v", tt.want, f.HasApiary)
			}
		})
	}
}
