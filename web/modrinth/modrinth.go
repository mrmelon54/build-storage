package modrinth

import (
	"archive/zip"
	"fmt"
	"github.com/MrMelon54/build-storage/manager"
	"github.com/MrMelon54/build-storage/structure"
	"github.com/MrMelon54/build-storage/utils"
	"github.com/MrMelon54/build-storage/web/mc"
	"github.com/MrMelon54/build-storage/web/uploader"
	"io"
	"net/http"
)

type modrinthServiceKeyType int

const KeyModrinthClient = modrinthServiceKeyType(iota)

var _ uploader.Uploader = &UploadToModrinth{}

type UploadToModrinth struct {
	module       utils.IModule
	configYml    structure.ConfigYaml
	buildManager *manager.BuildManager
}

func (u *UploadToModrinth) Name() string { return "modrinth" }

func (u *UploadToModrinth) Setup(module utils.IModule, configYml structure.ConfigYaml, buildManager *manager.BuildManager) {
	u.module = module
	u.configYml = configYml
	u.buildManager = buildManager
}

func (u *UploadToModrinth) DisplayProject(req *http.Request, groupName string, projectName string, group structure.GroupYaml, project structure.ProjectYaml, layers []string) (structure.CardView, error) {
	return mc.DisplayProject(u.buildManager, u.configYml, req, groupName, projectName, group, project, layers)
}

func (u *UploadToModrinth) PublishBuild(_ *http.Request, group structure.GroupYaml, project structure.ProjectYaml, upload structure.UploaderYaml, groupName string, projectName string, layers []string, filename string) error {
	open, err := u.buildManager.Open(groupName, projectName, layers, filename)
	if err != nil {
		return err
	}
	stat, err := open.Stat()
	if err != nil {
		return err
	}

	// open file as zip
	zr, err := zip.NewReader(open, stat.Size())
	if err != nil {
		return err
	}

	platforms := mc.DetectMcPlatforms(zr)

	// don't use zr after this
	zr = nil

	// no platforms
	if len(platforms) == 0 {
		return fmt.Errorf("failed to upload: no platform found")
	}

	// try seeking back to the start
	_, err = open.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// make an uploader and create the version
	uploader := NewVersionUploader(upload.Endpoint, upload.Token)
	err = uploader.CreateVersion(project.Id, filename, layers[1], "release", []string{layers[0][2:]}, platforms, true, filename, open)
	if err != nil {
		return err
	}

	// load metadata
	metaFile, err := u.buildManager.Create(groupName, projectName, layers, filename+".published")
	if err != nil {
		return err
	}
	return metaFile.Close()
}

func (u *UploadToModrinth) getClient(cb func(http.ResponseWriter, *http.Request, *utils.State, *VersionUploader)) func(rw http.ResponseWriter, req *http.Request) {
	return u.module.GetWebClient(func(rw http.ResponseWriter, req *http.Request, state *utils.State) {
		if v, ok := utils.GetStateValue[*VersionUploader](state, KeyModrinthClient); ok && v != nil {
			cb(rw, req, state, v)
			return
		}
		http.Redirect(rw, req, "/login", http.StatusTemporaryRedirect)
	})
}
