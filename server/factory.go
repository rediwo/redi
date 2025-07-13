package server

import (
	"io/fs"

	"github.com/rediwo/redi"
)

// Factory creates redi servers with various configurations
type Factory struct{}

// NewFactory creates a new server factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateServer creates a standard redi server
func (f *Factory) CreateServer(config *Config) (*redi.Server, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	
	server := redi.NewServerWithVersion(config.Root, config.Port, config.Version)
	server.SetGzipEnabled(config.EnableGzip)
	server.SetGzipLevel(config.GzipLevel)
	if config.RoutesDir != "" {
		server.SetRoutesDir(config.RoutesDir)
	}
	server.SetCacheEnabled(config.EnableCache)
	
	return server, nil
}

// CreateEmbeddedServer creates an embedded redi server
func (f *Factory) CreateEmbeddedServer(config *EmbedConfig) (*redi.Server, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	
	server := redi.NewServerWithFSAndVersion(config.EmbedFS, config.Port, config.Version)
	server.SetGzipEnabled(config.EnableGzip)
	server.SetGzipLevel(config.GzipLevel)
	if config.RoutesDir != "" {
		server.SetRoutesDir(config.RoutesDir)
	}
	server.SetCacheEnabled(config.EnableCache)
	
	return server, nil
}

// CreateServerFromFS creates a server from a filesystem
func (f *Factory) CreateServerFromFS(embedFS fs.FS, port int, version string) (*redi.Server, error) {
	config := &EmbedConfig{
		EmbedFS:    embedFS,
		Port:       port,
		Version:    version,
		EnableGzip: true,
		GzipLevel:  -1,
	}
	
	return f.CreateEmbeddedServer(config)
}

// CreateServerFromRoot creates a server from a root directory
func (f *Factory) CreateServerFromRoot(root string, port int, version string) (*redi.Server, error) {
	config := &Config{
		Root:       root,
		Port:       port,
		Version:    version,
		EnableGzip: true,
		GzipLevel:  -1,
	}
	
	return f.CreateServer(config)
}

// Default factory instance
var DefaultFactory = NewFactory()