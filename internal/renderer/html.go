package renderer

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/glincker/stacklit/assets"
	"github.com/glincker/stacklit/internal/schema"
)

// WriteHTML renders idx as a self-contained HTML file with an interactive
// force-directed dependency graph and writes it to path.
func WriteHTML(idx *schema.Index, path string) error {
	dataJSON, err := json.Marshal(idx)
	if err != nil {
		return err
	}
	html := strings.Replace(assets.TemplateHTML, "{{STACKLIT_DATA}}", string(dataJSON), 1)
	html = strings.Replace(html, "{{LANG_ICONS_JS}}", assets.LangIconsJS, 1)
	return os.WriteFile(path, []byte(html), 0644)
}
