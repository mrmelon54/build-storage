package web

import (
	"github.com/MrMelon54/build-storage/manager"
	"github.com/MrMelon54/build-storage/structure"
	"net/http"
)

type Uploader interface {
	Name() string
	Setup(*Module, structure.ConfigYaml, *manager.BuildManager)
	DisplayProject(*http.Request, string, string, structure.GroupYaml, structure.ProjectYaml, []string) structure.CardView
}
