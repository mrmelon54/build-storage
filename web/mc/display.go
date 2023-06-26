package mc

import (
	"encoding/json"
	"fmt"
	"github.com/MrMelon54/build-storage/manager"
	"github.com/MrMelon54/build-storage/structure"
	"github.com/MrMelon54/build-storage/utils"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
)

func DisplayProject(buildManager *manager.BuildManager, configYml structure.ConfigYaml, _ *http.Request, groupName string, projectName string, group structure.GroupYaml, project structure.ProjectYaml, layers []string) (structure.CardView, error) {
	files, err := buildManager.ListSingleLayer(groupName, projectName, layers)
	if err != nil {
		return structure.CardView{}, err
	}

	layers = removeEmptyLayers(layers)
	cardSections := make([]structure.CardSection, 0)
	switch len(layers) {
	case 0:
		for _, file := range files {
			f := path.Base(file.Name())
			sFiles, err := buildManager.ListSpecificFiles(groupName, projectName, []string{f})
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

				// open metadata file
				_, l2 := structure.GetUploadMeta(f2+utils.BsMetaExt, group.Parser)
				openMeta, err := buildManager.Open(groupName, projectName, l2, f2+utils.BsMetaExt)
				if err != nil {
					continue
				}

				openPublished, err := buildManager.Open(groupName, projectName, l2, f2+utils.BsPublished)
				canUpload := os.IsNotExist(err)
				_ = openPublished.Close()

				// read metadata
				m := make(map[string]string)
				err = json.NewDecoder(openMeta).Decode(&m)
				if err != nil {
					return structure.CardView{}, err
				}

				cardItems = append(cardItems, structure.CardItem{Name: f2, CanUpload: canUpload, Sha256: m["sha256"]})
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
		Title:    fmt.Sprintf("%s | %s | %s", project.Name, group.Name, configYml.Title),
		PagePath: fmt.Sprintf("%s / %s / %s", configYml.Title, group.Name, project.Name),
		BasePath: fmt.Sprintf("/%s/%s", groupName, projectName),
		Sections: cardSections,
	}, nil
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
