// Package merge provides shared helpers for updating shell env files with
// provider-specific credential blocks delimited by start/end markers.
package merge

import (
	"os"
)

// ClearBlock removes the content between startMarker and endMarker in path,
// leaving the markers in place but empty. It is a no-op if path is empty,
// the file does not exist, or the block is not present.
func ClearBlock(path, startMarker, endMarker string) error {
	if path == "" {
		return nil
	}
	existing := readFile(path)
	if existing == "" {
		return nil
	}
	if indexOf(existing, startMarker) == -1 {
		return nil
	}
	return IntoFile(path, startMarker, endMarker, "", "")
}

// IntoFile replaces the region between startMarker/endMarker in path with
// block, or appends a new region if the markers are absent. The file is
// created with the given header if it does not yet exist.
// It is a no-op when path is empty.
func IntoFile(path, startMarker, endMarker, block, header string) error {
	if path == "" {
		return nil
	}
	existing := readFile(path)
	if existing == "" {
		existing = header
	}
	start := startMarker + "\n"
	startIdx := indexOf(existing, start)
	endIdx := indexOf(existing, endMarker)

	var result string
	if startIdx == -1 || endIdx == -1 {
		sep := ""
		if len(existing) > 0 && existing[len(existing)-1] != '\n' {
			sep = "\n"
		}
		result = existing + sep + start + block + endMarker + "\n"
	} else {
		before := existing[:startIdx]
		after := existing[endIdx+len(endMarker):]
		if len(after) > 0 && after[0] == '\n' {
			after = after[1:]
		}
		result = before + start + block + endMarker + "\n" + after
	}
	return os.WriteFile(path, []byte(result), 0o600)
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
