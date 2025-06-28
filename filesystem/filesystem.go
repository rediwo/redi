package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
)

// FileSystem unified file system interface
type FileSystem interface {
	ReadFile(name string) ([]byte, error)
	Stat(name string) (fs.FileInfo, error)
	WalkDir(root string, fn fs.WalkDirFunc) error
	Sub(dir string) (FileSystem, error)
	// GetFS returns the underlying fs.FS interface for integration with standard library
	GetFS() fs.FS
	// IsReadOnly returns whether the file system is read-only
	IsReadOnly() bool
}

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

// EmbedFileSystem file system based on embed.FS
type EmbedFileSystem struct {
	embedFS fs.FS
}

// NewEmbedFileSystem creates a file system based on embed.FS
func NewEmbedFileSystem(embedFS fs.FS) *EmbedFileSystem {
	return &EmbedFileSystem{embedFS: embedFS}
}

func (efs *EmbedFileSystem) ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(efs.embedFS, name)
}

func (efs *EmbedFileSystem) Stat(name string) (fs.FileInfo, error) {
	return fs.Stat(efs.embedFS, name)
}

func (efs *EmbedFileSystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	return fs.WalkDir(efs.embedFS, root, fn)
}

func (efs *EmbedFileSystem) Sub(dir string) (FileSystem, error) {
	subFS, err := fs.Sub(efs.embedFS, dir)
	if err != nil {
		return nil, err
	}
	return NewEmbedFileSystem(subFS), nil
}

func (efs *EmbedFileSystem) GetFS() fs.FS {
	return efs.embedFS
}

func (efs *EmbedFileSystem) IsReadOnly() bool {
	return true
}