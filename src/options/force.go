package options

var force = false

// SetForce sets the force flag
//
func SetForce(toWhat bool) {
	force = toWhat
}

// IsForce gets the force flag
//
func IsForce() bool {
	return force
}
