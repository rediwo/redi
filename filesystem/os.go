package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
)

// OSFileSystem operating system-based file system
type OSFileSystem struct {
	root string
}

// NewOSFileSystem creates an operating system-based file system
func NewOSFileSystem(root string) *OSFileSystem {
	return &OSFileSystem{root: root}
}

func (osfs *OSFileSystem) ReadFile(name string) ([]byte, error) {
	fullPath := filepath.Join(osfs.root, name)
	return os.ReadFile(fullPath)
}

func (osfs *OSFileSystem) Stat(name string) (fs.FileInfo, error) {
	fullPath := filepath.Join(osfs.root, name)
	return os.Stat(fullPath)
}

func (osfs *OSFileSystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	fullPath := filepath.Join(osfs.root, root)
	return filepath.WalkDir(fullPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fn(path, d, err)
		}
		
		// Convert absolute path back to relative path from the filesystem root
		relPath, err := filepath.Rel(osfs.root, path)
		if err != nil {
			return fn(path, d, err)
		}
		
		// Normalize path separators to forward slashes for consistency
		relPath = filepath.ToSlash(relPath)
		
		return fn(relPath, d, nil)
	})
}

func (osfs *OSFileSystem) Sub(dir string) (FileSystem, error) {
	newRoot := filepath.Join(osfs.root, dir)
	return NewOSFileSystem(newRoot), nil
}

func (osfs *OSFileSystem) GetFS() fs.FS {
	return os.DirFS(osfs.root)
}

func (osfs *OSFileSystem) IsReadOnly() bool {
	return false
}