package util

import (
	"encoding/json"

	"gopkg.in/yaml.v2"
)

func JsonToYaml(data []byte) ([]byte, error) {
	var o any
	if err := json.Unmarshal(data, &o); err != nil {
		return nil, err
	}
	ret, err := yaml.Marshal(o)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
