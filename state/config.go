package state

import (
	"encoding/json"
	"errors"
	"os"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

const Path = "/tmp/kubegtk.json"

type Config struct {
	Navigation struct {
		Favourites []schema.GroupVersionKind `json:"favourites,omitempty"`
	} `json:"navigation,omitempty"`
}

func LoadConfig() (*Config, error) {
	if _, err := os.Stat(Path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			var config Config
			config.Defaults()
			return &config, nil
		}
		return nil, err
	}

	f, err := os.Open(Path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}

	config.Defaults()

	return &config, nil
}

func (c *Config) Defaults() {
	if len(c.Navigation.Favourites) == 0 {
		c.Navigation.Favourites = []schema.GroupVersionKind{
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

func (c *Config) Save() error {
	f, err := os.OpenFile(Path, os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(c)
}
