package config

// PolarisScaffoldParameter holds a parameter
//
type PolarisScaffoldParameter struct {
	Name    string
	Default string
}

// PolarisScaffoldComponent defines a component within a project spec
//
type PolarisScaffoldComponent struct {
	Name        string
	Description string
	Help        string
	Parameters  []PolarisScaffoldParameter
}

// PolarisScaffoldSpec defines a scaffold spec
//
type PolarisScaffoldSpec struct {
	Description string
	Help        string
	Parameters  []PolarisScaffoldParameter
	Components  []PolarisScaffoldComponent
}

// PolarisScaffold defines a Scaffold
//
type PolarisScaffold struct {
	Spec      PolarisScaffoldSpec
	Name      string
	LocalPath string
}

// PolarisScaffoldProject defines a scaffold project
// (deprecated)
type PolarisScaffoldProject struct {
	Name        string
	Application string
	Parameters  map[string]string
	Scaffold    *PolarisScaffold
}

// PolarisScaffoldApplication for generating an Application
// (deprecated)
type PolarisScaffoldApplication struct {
	Application string
	Parameters  map[string]string
	Scaffold    string
}

// PolarisScaffoldComponent for generating a Component within an Application
// (deprecated)
type notPolarisScaffoldComponent struct {
	Component       string
	Application     string
	ApplicationSpec *PolarisScaffoldApplication
	Parameters      map[string]string
}

// PolarisProject defines the structure for ./polaris-project.yaml within a local project
//
type PolarisProject struct {
	Project    string
	Parameters map[string]string
	Scaffold   string
}

// PolarisComponent for generating a Component within a project
//
type PolarisComponent struct {
	Project           string
	Component         string
	Parameters        map[string]string
	ProjectParameters map[string]string
	ProjectScaffold   string
	ComponentScaffold string
}
