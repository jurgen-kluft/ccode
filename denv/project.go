package denv

type Project struct {
	Name            string
	Author          string
	GUID            string
	Path            string
	ProjectID       string
	HdrGlobPaths    []string
	HdrVirtualPaths []string
	HdrFiles        []string
	SrcGlobPaths    []string
	SrcVirtualPaths []string
	SrcFiles        []string
	Platforms       []string
	Configs         []string
	Dependencies    []Project
}
