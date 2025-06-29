package filesystem

import "io/fs"

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