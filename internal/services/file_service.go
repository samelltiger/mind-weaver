package services

import (
	"errors"
	"io/ioutil"
	"mind-weaver/internal/utils"
	"os"
	"path/filepath"
	"strings"
)

type FileService struct {
	allowedExtensions map[string]bool
}

type FileNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Lines    int        `json:"lines"`
	IsDir    bool       `json:"is_dir"`
	Children []FileNode `json:"children,omitempty"`
}

func NewFileService() *FileService {
	return &FileService{
		allowedExtensions: utils.GetAllowedExtensions(), // 举例： map[string]bool{".go":    true,".js":    true}
	}
}

func (fs *FileService) ValidatePath(path string) error {
	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// Check if it's a directory
	if !info.IsDir() {
		return errors.New("path is not a directory")
	}

	// Check if we have read access
	file, err := os.Open(path)
	if err != nil {
		return errors.New("cannot access directory: " + err.Error())
	}
	file.Close()

	return nil
}

func (fs *FileService) ReadFile(path string) (string, error) {
	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	// Check if it's a file
	if info.IsDir() {
		return "", errors.New("path is a directory, not a file")
	}

	// Read file
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (fs *FileService) GetProjectFiles(projectPath string, maxDepth int) (FileNode, error) {
	// Validate the project path
	if err := fs.ValidatePath(projectPath); err != nil {
		return FileNode{}, err
	}

	// Create the root node
	root := FileNode{
		Name:  filepath.Base(projectPath),
		Path:  projectPath,
		IsDir: true,
	}

	// Recursively build the file tree
	lines, err := fs.buildFileTree(&root, projectPath, 0, maxDepth)
	if err != nil {
		return FileNode{}, err
	}
	root.Lines = lines

	return root, nil
}

func (fs *FileService) buildFileTree(node *FileNode, path string, depth, maxDepth int) (int, error) {
	if depth > maxDepth {
		return 0, nil
	}

	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return 0, err
	}

	totalLines := 0
	for _, entry := range entries {
		// Skip hidden files and directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(path, entry.Name())
		child := FileNode{
			Name:  entry.Name(),
			Path:  fullPath,
			IsDir: entry.IsDir(),
		}

		if entry.IsDir() {
			// Recursively process directories
			childLines, err := fs.buildFileTree(&child, fullPath, depth+1, maxDepth)
			if err != nil {
				return totalLines, err
			}
			child.Lines = childLines // Set directory's line count to sum of its children
			node.Children = append(node.Children, child)
			totalLines += childLines // Accumulate directory's lines to parent's total
		} else {
			// Only include files with allowed extensions
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if fs.allowedExtensions[ext] {
				child.Lines, err = utils.CountFileLines(fullPath)
				if err != nil {
					// You might want to handle this error differently
					child.Lines = 0
				}
				node.Children = append(node.Children, child)
				totalLines += child.Lines
			}
		}
	}

	return totalLines, nil
}
