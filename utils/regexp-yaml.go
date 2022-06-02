package utils

import (
	"gopkg.in/yaml.v3"
	"regexp"
)

type RegexpYaml struct {
	*regexp.Regexp
}

func (r *RegexpYaml) UnmarshalYAML(value *yaml.Node) error {
	regex, err := regexp.Compile(value.Value)
	if err != nil {
		return err
	}
	r.Regexp = regex
	return nil
}

func (r RegexpYaml) MarshalYAML() (interface{}, error) {
	if r.Regexp == nil {
		return nil, nil
	}
	return r.String(), nil
}
