package mcp

import (
	"reflect"
	"testing"
)

func TestValidateRecordTypesEmptyReturnsAllValid(t *testing.T) {
	valid := []string{"a", "b", "c"}
	got, err := validateRecordTypes(nil, valid)
	if err != nil {
		t.Fatalf("validateRecordTypes returned error: %v", err)
	}
	if !reflect.DeepEqual(got, valid) {
		t.Errorf("expected %v, got %v", valid, got)
	}
}

func TestValidateRecordTypesDeduplicates(t *testing.T) {
	got, err := validateRecordTypes([]string{"a", "a", "b"}, []string{"a", "b"})
	if err != nil {
		t.Fatalf("validateRecordTypes returned error: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("expected [a b], got %v", got)
	}
}

func TestValidateRecordTypesRejectsUnknown(t *testing.T) {
	if _, err := validateRecordTypes([]string{"nonsense"}, []string{"a", "b"}); err == nil {
		t.Fatal("expected an error for an unknown record type")
	}
}
