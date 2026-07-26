package mcp

import (
	"reflect"
	"testing"
)

func TestValidateAllowedValuesEmptyReturnsAllValid(t *testing.T) {
	valid := []string{"a", "b", "c"}
	got, err := validateAllowedValues(nil, valid, "thing")
	if err != nil {
		t.Fatalf("validateAllowedValues returned error: %v", err)
	}
	if !reflect.DeepEqual(got, valid) {
		t.Errorf("expected %v, got %v", valid, got)
	}
}

func TestValidateAllowedValuesDeduplicates(t *testing.T) {
	got, err := validateAllowedValues([]string{"a", "a", "b"}, []string{"a", "b"}, "thing")
	if err != nil {
		t.Fatalf("validateAllowedValues returned error: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("expected [a b], got %v", got)
	}
}

func TestValidateAllowedValuesRejectsUnknown(t *testing.T) {
	if _, err := validateAllowedValues([]string{"nonsense"}, []string{"a", "b"}, "thing"); err == nil {
		t.Fatal("expected an error for an unknown value")
	}
}
