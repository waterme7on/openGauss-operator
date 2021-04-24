/*
This files implements helpful utils to manage components of openGauss.
*/
package main

import (
	v1 "github.com/waterme7on/openGauss-controller/pkg/apis/opengausscontroller/v1"
	"github.com/waterme7on/openGauss-controller/util"
	autoscaling "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	scalerv1 "github.com/waterme7on/openGauss-controller/pkg/apis/autoscaler/v1"
)

func NewHorizontalPodAutoscaler(og *v1.OpenGauss, autoscaler *scalerv1.AutoScaler, id Identity) *autoscaling.HorizontalPodAutoscaler {
	hpa := &autoscaling.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v2beta2",
		},
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(og, v1.SchemeGroupVersion.WithKind("OpenGauss")),
			},
		},
	}
	var formatter util.StatefulsetFormatterInterface
	if id == Master {
		formatter = util.Master(og)
		hpa.Spec = *autoscaler.Spec.Master.Spec
	} else {
		formatter = util.Replica(og)
		hpa.Spec = *autoscaler.Spec.Worker.Spec
	}
	hpa.Name = formatter.StatefulSetName()
	return hpa
}
