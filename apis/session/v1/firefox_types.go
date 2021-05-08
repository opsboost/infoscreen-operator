/*
Copyright 2021 OpsBoost Crew <info@opsboost.dev>.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// FirefoxSpec defines the desired state of Firefox
type FirefoxSpec struct {
	SessionSpec `json:",inline"`
	Url         string           `json:"url"`
	Target      string           `json:"target,omitempty"`
	Credentials *CredentialsSpec `json:"credentials,omitempty"`
	Destination string           `json:"destination,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=firefoxes

// Firefox is the Schema for the firefoxes API
type Firefox struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FirefoxSpec  `json:"spec,omitempty"`
	Status ScreenStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FirefoxList contains a list of Firefox
type FirefoxList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Firefox `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Firefox{}, &FirefoxList{})
}
