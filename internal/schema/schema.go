package schema

type Index struct {
	Schema          string                `json:"$schema"`
	Version         string                `json:"version"`
	GeneratedAt     string                `json:"generated_at"`
	StacklitVersion string                `json:"stacklit_version"`
	MerkleHash      string                `json:"merkle_hash,omitempty"`
	Project         Project               `json:"project"`
	Tech            Tech                  `json:"tech"`
	Structure       Structure             `json:"structure"`
	Modules         map[string]ModuleInfo `json:"modules"`
	Dependencies    Dependencies          `json:"dependencies"`
	Git             GitInfo               `json:"git,omitempty"`
	Hints           Hints                 `json:"hints,omitempty"`
	Architecture    Architecture          `json:"architecture,omitempty"`
}

type Architecture struct {
	Pattern string `json:"pattern,omitempty"`   // e.g. "Clean Architecture"
	Summary string `json:"ai_summary,omitempty"` // AI-generated narrative
}

type Project struct {
	Name       string   `json:"name"`
	Root       string   `json:"root"`
	Type       string   `json:"type"`
	Workspaces []string `json:"workspaces,omitempty"`
}

type Tech struct {
	PrimaryLanguage string               `json:"primary_language"`
	Languages       map[string]LangStats `json:"languages"`
	Frameworks      []string             `json:"frameworks,omitempty"`
}

type LangStats struct {
	Files int `json:"files"`
	Lines int `json:"lines"`
}

type Structure struct {
	Entrypoints    []string          `json:"entrypoints"`
	TotalFiles     int               `json:"total_files"`
	TotalLines     int               `json:"total_lines"`
	KeyDirectories map[string]string `json:"key_directories,omitempty"`
}

type ModuleInfo struct {
	Purpose    string            `json:"purpose"`
	Language   string            `json:"language,omitempty"`
	Files      int               `json:"files"`
	Lines      int               `json:"lines"`
	FileList   []string          `json:"file_list,omitempty"`
	Exports    []string          `json:"exports,omitempty"`
	TypeDefs   map[string]string `json:"type_defs,omitempty"` // key type definitions
	DependsOn  []string          `json:"depends_on,omitempty"`
	DependedBy []string          `json:"depended_by,omitempty"`
	Activity   string            `json:"activity,omitempty"`
}

type Dependencies struct {
	Edges        [][2]string `json:"edges"`
	Entrypoints  []string    `json:"entrypoints"`
	MostDepended []string    `json:"most_depended,omitempty"`
	Isolated     []string    `json:"isolated,omitempty"`
}

type GitInfo struct {
	HotFiles []HotFile `json:"hot_files,omitempty"`
	Recent   []string  `json:"recent,omitempty"`
	Stable   []string  `json:"stable,omitempty"`
}

type HotFile struct {
	Path       string `json:"path"`
	Commits90d int    `json:"commits_90d"`
}

type Hints struct {
	AddFeature string   `json:"add_feature,omitempty"`
	TestCmd    string   `json:"test_command,omitempty"`
	EnvVars    []string `json:"env_vars,omitempty"`
	DoNotTouch []string `json:"do_not_touch,omitempty"`
}
