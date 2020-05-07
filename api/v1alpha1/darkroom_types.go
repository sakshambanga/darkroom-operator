/*


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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Kind string

const (
	WEB_FOLDER Kind = "WebFolder"
)

type Source struct {
	Kind    Kind   `json:"kind"`
	BaseURL string `json:"baseUrl"`
}

// DarkroomSpec defines the desired state of Darkroom
type DarkroomSpec struct {
	Version string `json:"version"`
	Source  Source `json:"source"`
}

// DarkroomStatus defines the observed state of Darkroom
type DarkroomStatus struct {
	URL string `json:"url"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.version",name=VERSION,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.source.kind",name=KIND,type=string

// Darkroom is the Schema for the darkrooms API
type Darkroom struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DarkroomSpec   `json:"spec,omitempty"`
	Status DarkroomStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DarkroomList contains a list of Darkroom
type DarkroomList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Darkroom `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Darkroom{}, &DarkroomList{})
}
