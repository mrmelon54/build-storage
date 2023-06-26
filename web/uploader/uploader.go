package uploader

import (
	"github.com/MrMelon54/build-storage/manager"
	"github.com/MrMelon54/build-storage/structure"
	"github.com/MrMelon54/build-storage/utils"
	"net/http"
)

type Uploader interface {
	Name() string
	Setup(utils.IModule, structure.ConfigYaml, *manager.BuildManager)
	DisplayProject(*http.Request, string, string, structure.GroupYaml, structure.ProjectYaml, []string) (structure.CardView, error)
	PublishBuild(*http.Request, structure.GroupYaml, structure.ProjectYaml, structure.UploaderYaml, string, string, []string, string) error
}
