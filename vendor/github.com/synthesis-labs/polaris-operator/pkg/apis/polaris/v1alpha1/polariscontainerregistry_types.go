package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PolarisContainerRegistrySpec defines the desired state of PolarisContainerRegistry
type PolarisContainerRegistrySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Name string `json:"name"`
}

// PolarisContainerRegistryStatus defines the observed state of PolarisContainerRegistry
type PolarisContainerRegistryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolarisContainerRegistry is the Schema for the polariscontainerregistries API
// +k8s:openapi-gen=true
// +genclient
type PolarisContainerRegistry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolarisContainerRegistrySpec   `json:"spec,omitempty"`
	Stack  PolarisCloudformationStatus    `json:"stack,omitempty"`
	Status PolarisContainerRegistryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolarisContainerRegistryList contains a list of PolarisContainerRegistry
type PolarisContainerRegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolarisContainerRegistry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PolarisContainerRegistry{}, &PolarisContainerRegistryList{})
}
