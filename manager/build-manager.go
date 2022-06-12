package manager

import (
	"github.com/MrMelon54/build-storage/structure"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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

func (b *BuildManager) Open(groupName, projectName string, projectLayers []string) (fs.File, error) {
	join := path.Join(b.baseDir, b.configYml.BuildDir, groupName, projectName, path.Join(projectLayers...))
	return os.Open(join)
}

func (b *BuildManager) ListAllFiles(groupName, projectName string) ([]string, error) {
	return b.ListSpecificFiles(groupName, projectName, []string{})
}

func (b *BuildManager) ListSpecificFiles(groupName, projectName string, projectLayers []string) ([]string, error) {
	join := path.Join(b.baseDir, b.configYml.BuildDir, groupName, projectName, path.Join(projectLayers...))
	a := make([]string, 0)
	err := filepath.WalkDir(join, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		a = append(a, p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (b *BuildManager) ListSingleLayer(groupName, projectName string, projectLayers []string) ([]fs.FileInfo, error) {
	join := path.Join(b.baseDir, b.configYml.BuildDir, groupName, projectName, path.Join(projectLayers...))
	dir, err := ioutil.ReadDir(join)
	if err != nil {
		return nil, err
	}
	return dir, nil
}
