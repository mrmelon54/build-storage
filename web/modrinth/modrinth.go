package modrinth

import (
	"fmt"
	"github.com/MrMelon54/build-storage/manager"
	"github.com/MrMelon54/build-storage/structure"
	"github.com/MrMelon54/build-storage/utils"
	"github.com/MrMelon54/build-storage/web"
	"log"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
)

type modrinthServiceKeyType int

const KeyModrinthClient = modrinthServiceKeyType(iota)

type UploadToModrinth struct {
	module       *web.Module
	configYml    structure.ConfigYaml
	buildManager *manager.BuildManager
}

func (u *UploadToModrinth) Name() string { return "modrinth" }

func (u *UploadToModrinth) Setup(module *web.Module, configYml structure.ConfigYaml, buildManager *manager.BuildManager) {
	u.module = module
	u.configYml = configYml
	u.buildManager = buildManager
}

func (u *UploadToModrinth) DisplayProject(_ *http.Request, groupName string, projectName string, group structure.GroupYaml, project structure.ProjectYaml, layers []string) structure.CardView {
	files, err := u.buildManager.ListSingleLayer(groupName, projectName, layers)
	if err != nil {
		log.Println(err)
		return structure.CardView{Title: "Failed to load builds"}
	}

	layers = removeEmptyLayers(layers)
	cardSections := make([]structure.CardSection, 0)
	switch len(layers) {
	case 0:
		for _, file := range files {
			f := path.Base(file.Name())
			sFiles, err := u.buildManager.ListSpecificFiles(groupName, projectName, []string{f})
			if err != nil {
				log.Println(err)
				return structure.CardView{Title: "Failed to load builds"}
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

				cardItems = append(cardItems, structure.CardItem{Name: f2})
			}

			sort.SliceStable(cardItems, func(i, j int) bool {
				return cardItems[i].Name > cardItems[j].Name
			})
			cardSections = append(cardSections, structure.CardSection{
				Name:  f,
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
	}
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
	return web.GetWebClient[modrinthServiceKeyType, VersionUploader](u.module, KeyModrinthClient, cb)
}
