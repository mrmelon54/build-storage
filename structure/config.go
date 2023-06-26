package structure

import "github.com/MrMelon54/build-storage/utils"

type ConfigYaml struct {
	Title    string               `yaml:"title"`
	Listen   ListenYaml           `yaml:"listen"`
	Login    LoginYaml            `yaml:"login"`
	BuildDir string               `yaml:"buildDir"`
	Groups   map[string]GroupYaml `yaml:"groups"`
}

type ListenYaml struct {
	Web string `yaml:"web"`
	Api string `yaml:"api"`
}

type LoginYaml struct {
	SessionKey   string `yaml:"session-key"`
	ClientId     string `yaml:"client-id"`
	ClientSecret string `yaml:"client-secret"`
	AuthorizeUrl string `yaml:"authorize-url"`
	TokenUrl     string `yaml:"token-url"`
	ResourceUrl  string `yaml:"resource-url"`
	OriginUrl    string `yaml:"origin-url"`
	RedirectUrl  string `yaml:"redirect-url"`
	Owner        string `yaml:"owner"`
}

type GroupYaml struct {
	Name     string                  `yaml:"name"`
	Icon     string                  `yaml:"icon"`
	Renderer string                  `yaml:"renderer"`
	Uploader map[string]UploaderYaml `yaml:"uploader"`
	Parser   ParserYaml              `yaml:"parser"`
	Projects map[string]ProjectYaml  `yaml:"projects"`
}

type UploaderYaml struct {
	Endpoint string `yaml:"endpoint"`
	Token    string `yaml:"token"`
}

type ParserYaml struct {
	Exp         *utils.RegexpYaml `yaml:"exp"`
	IgnoreFiles *utils.RegexpYaml `yaml:"ignore-files"`
	Name        string            `yaml:"name"`
	Layers      []string          `yaml:"layers"`
}

type ProjectYaml struct {
	Id     string `yaml:"id"`
	Name   string `yaml:"name"`
	Icon   string `yaml:"icon"`
	Bearer string `yaml:"bearer"`
}
