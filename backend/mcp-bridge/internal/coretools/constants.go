package coretools

// Tool output thresholds and limits
const (
	// MetadataThresholdLines — files above this return metadata-first response
	MetadataThresholdLines = 300

	// OutputLimitChars — smart truncation threshold for tool output
	OutputLimitChars = 30000

	// MaxLineTruncation — lines longer than this are truncated
	MaxLineTruncation = 2000

	// MinifiedLineThreshold — lines longer than this indicate minified content
	MinifiedLineThreshold = 10000

	// LargeFileReadbackLines — for write/edit read-back, show first N + last M lines
	LargeFileReadbackHead = 100
	LargeFileReadbackTail = 50
	LargeFileReadbackThreshold = 500
)

// DefaultIgnoreDirs are directories excluded from find/grep operations.
// These are common build artifacts, dependency caches, and VCS internals.
var DefaultIgnoreDirs = []string{
	"node_modules",
	".git",
	"dist",
	"build",
	"__pycache__",
	".pycache",
	"vendor",
	".venv",
	"venv",
	".next",
	".nuxt",
	".svelte-kit",
	".turbo",
	".cache",
	".parcel-cache",
	"coverage",
	".nyc_output",
	".tox",
	".eggs",
	"target",       // Rust/Java
	"out",          // Generic build output
	".terraform",
	".angular",
	"bower_components",
	".idea",
	".vscode",
}

// DangerousCommands are shell patterns that should be blocked
var DangerousCommands = []string{
	"rm -rf /",
	"rm -rf /*",
	"rm -rf ~",
	"mkfs.",
	"dd if=/dev/zero",
	"dd if=/dev/random",
	"dd if=/dev/urandom",
	":(){ :|:& };:", // fork bomb
	"chmod -R 000",
	"chmod -R 777 /",
	"> /dev/sda",
	"mv / ",
	"wget -O- | sh",
	"curl | sh",
	"curl | bash",
}
