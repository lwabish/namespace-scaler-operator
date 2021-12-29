/*
Copyright 2021.

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

// NSScalerSpec defines the desired state of NSScaler
type NSScalerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// 控制器作用于哪些命名空间？用前缀filter出来。
	// 对于不包含在prefix内的命名空间不会干预。
	ScopePrefix string `json:"scope_prefix,omitempty"`
	// scope内命名空间后缀有哪些是会用到的，这些不会被scale成0
	ActiveNamespaceSuffixes []string `json:"active_namespace_suffixes,omitempty"`
}

// NSScalerStatus defines the observed state of NSScaler
type NSScalerStatus struct {
	// 被prefix包含，且不活跃的命名空间内没有一个pod时，Done为true
	Done bool `json:"done,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// NSScaler is the Schema for the nsscalers API
type NSScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NSScalerSpec   `json:"spec,omitempty"`
	Status NSScalerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NSScalerList contains a list of NSScaler
type NSScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NSScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NSScaler{}, &NSScalerList{})
}
