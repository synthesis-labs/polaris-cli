package v1alpha1

// PolarisCloudformationStatus defines the observed state of the particular Polaris Resource
type PolarisCloudformationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	StackCreationAttempted bool   `json:"stackCreationAttempted,omitempty"`
	StackResponse          string `json:"stackResponse,omitempty"`
	StackError             string `json:"stackError,omitempty"`
	StackName              string `json:"stackName,omitempty"`
}
