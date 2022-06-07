package structure

import "build-storage/utils"

type ConfigYaml struct {
	Listen   ListenYaml           `yaml:"listen"`
	BuildDir string               `yaml:"buildDir"`
	Groups   map[string]GroupYaml `yaml:"groups"`
}

type ListenYaml struct {
	Web string `yaml:"web"`
	Api string `yaml:"api"`
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
