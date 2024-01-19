package state

import (
	"encoding/json"
	"errors"
	"os"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/clientcmd/api"
)

const Path = "/tmp/kubegtk.json"

type Preferences struct {
	Clusters []ClusterPreferences
}

type ClusterPreferences struct {
	Name     string
	AuthInfo api.AuthInfo

	Navigation struct {
		Favourites []schema.GroupVersionKind `json:"favourites,omitempty"`
	} `json:"navigation,omitempty"`
}

func LoadPreferences() (*Preferences, error) {
	if _, err := os.Stat(Path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			var prefs Preferences
			prefs.Defaults()
			return &prefs, nil
		}
		return nil, err
	}

	f, err := os.Open(Path)
	if err != nil {
		return nil, err
	}

	var prefs Preferences
	if err := json.NewDecoder(f).Decode(&prefs); err != nil {
		return nil, err
	}

	prefs.Defaults()

	return &prefs, nil
}

func (c *Preferences) Defaults() {
	for i, cluster := range c.Clusters {
		if len(cluster.Navigation.Favourites) == 0 {
			c.Clusters[i].Navigation.Favourites = []schema.GroupVersionKind{
				{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
				{
					Group:   "apps",
					Version: "v1",
					Kind:    "StatefulSet",
				},
				{
					Version: "v1",
					Kind:    "Pod",
				},
				{
					Version: "v1",
					Kind:    "ConfigMap",
				},
				{
					Version: "v1",
					Kind:    "Secret",
				},
				{
					Version: "v1",
					Kind:    "Namespace",
				},
			}
		}
	}
}

func (c *Preferences) Save() error {
	f, err := os.OpenFile(Path, os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(c)
}
