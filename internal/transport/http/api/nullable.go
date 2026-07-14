package api

import (
	"bytes"
	"encoding/json"
)

// nullableString distinguishes omitted JSON fields from explicit null/empty.
// UnmarshalJSON is only called when the key is present.
type nullableString struct {
	Set   bool
	Value *string
}

func (n *nullableString) UnmarshalJSON(data []byte) error {
	n.Set = true
	if bytes.Equal(data, []byte("null")) {
		n.Value = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n.Value = &s
	return nil
}
