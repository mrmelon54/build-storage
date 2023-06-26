package curseforge

import (
	"github.com/MrMelon54/build-storage/manager"
	"github.com/MrMelon54/build-storage/structure"
	"github.com/MrMelon54/build-storage/utils"
	"github.com/MrMelon54/build-storage/web/mc"
	"github.com/MrMelon54/build-storage/web/uploader"
	"net/http"
)

type curseforgeServiceKeyType int

const KeyCurseforgeClient = curseforgeServiceKeyType(iota)

var _ uploader.Uploader = &UploadToCurseforge{}

type UploadToCurseforge struct {
	module       utils.IModule
	configYml    structure.ConfigYaml
	buildManager *manager.BuildManager
}

func (u *UploadToCurseforge) Name() string { return "curseforge" }

func (u *UploadToCurseforge) Setup(module utils.IModule, configYml structure.ConfigYaml, buildManager *manager.BuildManager) {
	u.module = module
	u.configYml = configYml
	u.buildManager = buildManager
}

func (u *UploadToCurseforge) DisplayProject(req *http.Request, groupName string, projectName string, group structure.GroupYaml, project structure.ProjectYaml, layers []string) (structure.CardView, error) {
	return mc.DisplayProject(u.buildManager, u.configYml, req, groupName, projectName, group, project, layers)
}

func (u *UploadToCurseforge) PublishBuild(_ *http.Request, group structure.GroupYaml, project structure.ProjectYaml, upload structure.UploaderYaml, groupName string, projectName string, layers []string, filename string) error {
	return nil
}

func (u *UploadToCurseforge) getClient(cb func(http.ResponseWriter, *http.Request, *utils.State, *VersionUploader)) func(rw http.ResponseWriter, req *http.Request) {
	return u.module.GetWebClient(func(rw http.ResponseWriter, req *http.Request, state *utils.State) {
		if v, ok := utils.GetStateValue[*VersionUploader](state, KeyCurseforgeClient); ok && v != nil {
			cb(rw, req, state, v)
			return
		}
		http.Redirect(rw, req, "/login", http.StatusTemporaryRedirect)
	})
}
