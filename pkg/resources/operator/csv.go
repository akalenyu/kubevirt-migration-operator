package operator

import (
	csvv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// ClusterServiceVersionData - Data arguments used to create CDI's CSV manifest
type ClusterServiceVersionData struct {
	CsvVersion         string
	ReplacesCsvVersion string
	Namespace          string
	ImagePullPolicy    string
	ImagePullSecrets   []corev1.LocalObjectReference
	IconBase64         string
	Verbosity          string

	OperatorVersion string

	ControllerImage string
	OperatorImage   string
}

// NewMigControllerCrd - provides MigController CRD
func NewMigControllerCrd() *extv1.CustomResourceDefinition {
	return createMigControllerCRD()
}

// NewClusterServiceVersion - generates CSV for CDI
func NewClusterServiceVersion(data *ClusterServiceVersionData) (*csvv1.ClusterServiceVersion, error) {
	return createClusterServiceVersion(data)
}
