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

func YamlToJson(data []byte) ([]byte, error) {
	var o map[interface{}]interface{}
	if err := yaml.Unmarshal(data, &o); err != nil {
		return nil, err
	}
	ret, err := json.Marshal(convertStringKeys(o))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func convertStringKeys(i any) any {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convertStringKeys(v)
		}
		return m2
	case map[string]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k] = convertStringKeys(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convertStringKeys(v)
		}
	}
	return i
}
