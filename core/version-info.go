package corepkg

import (
	"runtime/debug"
	"time"
)

// VersionInfo holds version information about the current Go executable.
type VersionInfo struct {
	Version    string    `json:"version"`
	Revision   string    `json:"revision"`
	LastCommit time.Time `json:"lastcommit"`
	DirtyBuild bool      `json:"dirtybuild"`
}

func NewVersionInfo() *VersionInfo {

	vi := &VersionInfo{Version: "0.0.0", Revision: "", LastCommit: time.Time{}, DirtyBuild: false}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return vi
	}
	if info.Main.Version != "" {
		vi.Version = info.Main.Version
	}
	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			vi.Revision = kv.Value
		case "vcs.time":
			vi.LastCommit, _ = time.Parse(time.RFC3339, kv.Value)
		case "vcs.modified":
			vi.DirtyBuild = kv.Value == "true"
		}
	}

	return vi
}
