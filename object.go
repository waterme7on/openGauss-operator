/*
This files implements helpful utils to manage components of openGauss.
*/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	v1 "github.com/waterme7on/openGauss-controller/pkg/apis/opengausscontroller/v1"
	"github.com/waterme7on/openGauss-controller/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/intstr"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

// Identity represents the type of statefulsets
// Options: Master, Replicas
type Identity int

const (
	Master Identity = iota + 1
	Replicas
)

// NewPersistentVolumeCLaim returns pvc according to og's configuration
func NewPersistentVolumeClaim(og *v1.OpenGauss) *corev1.PersistentVolumeClaim {
	formatter := util.PersistentVolumeClaimFormatter(og.Name)
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: formatter.PersistentVolumeCLaimName(),
			Labels: map[string]string{
				"app": og.Name,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: *og.Spec.Resources,
		},
	}
	if og.Spec.StorageClassName != "" {
		pvc.Spec.StorageClassName = &og.Spec.StorageClassName
	}
	return pvc
}

// NewMasterStatefulsets returns master statefulset object
func NewMasterStatefulsets(og *v1.OpenGauss) (sts *appsv1.StatefulSet) {
	return NewStatefulsets(Master, og)
}

// NewReplicaStatefulsets returns replica statefulset object
func NewReplicaStatefulsets(og *v1.OpenGauss) (sts *appsv1.StatefulSet) {
	return NewStatefulsets(Replicas, og)
}

// NewStatefulsets returns a statefulset object according to id
func NewStatefulsets(id Identity, og *v1.OpenGauss) (res *appsv1.StatefulSet) {
	res = statefulsetTemplate()
	res.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(og, v1.SchemeGroupVersion.WithKind("OpenGauss")),
	}
	var formatter util.StatefulsetFormatterInterface
	switch id {
	case Master:
		formatter = util.Master(og.Name)
		res.Spec.Replicas = util.Int32Ptr(*og.Spec.OpenGauss.Master.Replicas)
		break
	case Replicas:
		formatter = util.Replica(og.Name)
		res.Spec.Replicas = util.Int32Ptr(*og.Spec.OpenGauss.Worker.Replicas)
		break
	default:
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

type configmap struct {
	ApiVersion string            `json:"apiVersion"`
	Data       map[string]string `json:"data"`
	Kind       string            `json:"kind"`
	Metadata   map[string]string `json:"metadata"`
}

func NewMasterConfigMap(og *v1.OpenGauss) (*unstructured.Unstructured, schema.GroupVersionResource) {
	return NewConfigMap(Master, og)
}

func NewReplicaConfigMap(og *v1.OpenGauss) (*unstructured.Unstructured, schema.GroupVersionResource) {
	return NewConfigMap(Replicas, og)
}

// NewConfigMap: return New Configmap as unstructured.Unstructured and configMap Schema
// modify replConnInfo of configmap data["postgresql.conf"] according to the id of og
func NewConfigMap(id Identity, og *v1.OpenGauss) (*unstructured.Unstructured, schema.GroupVersionResource) {
	unstructuredMap := loadConfigMapTemplate()
	var replConnInfo string
	var formatter util.StatefulsetFormatterInterface
	if id == Master {
		formatter = util.Master(og.Name)
	} else {
		formatter = util.Replica(og.Name)
	}
	replConnInfo = "\n" + formatter.ReplConnInfo() + "\n"
	configMap := &unstructured.Unstructured{Object: unstructuredMap}

	// transform configMap from unstructured to []bytes
	s, _ := configMap.MarshalJSON()
	configStruct := configmap{}
	// transform []bytes to struct configmap to modify Data["postgresql.conf"]
	json.Unmarshal(s, &configStruct)
	// add replConnInfo according to og's id
	configStruct.Data["postgresql.conf"] += replConnInfo
	s, _ = json.Marshal(configStruct)
	configMap.UnmarshalJSON(s)

	configMapRes := schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}
	configMap.SetName(formatter.ConfigMapName())
	return configMap, configMapRes
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
			Replicas: util.Int32Ptr(1),
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

					TerminationGracePeriodSeconds: util.Int64Ptr(10),
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
								Privileged: util.BoolPtr(true),
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
						StorageClassName: util.StrPtr("csi-lvm"),
					},
				},
			},
		},
	}
	return template
}

// load configmap file from /config/config.yaml
func loadConfigMapTemplate() map[string]interface{} {

	// configMap := corev1.ConfigMap{}
	fileBytes, err := ioutil.ReadFile("configs/config.yaml")
	if err != nil {
		fmt.Println("[error]", err)
	}
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(fileBytes), 100)
	var rawObj runtime.RawExtension
	if err = decoder.Decode(&rawObj); err != nil {
		fmt.Println("[error]", err)
	}
	obj, _, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	// log.Printf("[log] map type:%T\n", unstructuredMap["data"])
	// log.Println("[log] map: ", unstructuredMap["data"])
	return unstructuredMap
}
