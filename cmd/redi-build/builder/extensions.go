package builder

import (
	"strings"
)

// expandExtensions expands single-word extensions to full GitHub URLs
func expandExtensions(extensions []string) []string {
	var expanded []string
	
	for _, ext := range extensions {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			continue
		}
		
		// If it's a single word (no slashes), assume it's an official redi extension
		if !strings.Contains(ext, "/") {
			ext = "github.com/rediwo/redi-" + ext
		}
		
		expanded = append(expanded, ext)
	}
	
	return expanded
}