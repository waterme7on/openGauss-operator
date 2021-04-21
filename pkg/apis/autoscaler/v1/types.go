package v1

import (
	autoscaling "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenGauss is a top-level type
type AutoScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              *AutoScalerSpec `json:"spec,omitempty"`
}

type AutoScalerSpec struct {
	Cluster ClusterInfo `json:"cluster,omitempty"`
	Master  *Component  `json:"master,omitempty"`
	Worker  *Component  `json:"worker,omitempty"`
}

type Component struct {
	Spec *autoscaling.HorizontalPodAutoscalerSpec `json:"spec,omitempty"`
	// Status autoscaling.HorizontalPodAutoscalerStatus `json:"status,omitempty"`
}

type ClusterInfo struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// no client needed for list as it's been created in above
type AutoScalerList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutoScaler `json:"items"`
}
