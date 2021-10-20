package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenGauss is a top-level type
type OpenGauss struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +optional
	Status *OpenGaussStatus `json:"status,omitempty"`
	Spec   *OpenGaussSpec   `json:"spec,omitempty"`
}

type OpenGaussSpec struct {
	Image            string                         `json:"image"`
	ImagePullPolicy  string                         `json:"imagePullPolicy"`
	OpenGauss        *OpenGaussClusterConfiguration `json:"opengauss"`
	Resources        *corev1.ResourceRequirements   `json:"resources,omitempty"`
	StorageClassName string                         `json:"storageClassName,omitempty"`
}

// Define OpenGauss's needs for master and replicas
type OpenGaussClusterConfiguration struct {
	Master *OpenGaussStatefulSet `json:"master"` // Master Configuration
	Worker *OpenGaussStatefulSet `json:"worker"` // Replicas Configuration
}

type OpenGaussStatefulSet struct {
	Replicas  *int32                       `json:"replicas"`
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

// OpenGauss Cluster's status
type OpenGaussStatus struct {
	OpenGaussStatus           string   `json:"opengaussStatus"`                     // OpenGauss if ready or not
	ReadyMaster               string   `json:"readyMaster,omitempty"`               // Ready Master number
	ReadyReplicas             string   `json:"readyReplicas,omitempty"`             // Ready Replicas number
	MasterStatefulset         string   `json:"masterStatefulset,omitempty"`         // name of master statefulset
	ReplicasStatefulset       string   `json:"replicasStatefulset,omitempty"`       // name of replicas statefulset
	PersistentVolumeClaimName string   `json:"persistentVolumeClaimName,omitempty"` // name of pvc
	MasterIPs                 []string `json:"masterIPs,omitempty"`                 // master ips
	ReplicasIps               []string `json:"replicasIPs,omitempty"`               // replicas ips
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// no client needed for list as it's been created in above
type OpenGaussList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenGauss `json:"items"`
}
