package internal

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	PreferencesUpdatedEvent = iota
	ResourceChangedEvent    = iota
)

type PreferencesUpdated struct{}

func (ev PreferencesUpdated) Type() uint32 {
	return PreferencesUpdatedEvent
}

type ResourceChanged struct {
	*metav1.APIResource
}

func (ev ResourceChanged) Type() uint32 {
	return ResourceChangedEvent
}
