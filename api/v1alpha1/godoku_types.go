package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GodokuSpec struct {
	Namespace string                 `json:"namespace,omitempty"`
	Name      string                 `json:"name"`
	Replicas  int32                  `json:"replicas,omitempty"`
	Image     string                 `json:"image"`
	Ports     []corev1.ContainerPort `json:"ports"`
	EnvFrom   []corev1.EnvFromSource `json:"envFrom"`
}

type GodokuStatus struct {
	Pods              []string `json:"pods"`
	AvailableReplicas int      `json:"availableReplicas"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Godoku struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GodokuSpec   `json:"spec,omitempty"`
	Status GodokuStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type GodokuList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Godoku `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Godoku{}, &GodokuList{})
}
