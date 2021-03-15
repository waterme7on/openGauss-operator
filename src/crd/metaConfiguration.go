package crd

import (
	types "k8s.io/apimachinery/pkg/types"
)

type TypeMetaApplyConfiguration struct {
	Kind       *string `json:"kind,omitempty"`
	APIVersion *string `json:"apiVersion,omitempty"`
}

type ObjectMetaApplyConfiguration struct {
	Name                       *string           `json:"name,omitempty"`
	GenerateName               *string           `json:"generateName,omitempty"`
	Namespace                  *string           `json:"namespace,omitempty"`
	SelfLink                   *string           `json:"selfLink,omitempty"`
	UID                        *types.UID        `json:"uid,omitempty"`
	ResourceVersion            *string           `json:"resourceVersion,omitempty"`
	Generation                 *int64            `json:"generation,omitempty"`
	DeletionGracePeriodSeconds *int64            `json:"deletionGracePeriodSeconds,omitempty"`
	Labels                     map[string]string `json:"labels,omitempty"`
	Annotations                map[string]string `json:"annotations,omitempty"`
	Finalizers                 []string          `json:"finalizers,omitempty"`
	ClusterName                *string           `json:"clusterName,omitempty"`
	// CreationTimestamp          *v1.Time          `json:"creationTimestamp,omitempty"`
	// DeletionTimestamp          *v1.Time          `json:"deletionTimestamp,omitempty"`
	// OwnerReferences            []OwnerReferenceApplyConfiguration `json:"ownerReferences,omitempty"`
}
