package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Rule describes Istio rule
type Rule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              *RuleSpec `json:"spec"`
}

// RuleSpec is the spec for Rule resource
type RuleSpec struct {
	Match   string    `json:"match,omitempty"`
	Actions []*Action `json:"actions,omitempty"`
}

// Action describes action for the rule
type Action struct {
	Handler   string   `json:"handler,omitempty"`
	Instances []string `json:"instances,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RuleList is a list of Rule resources
type RuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rule `json:"items"`
}
