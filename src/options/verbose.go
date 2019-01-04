package options

var verbose = false

// SetVerbose sets the verbose flag
//
func SetVerbose(toWhat bool) {
	verbose = toWhat
}

// IsVerbose gets the verbose flag
//
func IsVerbose() bool {
	return verbose
}
