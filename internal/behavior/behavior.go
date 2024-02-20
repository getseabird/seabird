package behavior

import (
	"github.com/getseabird/seabird/api"
	"github.com/imkira/go-observer/v2"
)

type Behavior struct {
	Preferences observer.Property[api.Preferences]
}

func NewBehavior() (*Behavior, error) {
	prefs, err := api.LoadPreferences()
	if err != nil {
		return nil, err
	}
	prefs.Defaults()

	return &Behavior{
		Preferences: observer.NewProperty(*prefs),
	}, nil
}
