package crd

import "encoding/json"

// OpenGauss runtime object
type OpenGaussConfiguration struct {
	// apiVersion                       string `json:"apiVersion"`
	// kind                             string `json:"kind"`
	TypeMeata  *TypeMetaApplyConfiguration   `json:",inline"`
	ObjectMeta *ObjectMetaApplyConfiguration `json:"metadata,omitempty"`
	Spec       *OpenGaussSpecConfiguration   `json:"spec,omitempty"`
	Status     *OpenGaussStatusConfiguration `json:"status,omitempty"`
}

type OpenGaussSpecConfiguration struct {
	Image           *string                        `json:"image"`
	ImagePullPolicy *string                        `json:"imagePullPolicy"`
	OpenGauss       *OpenGaussClusterConfiguration `json:"opengauss"`
}

// Selector        *v1.LabelSelectorApplyConfiguration `json:"selector,omitempty"`

type OpenGaussClusterConfiguration struct {
	Master   *int32 `json:"master"`
	Replicas *int32 `json:"replicas"`
}

type OpenGaussStatusConfiguration struct {
}

type OpenGaussListConfiguration struct {
	TypeMeata  *TypeMetaApplyConfiguration   `json:",inline"`
	ObjectMeta *ObjectMetaApplyConfiguration `json:"metadata,omitempty"`
	OgList     []OpenGaussConfiguration      `json:"items"`
}

type Configuration interface {
	String() string
}

func (config *OpenGaussConfiguration) String() string {
	ret, _ := json.Marshal(config)
	return string(ret)
}

func (config *OpenGaussListConfiguration) String() string {
	ret, _ := json.Marshal(config)
	return string(ret)
}
