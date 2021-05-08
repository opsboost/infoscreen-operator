package v1

import (
	kmapi "kmodules.xyz/client-go/api/v1"
)

// +kubebuilder:validation:Enum=Provisioning;Ready;NotReady
type ScreenPhase string

const (
	// used for Clusters that are currently provisioning
	ScreenStatusProvisioning ScreenPhase = "Provisioning"
	// used for Clusters that are currently Active or Ready
	ScreenStatusReady ScreenPhase = "Ready"
	// used for Clusters without child resources
	ScreenStatusNotReady ScreenPhase = "NotReady"
)

// ClusterStatus defines the observed state of Cluster
type ScreenStatus struct {
	// Specifies the current phase of the database
	// +optional
	Phase ScreenPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=ScreenPhase"`
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,2,opt,name=observedGeneration"`
	// Conditions applied to the database, such as approval or denial.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
}
