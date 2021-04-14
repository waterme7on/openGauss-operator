/*
	this file defines formatter for names and attributes used in OpenGauss-controller
*/
package util

type FormatterInterface interface {
	StatefulSetName() string
	ServiceName() string
	ReplConnInfo() string
	ConfigMapName() string
}

func Master(configName string) FormatterInterface {
	return &MasterFormatter{configName: configName}
}
func Replica(configName string) FormatterInterface {
	return &ReplicaFormatter{configName: configName}
}

type MasterFormatter struct {
	configName string
}

func (formatter *MasterFormatter) StatefulSetName() string {
	return formatter.configName + "-masters"
}

func (formatter *MasterFormatter) ServiceName() string {
	return formatter.configName + "-master-service"
}

func (formatter *MasterFormatter) ReplConnInfo() string {
	return formatter.configName + "-masters"
}

func (formatter *MasterFormatter) ConfigMapName() string {
	return formatter.configName + "-master-config"
}

type ReplicaFormatter struct {
	configName string
}

func (formatter *ReplicaFormatter) StatefulSetName() string {
	return formatter.configName + "-replicas"
}

func (formatter *ReplicaFormatter) ServiceName() string {
	return formatter.configName + "-replicas-service"
}

func (formatter *ReplicaFormatter) ReplConnInfo() string {
	return formatter.configName + "-replicas"
}

func (formatter *ReplicaFormatter) ConfigMapName() string {
	return formatter.configName + "-replicas-config"
}
