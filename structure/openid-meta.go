package structure

type OpenIdMeta struct {
	Sub     string `yaml:"sub" json:"sub"`
	Name    string `yaml:"name" json:"name"`
	Login   string `yaml:"login" json:"login"`
	Picture string `yaml:"picture" json:"picture"`
	Admin   bool   `json:"admin"`
}
