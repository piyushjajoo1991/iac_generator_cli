package version

// Version information
var (
	// Version is the current version of the application
	Version = "0.1.0"
	
	// CommitHash is the git commit hash at build time
	CommitHash = "unknown"
	
	// BuildDate is the date when the binary was built
	BuildDate = "unknown"
)

// GetVersionInfo returns a formatted version string
func GetVersionInfo() string {
	return Version
}

// GetFullVersionInfo returns detailed version information
func GetFullVersionInfo() string {
	return "Version: " + Version + " Commit: " + CommitHash + " BuildDate: " + BuildDate
}