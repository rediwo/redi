package filesystem

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// MemoryFileInfo represents file metadata
type MemoryFileData struct {
	content []byte
	modTime time.Time
}

// MemoryFileSystem in-memory file system for testing
type MemoryFileSystem struct {
	files map[string]*MemoryFileData
	mu    sync.RWMutex
}

// MemoryFileInfo implements fs.FileInfo for memory files
type MemoryFileInfo struct {
	name string
	size int64
	mode fs.FileMode
	time time.Time
}

func (m *MemoryFileInfo) Name() string       { return m.name }
func (m *MemoryFileInfo) Size() int64        { return m.size }
func (m *MemoryFileInfo) Mode() fs.FileMode  { return m.mode }
func (m *MemoryFileInfo) ModTime() time.Time { return m.time }
func (m *MemoryFileInfo) IsDir() bool        { return m.mode.IsDir() }
func (m *MemoryFileInfo) Sys() any           { return nil }

// NewMemoryFileSystem creates a new in-memory file system
func NewMemoryFileSystem() *MemoryFileSystem {
	return &MemoryFileSystem{
		files: make(map[string]*MemoryFileData),
	}
}

// WriteFile adds a file to the memory filesystem
func (mfs *MemoryFileSystem) WriteFile(name string, data []byte) error {
	mfs.mu.Lock()
	defer mfs.mu.Unlock()
	mfs.files[name] = &MemoryFileData{
		content: data,
		modTime: time.Now(),
	}
	return nil
}

func (mfs *MemoryFileSystem) ReadFile(name string) ([]byte, error) {
	mfs.mu.RLock()
	defer mfs.mu.RUnlock()
	
	fileData, exists := mfs.files[name]
	if !exists {
		return nil, fs.ErrNotExist
	}
	
	// Return a copy to prevent modification
	result := make([]byte, len(fileData.content))
	copy(result, fileData.content)
	return result, nil
}

func (mfs *MemoryFileSystem) Stat(name string) (fs.FileInfo, error) {
	mfs.mu.RLock()
	defer mfs.mu.RUnlock()
	
	fileData, exists := mfs.files[name]
	if !exists {
		return nil, fs.ErrNotExist
	}
	
	return &MemoryFileInfo{
		name: filepath.Base(name),
		size: int64(len(fileData.content)),
		mode: 0644,
		time: fileData.modTime,
	}, nil
}

