/*
	this file defines formatter for names and attributes used in OpenGauss-controller
*/
package util

import (
	"fmt"
	v1 "github.com/waterme7on/openGauss-controller/pkg/apis/opengausscontroller/v1"
)

type PersistentVolumeClaimFormatterInterface interface {
	PersistentVolumeCLaimName() string
}

func PersistentVolumeClaimFormatter(og *v1.OpenGauss) *persistentVolumeClaimFormatter {
	return &persistentVolumeClaimFormatter{
		OpenGauss: og,
	}
}

type persistentVolumeClaimFormatter struct {
	OpenGauss *v1.OpenGauss
}

func (formatter *persistentVolumeClaimFormatter) PersistentVolumeCLaimName() string {
	return formatter.OpenGauss.Name + "-pvc"
}

type StatefulsetFormatterInterface interface {
	StatefulSetName() string
	ServiceName() string
	ReplConnInfo() string
	ConfigMapName() string
}

func Master(og *v1.OpenGauss) StatefulsetFormatterInterface {
	return &MasterFormatter{OpenGauss: og}
}
func Replica(og *v1.OpenGauss) StatefulsetFormatterInterface {
	return &ReplicaFormatter{OpenGauss: og}
}

type MasterFormatter struct {
	OpenGauss *v1.OpenGauss
}

func (formatter *MasterFormatter) StatefulSetName() string {
	return formatter.OpenGauss.Name + "-masters"
}

func (formatter *MasterFormatter) ServiceName() string {
	return formatter.OpenGauss.Name + "-master-service"
}

func (formatter *MasterFormatter) ReplConnInfo() string {
	replica := Replica(formatter.OpenGauss)
	replicaStatefulsetName := replica.StatefulSetName()
	// workerSize := int(math.Max(float64(*formatter.OpenGauss.Spec.OpenGauss.Worker.Replicas), 1))
	replInfo := ""
	for i := 0; i < 1; i++ {
		replInfo += fmt.Sprintf("replconninfo%d='localhost=%s-0 remotehost=%s-%d", i+1, formatter.StatefulSetName(), replicaStatefulsetName, i)
		replInfo += " localport=5434 localservice=5432 remoteport=5434 remoteservice=5432'\n"
	}
	return replInfo
}

func (formatter *MasterFormatter) ConfigMapName() string {
	return formatter.OpenGauss.Name + "-master-config"

}

type ReplicaFormatter struct {
	OpenGauss *v1.OpenGauss
}

func (formatter *ReplicaFormatter) StatefulSetName() string {
	return formatter.OpenGauss.Name + "-replicas"
}

func (formatter *ReplicaFormatter) ServiceName() string {
	return formatter.OpenGauss.Name + "-replicas-service"
}

func (formatter *ReplicaFormatter) ReplConnInfo() string {
	master := Master(formatter.OpenGauss)
	masterStatefulsetName := master.StatefulSetName()
	replInfo := ""
	replInfo += fmt.Sprintf("replconninfo1='localhost=127.0.0.1 remotehost=%s-0", masterStatefulsetName)
	replInfo += " localport=5434 localservice=5432 remoteport=5434 remoteservice=5432'\n"
	// workerSize := int(math.Max(float64(*formatter.OpenGauss.Spec.OpenGauss.Worker.Replicas), 1))
	// replInfo := ""
	// for i := 0; i < workerSize; i++ {
	// 	replInfo += fmt.Sprintf("replconninfo%d='localhost=%s-%d remotehost=%s-0", i+1, formatter.StatefulSetName(), i, masterStatefulsetName)
	// 	replInfo += " localport=5434 localservice=5432 remoteport=5434 remoteservice=5432'\n"
	// }
	return replInfo
}

func (formatter *ReplicaFormatter) ConfigMapName() string {
	return formatter.OpenGauss.Name + "-replicas-config"
}
