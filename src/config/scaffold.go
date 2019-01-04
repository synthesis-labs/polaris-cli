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
	Name      string
	LocalPath string
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