func (mfs *MemoryFileSystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	mfs.mu.RLock()
	defer mfs.mu.RUnlock()
	
	for path := range mfs.files {
		if hasPathPrefix(path, root) {
			info, err := mfs.Stat(path)
			if err != nil {
				return err
			}
			
			dirEntry := &memoryDirEntry{info: info}
			if err := fn(path, dirEntry, nil); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (mfs *MemoryFileSystem) Sub(dir string) (FileSystem, error) {
	mfs.mu.RLock()
	defer mfs.mu.RUnlock()
	
	subFS := NewMemoryFileSystem()
	
	for path, fileData := range mfs.files {
		if hasPathPrefix(path, dir) {
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				continue
			}
			subFS.WriteFile(relPath, fileData.content)
		}
	}
	
	return subFS, nil
}

func (mfs *MemoryFileSystem) GetFS() fs.FS {
	return &memoryFS{mfs: mfs}
}

func (mfs *MemoryFileSystem) IsReadOnly() bool {
	return false
}

func (mfs *MemoryFileSystem) GetRoot() string {
	// Memory filesystem has no physical root
	return "."
}

func (mfs *MemoryFileSystem) ReadDir(dirname string) ([]fs.DirEntry, error) {
	mfs.mu.RLock()
	defer mfs.mu.RUnlock()

	// Get all entries in the directory
	entries := make(map[string]fs.DirEntry)
	dirPrefix := dirname
	if dirPrefix != "" && !strings.HasSuffix(dirPrefix, "/") {
		dirPrefix += "/"
	}

	for path, fileData := range mfs.files {
		// Skip if not in the target directory
		if dirname != "" && !strings.HasPrefix(path, dirPrefix) {
			continue
		}

		// Get the relative path from the directory
		relPath := path
		if dirname != "" {
			relPath = strings.TrimPrefix(path, dirPrefix)
		}

		// Skip if it's in a subdirectory
		if strings.Contains(relPath, "/") {
			// This is in a subdirectory, add the subdirectory itself
			subdirName := strings.Split(relPath, "/")[0]
			if _, exists := entries[subdirName]; !exists {
				// Create a directory entry
				entries[subdirName] = &memoryDirEntry{
					info: &MemoryFileInfo{
						name: subdirName,
						size: 0,
						time: time.Now(),
						mode: fs.ModeDir | 0755,
					},
				}
			}
		} else if relPath != "" {
			// This is a file in the directory
			entries[relPath] = &memoryDirEntry{
				info: &MemoryFileInfo{
					name: filepath.Base(relPath),
					size: int64(len(fileData.content)),
					time: fileData.modTime,
					mode: 0644,
				},
			}
		}
	}

	// Convert map to slice
	result := make([]fs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, entry)
	}

	return result, nil
}

// hasPathPrefix checks if path has the given prefix (replaces deprecated filepath.HasPrefix)
func hasPathPrefix(path, prefix string) bool {
	if prefix == "" {
		return true
	}
	return len(path) >= len(prefix) && path[:len(prefix)] == prefix && (len(path) == len(prefix) || path[len(prefix)] == '/')
}

// memoryDirEntry implements fs.DirEntry
type memoryDirEntry struct {
	info fs.FileInfo
}

func (m *memoryDirEntry) Name() string               { return m.info.Name() }
func (m *memoryDirEntry) IsDir() bool                { return m.info.IsDir() }
func (m *memoryDirEntry) Type() fs.FileMode          { return m.info.Mode().Type() }
func (m *memoryDirEntry) Info() (fs.FileInfo, error) { return m.info, nil }

// memoryFS implements fs.FS interface for MemoryFileSystem
type memoryFS struct {
	mfs *MemoryFileSystem
}

func (m *memoryFS) Open(name string) (fs.File, error) {
	// Clean the path - remove leading slash if present
	cleanName := name
	if len(cleanName) > 0 && cleanName[0] == '/' {
		cleanName = cleanName[1:]
	}
	
	// Handle root directory
	if cleanName == "" || cleanName == "." {
		// Return a directory listing
		return &memoryDir{mfs: m.mfs, path: "."}, nil
	}
	
	// Check if it's a directory
	isDir := false
	m.mfs.mu.RLock()
	for path := range m.mfs.files {
		if hasPathPrefix(path, cleanName) && path != cleanName {
			isDir = true
			break
		}
	}
	m.mfs.mu.RUnlock()
	
	if isDir {
		return &memoryDir{mfs: m.mfs, path: cleanName}, nil
	}
	
	// Try to open as file
	data, err := m.mfs.ReadFile(cleanName)
	if err != nil {
		return nil, err
	}
	
	info, err := m.mfs.Stat(cleanName)
	if err != nil {
		return nil, err
	}
	
	return &memoryFile{
		data: data,
		info: info,
	}, nil
}

// memoryDir implements fs.File for directories
type memoryDir struct {
	mfs  *MemoryFileSystem
	path string
}

func (m *memoryDir) Stat() (fs.FileInfo, error) {
	return &MemoryFileInfo{
		name: m.path,
		size: 0,
		mode: fs.ModeDir | 0755,
		time: time.Now(),
	}, nil
}

func (m *memoryDir) Read([]byte) (int, error) {
	return 0, errors.New("cannot read directory")
}

func (m *memoryDir) Close() error {
	return nil
}

// memoryFile implements fs.File
type memoryFile struct {
	data []byte
	info fs.FileInfo
	pos  int
}

func (m *memoryFile) Stat() (fs.FileInfo, error) {
	return m.info, nil
}

func (m *memoryFile) Read(p []byte) (int, error) {
	if m.pos >= len(m.data) {
		return 0, errors.New("EOF")
	}
	
	n := copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

func (m *memoryFile) Close() error {
	return nil
}