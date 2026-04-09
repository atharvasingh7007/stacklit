package renderer

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/GLINCKER/stacklit/assets"
	"github.com/GLINCKER/stacklit/internal/schema"
)

// WriteHTML renders idx as a self-contained HTML file with an interactive
// force-directed dependency graph and writes it to path.
func WriteHTML(idx *schema.Index, path string) error {
	dataJSON, err := json.Marshal(idx)
	if err != nil {
		return err
	}
	html := strings.Replace(assets.TemplateHTML, "{{STACKLIT_DATA}}", string(dataJSON), 1)
	return os.WriteFile(path, []byte(html), 0644)
}
