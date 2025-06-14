/*
Copyright 2023.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DvlsSecretSpec defines the desired state of DvlsSecret
type DvlsSecretSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	EntryID string `json:"entryId"` // entry id on dvls
	VaultID string `json:"vaultId"` // vault id on dvls
}

// DvlsSecretStatus defines the observed state of DvlsSecret
type DvlsSecretStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions        []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
	EntryModifiedDate metav1.Time        `json:"entryModifiedDate"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DvlsSecret is the Schema for the dvlssecrets API
type DvlsSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DvlsSecretSpec   `json:"spec,omitempty"`
	Status DvlsSecretStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DvlsSecretList contains a list of DvlsSecret
type DvlsSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DvlsSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DvlsSecret{}, &DvlsSecretList{})
}
