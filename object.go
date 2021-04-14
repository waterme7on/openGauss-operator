/*
This files implements helpful utils to manage components of openGauss.
*/
package main

import (
	"fmt"

	v1 "github.com/waterme7on/openGauss-controller/pkg/apis/opengausscontroller/v1"
	"github.com/waterme7on/openGauss-controller/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Identity represents the type of statefulsets
// Options: Master, Replicas
type Identity int

const (
	Master Identity = iota + 1
	Replicas
)

// MasterStatefulsets returns master statefulset object
func MasterStatefulsets(og *v1.OpenGauss) (sts *appsv1.StatefulSet, err error) {
	return NewStatefulsets(Master, og)
}

// ReplicaStatefulsets returns replica statefulset object
func ReplicaStatefulsets(og *v1.OpenGauss) (sts *appsv1.StatefulSet, err error) {
	return NewStatefulsets(Replicas, og)
}

// NewStatefulsets returns a statefulset object according to id
func NewStatefulsets(id Identity, og *v1.OpenGauss) (res *appsv1.StatefulSet, err error) {
	res = statefulsetTemplate()
	var formatter util.FormatterInterface
	switch id {
	case Master:
		formatter = util.Master(og.Name)
		res.Spec.Replicas = int32Ptr(og.Spec.OpenGauss.Master)
		break
	case Replicas:
		formatter = util.Replica(og.Name)
		res.Spec.Replicas = int32Ptr(og.Spec.OpenGauss.Replicas)
		break
	default:
		err = fmt.Errorf("wrong identity")
		return
	}
	res.Name = formatter.StatefulSetName()
	res.Spec.Selector.MatchLabels["app"] = res.Name
	res.Spec.Template.ObjectMeta.Labels["app"] = res.Name
	res.Spec.Template.Spec.Containers[0].Name = res.Name
	res.Spec.Template.Spec.Containers[0].Env[0].Value = formatter.ReplConnInfo()
	res.Spec.Template.Spec.Volumes[0].ConfigMap.Name = formatter.ConfigMapName()
	return
}

// serviceTemplate returns a service template of type corev1.Service
func serviceTemplate() *corev1.Service {
	template := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "opengauss-service",
			Labels: map[string]string{
				"app": "opengauss",
			},
		},
		Spec: corev1.ServiceSpec{
			Type: "NodePort",
			Ports: []corev1.ServicePort{
				{
					Name:     "opengauss-port",
					Port:     5432,
					Protocol: corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						IntVal: 5432,
					},
					NodePort: 31001,
				},
			},
			Selector: map[string]string{
				"app": "opengauss",
			},
		},
	}
	return template
}

// statefulsetTemplate returns a statefulset template of opengauss
func statefulsetTemplate() *appsv1.StatefulSet {
	template := &appsv1.StatefulSet{
		// TypeMeta: metav1.TypeMeta{
		// 	Kind:       "StatefulSet",
		// 	APIVersion: "apps/v1",
		// },
		ObjectMeta: metav1.ObjectMeta{
			Name: "opengauss-statefulset",
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "opengauss",
				},
			},
			ServiceName: "",
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "opengauss",
					},
				},
				Spec: corev1.PodSpec{

					TerminationGracePeriodSeconds: int64Ptr(10),
					Containers: []corev1.Container{
						{
							Name:  "opengauss",
							Image: "opengauss:debug",
							Args: []string{
								"-M",
								"primary",
								"-c",
								"config_file=/etc/opengauss/postgresql.conf",
								"-c",
								"hba_file=/etc/opengauss/pg_hba.conf",
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: boolPtr(true),
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "opengauss",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 5432,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "REPL_CONN_INFO",
									// missing remotehost=... and localhost=...
									// master: set localhost to the first master pod name and remote host to first replica pod name
									// replica: set localhost to $POD_IP and remote host to first master pod name
									Value: "replconninfo1 = 'localport=5434 localservice=5432 remoteport=5434 remoteservice=5432'\n",
								},
								{
									Name:  "GS_PORT",
									Value: "5432",
								},
								{
									Name: "POD_IP",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "status.podIP",
										},
									},
								},
								{
									Name: "NODE_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.nodeName",
										},
									},
								},
								{
									Name:  "GS_PASSWORD",
									Value: "Enmo@123",
								},
							},
							ImagePullPolicy: corev1.PullNever,
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/var/lib/opengauss",
									Name:      "opengauss-pvc",
									// SubPath:   "",
								},
								{
									MountPath: "/etc/opengauss/postgresql.conf",
									Name:      "opengauss-config",
									SubPath:   "postgresql.conf",
								},
								{
									MountPath: "/etc/opengauss/pg_hba.conf",
									Name:      "opengauss-config",
									SubPath:   "pg_hba.conf",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							"opengauss-config",
							// defined by files in /configs
							corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									corev1.LocalObjectReference{
										Name: "opengauss-configmap",
									},
									[]corev1.KeyToPath{},
									nil,
									nil,
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "opengauss-pvc",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"storage": resource.MustParse("500Mi"),
							},
						},
						StorageClassName: strPtr("csi-lvm"),
					},
				},
			},
		},
	}
	return template
}

func int64Ptr(i int64) *int64 { return &i }

func int32Ptr(i int32) *int32 { return &i }

func strPtr(s string) *string { return &s }

func boolPtr(s bool) *bool { return &s }
