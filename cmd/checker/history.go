package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// loadHistory reads the monitoring history from the specified JSON file.
// It returns an empty History map if the file doesn't exist or is empty.
// If the file exists and contains valid JSON, it parses and returns the History data.
// Returns an error if the file cannot be read (except for non-existent files)
// or if the JSON content is malformed.
func loadHistory(path string) (History, error) {
	h := History{}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return h, nil
		}
		return nil, err
	}
	if len(b) == 0 {
		return h, nil
	}
	if err := json.Unmarshal(b, &h); err != nil {
		return nil, err
	}
	return h, nil
}

// saveJSON marshals the given value to JSON with indentation and writes it to the specified file path.
// It creates any necessary parent directories if they don't exist.
// Returns an error if marshaling fails or the file cannot be written.
func saveJSON(path string, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}
	return os.WriteFile(path, b, 0644)
}

// limitHistoryPoints trims the history slice to keep only the most recent entries.
// If the slice has fewer points than the limit, it returns the original slice unchanged.
// Otherwise, it returns only the last 'limit' number of points, removing older entries.
// This is used to cap history storage to approximately 24 hours of data.
func limitHistoryPoints(points []HistoryPoint, limit int) []HistoryPoint {
	if len(points) <= limit {
		return points
	}
	return points[len(points)-limit:]
}
