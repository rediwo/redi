package filesystem

import "io/fs"

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
	// GetRoot returns the root directory path for this filesystem
	GetRoot() string
	// ReadDir reads the directory named by dirname
	ReadDir(dirname string) ([]fs.DirEntry, error)
}