package config

// PolarisScaffoldParameter holds a parameter
//
type PolarisScaffoldParameter struct {
	Name    string
	Default string
}

// PolarisScaffoldSpec defines a scaffold spec
//
type PolarisScaffoldSpec struct {
	Description string
	Help        string
	Parameters  []PolarisScaffoldParameter
}

// PolarisScaffold defines a Scaffold
//
type PolarisScaffold struct {
	Spec      PolarisScaffoldSpec
	LocalPath string
}

// PolarisScaffoldProject defines a scaffold project
//
type PolarisScaffoldProject struct {
	Name       string
	Parameters map[string]string
	Scaffold   *PolarisScaffold
}
