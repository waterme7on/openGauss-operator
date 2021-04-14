/*
	This files implements helpful utils to manage components of openGauss.
*/
package v1

const (
	readyStatus   string = "READY"
	unreadyStatus        = "NOT-READY"
)

// IsReady check if opengauss is ready
func (og *OpenGauss) IsReady() bool {
	return og.Status.OpenGaussStatus == readyStatus
}

// IsMasterDeployed check if opengauss's master is deployed
func (og *OpenGauss) IsMasterDeployed() bool {
	return og.Status.ReadyMaster == og.Spec.OpenGauss.Master
}

// IsReplicaDeployed check if opengauss's replicas is deployed
func (og *OpenGauss) IsReplicaDeployed() bool {
	return og.Status.ReadyReplicas == og.Spec.OpenGauss.Replicas
}
