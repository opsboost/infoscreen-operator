package v1

import (
	corev1 "k8s.io/api/core/v1"
)

type CredentialsSpec struct {
	SecretRef *corev1.SecretEnvSource `json:"secretRef,omitempty"`
}

type SessionResolutionSpec struct {
	Width  uint16 `json:"width,omitempty"`
	Height uint16 `json:"height,omitempty"`
}

type SessionClusterRefSpec struct {
	Name string `json:"name,omitempty"`
}

type SessionSpec struct {
	Resolution   *SessionResolutionSpec `json:"resolution,omitempty"`
	ClusterRef   *SessionClusterRefSpec `json:"clusterRef,omitempty"`
	BitsPerPixel uint8                  `json:"bitsPerPixel,omitempty"`
}
