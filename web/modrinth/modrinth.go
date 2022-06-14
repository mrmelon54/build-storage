package modrinth

import (
	"fmt"
	"github.com/MrMelon54/build-storage/manager"
	"github.com/MrMelon54/build-storage/structure"
	"github.com/MrMelon54/build-storage/utils"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
)

type modrinthServiceKeyType int

const KeyModrinthClient = modrinthServiceKeyType(iota)

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

func (u *UploadToModrinth) DisplayProject(_ *http.Request, groupName string, projectName string, group structure.GroupYaml, project structure.ProjectYaml, layers []string) (structure.CardView, error) {
	files, err := u.buildManager.ListSingleLayer(groupName, projectName, layers)
	if err != nil {
		return structure.CardView{}, err
	}

	layers = removeEmptyLayers(layers)
	cardSections := make([]structure.CardSection, 0)
	switch len(layers) {
	case 0:
		for _, file := range files {
			f := path.Base(file.Name())
			sFiles, err := u.buildManager.ListSpecificFiles(groupName, projectName, []string{f})
			if err != nil {
				return structure.CardView{}, err
			}

			cardItems := make([]structure.CardItem, 0)
			for i := range sFiles {
				f2 := path.Base(sFiles[i])
				if group.Parser.IgnoreFiles.MatchString(f2) {
					continue
				}

				values := url.Values{}
				for i := 0; i <= len(layers); i++ {
					if i >= len(group.Parser.Layers) {
						break
					}
					layer := ""
					if i < len(layers) {
						layer = layers[i]
					}
					values.Set(strings.ToLower(group.Parser.Layers[i]), layer)
				}

				cardItems = append(cardItems, structure.CardItem{Name: f2, CanUpload: true})
			}

			sort.SliceStable(cardItems, func(i, j int) bool {
				return cardItems[i].Name > cardItems[j].Name
			})
			cardSections = append(cardSections, structure.CardSection{
				Name:  f,
				Style: "list",
				Cards: cardItems,
			})
		}
	}

	sort.SliceStable(cardSections, func(i, j int) bool {
		return cardSections[i].Name > cardSections[j].Name
	})

	return structure.CardView{
		Title:    fmt.Sprintf("%s | %s | %s", project.Name, group.Name, u.configYml.Title),
		PagePath: fmt.Sprintf("%s / %s / %s", u.configYml.Title, group.Name, project.Name),
		BasePath: fmt.Sprintf("/%s/%s", groupName, projectName),
		Sections: cardSections,
	}, nil
}

func (u *UploadToModrinth) PublishBuild(_ *http.Request, group structure.GroupYaml, project structure.ProjectYaml, groupName string, projectName string, layers []string, filename string) error {
	open, err := u.buildManager.Open(groupName, projectName, layers, filename)
	if err != nil {
		return err
	}
	uploader := NewVersionUploader(group.UploadEndpoint, group.UploadToken)
	return uploader.CreateVersion(project.Id, filename, layers[1], "release", []string{layers[0][2:]}, []string{"fabric"}, true, filename, open)
}

func removeEmptyLayers(layers []string) []string {
	i := 0
	for i < len(layers) {
		if layers[i] == "" {
			break
		}
		i++
	}
	if i < len(layers) {
		return layers[:i]
	}
	return layers[:]
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
