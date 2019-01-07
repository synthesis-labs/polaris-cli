package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PolarisStackSpec defines the desired state of PolarisStack
type PolarisStackSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Nickname   string            `json:"nickname"`
	Template   string            `json:"template"`
	Parameters map[string]string `json:"parameters"`
	Finalizers []string          `json:"finalizers"`
}

// PolarisStackStatus defines the observed state of PolarisStack
type PolarisStackStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Status       string            `json:"status,omitempty"`
	Name         string            `json:"name,omitempty"`
	ID           string            `json:"id,omitempty"`
	Outputs      map[string]string `json:"output,omitempty"`
	StatusReason string            `json:"statusreason,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolarisStack is the Schema for the polarisstacks API
// +k8s:openapi-gen=true
// +genclient
type PolarisStack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolarisStackSpec   `json:"spec,omitempty"`
	Status PolarisStackStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolarisStackList contains a list of PolarisStack
type PolarisStackList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolarisStack `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PolarisStack{}, &PolarisStackList{})
}
