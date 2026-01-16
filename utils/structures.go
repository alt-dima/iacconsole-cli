package utils

type State struct {
	UnitName          string
	OrgName           string
	DimensionsFlags   []string
	UnitPath          string
	SharedModulesPath string
	InventoryPath     string
	UnitManifestPath  string
	ParsedDimensions  map[string]string
	CmdWorkTempDir    string
	UnitManifest      unitManifestStruct
	StateS3Path       string
	IacconsoleApiUrl  string
	Workspace         string
}

type unitManifestStruct struct {
	Dimensions []string
}

type IaCConsoleDBResponse struct {
	Error      string
	Dimensions []DimensionInIaCConsoleDB
}

type DimensionInIaCConsoleDB struct {
	ID        string
	WorkSpace string
	DimData   map[string]interface{}
}
