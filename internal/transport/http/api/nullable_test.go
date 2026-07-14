package api

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestNullableStringUnmarshal(t *testing.T) {
	t.Parallel()

	type payload struct {
		DueDate     nullableString `json:"due_date"`
		Description nullableString `json:"description"`
		Title       *string        `json:"title"`
	}

	var omitted payload
	if err := json.Unmarshal([]byte(`{"title":"x"}`), &omitted); err != nil {
		t.Fatal(err)
	}
	if omitted.DueDate.Set || omitted.Description.Set {
		t.Fatalf("omitted fields should not be set: %+v", omitted)
	}

	var nulled payload
	if err := json.Unmarshal([]byte(`{"due_date":null,"description":null}`), &nulled); err != nil {
		t.Fatal(err)
	}
	if !nulled.DueDate.Set || nulled.DueDate.Value != nil {
		t.Fatalf("due_date null: %+v", nulled.DueDate)
	}
	if !nulled.Description.Set || nulled.Description.Value != nil {
		t.Fatalf("description null: %+v", nulled.Description)
	}

	var valued payload
	if err := json.Unmarshal([]byte(`{"due_date":"2026-07-14","description":" hi "}`), &valued); err != nil {
		t.Fatal(err)
	}
	if !valued.DueDate.Set || valued.DueDate.Value == nil || *valued.DueDate.Value != "2026-07-14" {
		t.Fatalf("due_date value: %+v", valued.DueDate)
	}
	if !valued.Description.Set || valued.Description.Value == nil || *valued.Description.Value != " hi " {
		t.Fatalf("description value: %+v", valued.Description)
	}

	// Ensure DisallowUnknownFields decoder path still works with raw bytes like respond.decodeJSON.
	dec := json.NewDecoder(bytes.NewReader([]byte(`{"description":null}`)))
	dec.DisallowUnknownFields()
	var viaDec payload
	if err := dec.Decode(&viaDec); err != nil {
		t.Fatal(err)
	}
	if !viaDec.Description.Set || viaDec.Description.Value != nil {
		t.Fatalf("decoder null: %+v", viaDec.Description)
	}
}
