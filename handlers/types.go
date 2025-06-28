package handlers

// Route represents a single route in the application
type Route struct {
	Path      string
	FilePath  string
	FileType  string
	IsDynamic bool
	ParamName string
}