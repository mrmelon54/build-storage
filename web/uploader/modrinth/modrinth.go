package modrinth

import (
	"fmt"
	"github.com/MrMelon54/build-storage/manager"
	"github.com/MrMelon54/build-storage/structure"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type UploadToModrinth struct {
	configYml    structure.ConfigYaml
	buildManager *manager.BuildManager
}

func (u *UploadToModrinth) Name() string { return "modrinth" }

func (u *UploadToModrinth) Setup(configYml structure.ConfigYaml, buildManager *manager.BuildManager) {
	u.configYml = configYml
	u.buildManager = buildManager
}

func (u *UploadToModrinth) DisplayProject(_ *http.Request, groupName string, projectName string, group structure.GroupYaml, project structure.ProjectYaml, layers []string) structure.CardView {
	files, err := u.buildManager.ListSingleLayer(groupName, projectName, layers)
	if err != nil {
		return structure.CardView{Title: fmt.Sprintf("%d", len(files))}
	}

	a := make(map[string]structure.CardItem)
	for _, file := range files {
		f := path.Base(file)
		_, layers := structure.GetUploadMeta(f, group.Parser)
		values := url.Values{}
		for i, layer := range group.Parser.Layers {
			values.Set(strings.ToLower(layer), layers[i])
		}
		a["?"+values.Encode()] = structure.CardItem{Name: f}
	}

	return structure.CardView{
		Title:    fmt.Sprintf("%s | %s | %s", project.Name, group.Name, u.configYml.Title),
		PagePath: fmt.Sprintf("%s / %s / %s", u.configYml.Title, group.Name, project.Name),
		PageName: "Files",
		BasePath: "",
		Sections: map[string]structure.CardSection{}, // TODO: make sections work
	}
}
