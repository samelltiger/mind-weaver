package treesitter

// IgnoreController defines the interface for checking if files should be ignored.
type IgnoreController interface {
	// Match checks if a given path should be ignored.
	Match(path string) bool
	// FilterPaths filters a slice of paths, returning only those not ignored.
	FilterPaths(paths []string) []string
}

// NoopIgnoreController is a default implementation that ignores nothing.
type NoopIgnoreController struct{}

func (n *NoopIgnoreController) Match(path string) bool {
	return false // Never ignore
}

func (n *NoopIgnoreController) FilterPaths(paths []string) []string {
	return paths // Return all paths
}

// NewNoopIgnoreController creates a controller that doesn't ignore any files.
func NewNoopIgnoreController() IgnoreController {
	return &NoopIgnoreController{}
}

// You would replace NoopIgnoreController with a real implementation using
// a library like github.com/sabhiram/go-gitignore if needed.
