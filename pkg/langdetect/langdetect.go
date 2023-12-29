package langdetect

import (
	"os"
	"path/filepath"
	"strings"
)

type (
	// ProjectType is the type of project
	ProjectType string
)

const (
	// GoModule is the project type for Go modules
	GoModule ProjectType = "GoModule"
	// Unknown is the project type for unknown project types
	Unknown ProjectType = "Unknown"
)

// ProjectDetails holds the details of the current project
type ProjectDetails struct {
	Entrypoint string
	Name       string
}

// FindProjectType detects the programming language/package manager of the current project.
func FindProjectType() (ProjectType, *ProjectDetails, error) {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}

	// List files in the current directory
	files, err := filepath.Glob(filepath.Join(currentDir, "*"))
	if err != nil {
		return "", nil, err
	}

	// Check for specific project files
	for _, file := range files {
		switch filepath.Base(file) {
		case "go.mod":
			// read the file and check if it has a module name
			f, err := os.ReadFile(file)
			if err != nil {
				return "", nil, err
			}
			var moduleName string
			lines := strings.Split(string(f), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "module") {
					moduleName = strings.TrimSpace(line[6:])
					break
				}
			}
			return GoModule, &ProjectDetails{
				Name: moduleName,
			}, nil
		}
	}

	return Unknown, nil, nil
}