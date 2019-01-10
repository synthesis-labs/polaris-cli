package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PolarisBuildStepBuildSpec defines the desired state of each build spec
type PolarisBuildStepBuildSpec struct {
	Name               string `json:"name"`
	DockerfileLocation string `json:"dockerfilelocation"`
	ContainerRegistry  string `json:"containerregistry"`
	Tag                string `json:"tag"`
}

// PolarisBuildStepSpec defines the desired state of PolarisBuildStep
type PolarisBuildStepSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Pipeline string                      `json:"pipeline"`
	Builds   []PolarisBuildStepBuildSpec `json:"builds"`
}

// PolarisBuildStepStatus defines the observed state of PolarisBuildStep
type PolarisBuildStepStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Status string `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolarisBuildStep is the Schema for the polarisbuildsteps API
// +k8s:openapi-gen=true
// +genclient
type PolarisBuildStep struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolarisBuildStepSpec   `json:"spec,omitempty"`
	Status PolarisBuildStepStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolarisBuildStepList contains a list of PolarisBuildStep
type PolarisBuildStepList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolarisBuildStep `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PolarisBuildStep{}, &PolarisBuildStepList{})
}
