package structure

import "build-storage/utils"

type ConfigYaml struct {
	Listen   string               `yaml:"listen"`
	BuildDir string               `yaml:"buildDir"`
	Groups   map[string]GroupYaml `yaml:"groups"`
}

type GroupYaml struct {
	Name   string            `yaml:"name"`
	Bearer map[string]string `yaml:"bearer"`
	Parser ParserYaml        `yaml:"parser"`
}

type ParserYaml struct {
	Exp    *utils.RegexpYaml `yaml:"exp"`
	Name   string            `yaml:"name"`
	Layers []string          `yaml:"layers"`
}
