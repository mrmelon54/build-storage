package structure

import "build-storage/utils"

type ConfigYaml struct {
	Title    string               `yaml:"title"`
	Listen   ListenYaml           `yaml:"listen"`
	BuildDir string               `yaml:"buildDir"`
	Groups   map[string]GroupYaml `yaml:"groups"`
}

type ListenYaml struct {
	Web string `yaml:"web"`
	Api string `yaml:"api"`
}

type GroupYaml struct {
	Name     string                 `yaml:"name"`
	Icon     string                 `yaml:"icon"`
	Parser   ParserYaml             `yaml:"parser"`
	Projects map[string]ProjectYaml `yaml:"projects"`
}

type ParserYaml struct {
	Exp    *utils.RegexpYaml `yaml:"exp"`
	Name   string            `yaml:"name"`
	Layers []string          `yaml:"layers"`
}

type ProjectYaml struct {
	Name   string `yaml:"name"`
	Icon   string `yaml:"icon"`
	Bearer string `yaml:"bearer"`
}
