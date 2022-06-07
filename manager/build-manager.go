package manager

import (
	"build-storage/structure"
	"io"
	"os"
	"path"
)

type BuildManager struct {
	baseDir   string
	configYml structure.ConfigYaml
}

func New(baseDir string, configYml structure.ConfigYaml) *BuildManager {
	return &BuildManager{baseDir, configYml}
}

func (b *BuildManager) Upload(fileName string, fileData io.Reader, groupName, projectName string, projectLayers []string) error {
	join := path.Join(b.baseDir, b.configYml.BuildDir, groupName, projectName, path.Join(projectLayers...))
	err := os.MkdirAll(join, 0770)
	if err != nil {
		return err
	}
	create, err := os.Create(path.Join(join, fileName))
	if err != nil {
		return err
	}
	_, err = io.Copy(create, fileData)
	return err
}

func (b *BuildManager) GetAllGroups() map[string]structure.GroupYaml {
	return b.configYml.Groups
}

func (b *BuildManager) GetGroup(name string) (structure.GroupYaml, bool) {
	group, ok := b.configYml.Groups[name]
	return group, ok
}
