package themepack

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
)

// readZipEntries extracts all regular entries with safe relative paths,
// enforcing the count and size caps. Unsafe paths (absolute, dot-dot,
// backslashes) are rejected rather than skipped: a container that needs them
// is malformed.
func readZipEntries(data []byte) (map[string][]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("not a valid .peltontheme (zip) file: %w", err)
	}
	if len(r.File) > maxZipEntries {
		return nil, fmt.Errorf("container has more than %d entries", maxZipEntries)
	}
	files := make(map[string][]byte, len(r.File))
	var total int64
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name, err := safePath(f.Name)
		if err != nil {
			return nil, err
		}
		if f.UncompressedSize64 > maxEntryBytes {
			return nil, fmt.Errorf("%s is larger than %d MB", name, maxEntryBytes>>20)
		}
		total += int64(f.UncompressedSize64)
		if total > maxTotalBytes {
			return nil, fmt.Errorf("container unpacks to more than %d MB", maxTotalBytes>>20)
		}
		content, err := readZipFile(f)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		files[name] = content
	}
	return files, nil
}

// readZipFile reads one entry fully, capped at maxEntryBytes even if the
// declared size lied.
func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	content, err := io.ReadAll(io.LimitReader(rc, maxEntryBytes+1))
	if err != nil {
		return nil, err
	}
	if len(content) > maxEntryBytes {
		return nil, fmt.Errorf("entry larger than its declared size cap")
	}
	return content, nil
}

// safePath normalizes a zip entry name and rejects anything that could
// escape the extraction directory.
func safePath(name string) (string, error) {
	if strings.Contains(name, "\\") {
		return "", fmt.Errorf("entry %q uses backslashes", name)
	}
	clean := path.Clean(name)
	if path.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("entry %q escapes the container", name)
	}
	return clean, nil
}

// normalizePath cleans a manifest-referenced path the same way zip entry
// names are cleaned, so lookups match.
func normalizePath(p string) string {
	return path.Clean(strings.TrimPrefix(p, "./"))
}
