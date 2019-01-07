package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PolarisSourceRepositorySpec defines the desired state of PolarisSourceRepository
type PolarisSourceRepositorySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Name string `json:"name"`
}

// PolarisSourceRepositoryStatus defines the observed state of PolarisSourceRepository
type PolarisSourceRepositoryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolarisSourceRepository is the Schema for the polarissourcerepositories API
// +k8s:openapi-gen=true
// +genclient
type PolarisSourceRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolarisSourceRepositorySpec   `json:"spec,omitempty"`
	Stack  PolarisCloudformationStatus   `json:"stack,omitempty"`
	Status PolarisSourceRepositoryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolarisSourceRepositoryList contains a list of PolarisSourceRepository
type PolarisSourceRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolarisSourceRepository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PolarisSourceRepository{}, &PolarisSourceRepositoryList{})
}
