package renderer

import (
	"encoding/json"
	"os"
	"time"

	"github.com/GLINCKER/stacklit/internal/schema"
)

const schemaURL = "https://stacklit.dev/schema/v1.json"

var version = "dev"

// WriteJSON populates metadata fields on idx and writes it as indented JSON to path.
func WriteJSON(idx *schema.Index, path string) error {
	idx.Schema = schemaURL
	idx.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	idx.StacklitVersion = version

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
