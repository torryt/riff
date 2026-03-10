package internal

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ProjectInfo holds summary information about a single riff-managed project.
type ProjectInfo struct {
	ID          string
	Path        string
	Description string // empty if none
	Template    string // empty if none
	Created     string // ISO 8601 timestamp, empty if unknown
}

// ProjectMetadata is the on-disk structure stored in .riff.json.
type ProjectMetadata struct {
	Description string   `json:"description"`
	Created     string   `json:"created"`
	Template    string   `json:"template"`
	Tags        []string `json:"tags"`
}

// packageJSON is used only to extract the description field from package.json.
type packageJSON struct {
	Description string `json:"description"`
}

// EnsureRiffDir creates ~/.riff/ and ~/.riff/projects/ if they do not exist.
func EnsureRiffDir() error {
	return os.MkdirAll(ProjectsDir, 0o755)
}

// GetProjects reads all project directories inside ProjectsDir and returns a
// sorted slice of ProjectInfo. Hidden entries (names starting with ".") are
// skipped. For each project directory the function first tries to read
// .riff.json; if that is absent it falls back to package.json for the
// description field.
func GetProjects() ([]ProjectInfo, error) {
	entries, err := os.ReadDir(ProjectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var projects []ProjectInfo

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files/directories.
		if len(name) > 0 && name[0] == '.' {
			continue
		}

		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(ProjectsDir, name)
		info := ProjectInfo{
			ID:   name,
			Path: projectPath,
		}

		// Attempt to read .riff.json first.
		meta, err := ReadMetadata(projectPath)
		if err == nil {
			info.Description = meta.Description
			info.Template = meta.Template
			info.Created = meta.Created
		} else {
			// Fall back to package.json for description only.
			pkgPath := filepath.Join(projectPath, "package.json")
			if data, readErr := os.ReadFile(pkgPath); readErr == nil {
				var pkg packageJSON
				if jsonErr := json.Unmarshal(data, &pkg); jsonErr == nil {
					info.Description = pkg.Description
				}
			}
		}

		projects = append(projects, info)
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].ID < projects[j].ID
	})

	return projects, nil
}

// GenerateID returns a random alphanumeric string of the given length.
// The character set is [a-zA-Z0-9] (62 characters). If length is <= 0, the
// default length of 7 is used.
func GenerateID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if length <= 0 {
		length = 7
	}

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range result {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			// Extremely unlikely; fall back to index 0 for that position.
			result[i] = charset[0]
			continue
		}
		result[i] = charset[n.Int64()]
	}

	return string(result)
}

// WriteCdPath writes path to CdPathFile, creating or truncating the file.
// The shell wrapper reads this file after riff exits to cd into the project.
func WriteCdPath(path string) error {
	return os.WriteFile(CdPathFile, []byte(path), 0o644)
}

// ReadMetadata reads and parses the .riff.json file inside projectPath.
func ReadMetadata(projectPath string) (ProjectMetadata, error) {
	metaPath := filepath.Join(projectPath, MetadataFile)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return ProjectMetadata{}, err
	}

	var meta ProjectMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return ProjectMetadata{}, err
	}

	return meta, nil
}

// WriteMetadata serialises meta as indented JSON and writes it to
// .riff.json inside projectPath, creating the file if necessary.
func WriteMetadata(projectPath string, meta ProjectMetadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	metaPath := filepath.Join(projectPath, MetadataFile)
	return os.WriteFile(metaPath, data, 0o644)
}

// FormatAge returns a human-readable relative time string for an ISO 8601
// timestamp. If the timestamp is empty or unparseable it returns "unknown".
func FormatAge(iso string) string {
	if iso == "" {
		return "unknown"
	}
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return "unknown"
	}

	dur := time.Since(t)
	if dur < 0 {
		return "just now"
	}

	seconds := int(math.Floor(dur.Seconds()))
	minutes := int(math.Floor(dur.Minutes()))
	hours := int(math.Floor(dur.Hours()))
	days := hours / 24

	switch {
	case seconds < 60:
		return "just now"
	case minutes == 1:
		return "1 minute ago"
	case minutes < 60:
		return fmt.Sprintf("%d minutes ago", minutes)
	case hours == 1:
		return "1 hour ago"
	case hours < 24:
		return fmt.Sprintf("%d hours ago", hours)
	case days == 1:
		return "1 day ago"
	case days < 30:
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 02, 2006")
	}
}
