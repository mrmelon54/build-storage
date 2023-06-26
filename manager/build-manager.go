package manager

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/MrMelon54/build-storage/structure"
	"github.com/MrMelon54/build-storage/utils"
	"io"
	"io/fs"
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
	p1 := path.Join(join, fileName)
	create, err := os.Create(p1)
	if err != nil {
		return err
	}

	// prepare hash calculation
	h256 := sha256.New()

	// write to output and hash calculations
	wr := io.MultiWriter(h256, create)
	_, err = io.Copy(wr, fileData)
	if err != nil {
		return err
	}

	p2 := p1 + utils.BsMetaExt
	create2, err := os.Create(p2)
	if err != nil {
		return err
	}

	err = json.NewEncoder(create2).Encode(map[string]string{
		"sha256": hex.EncodeToString(h256.Sum(nil)),
	})
	if err != nil {
		return err
	}
	return err
}

func (b *BuildManager) GetAllGroups() map[string]structure.GroupYaml {
	return b.configYml.Groups
}

func (b *BuildManager) GetGroup(name string) (structure.GroupYaml, bool) {
	group, ok := b.configYml.Groups[name]
	return group, ok
}

func (b *BuildManager) Open(groupName, projectName string, projectLayers []string, filename string) (utils.ReadAtSeekWriterFile, error) {
	return os.Open(b.joinProjectFilePath(groupName, projectName, projectLayers, filename))
}

func (b *BuildManager) Create(groupName, projectName string, projectLayers []string, filename string) (*os.File, error) {
	return os.Create(b.joinProjectFilePath(groupName, projectName, projectLayers, filename))
}

func (b *BuildManager) joinProjectFilePath(groupName, projectName string, projectLayers []string, filename string) string {
	return path.Join(b.baseDir, b.configYml.BuildDir, groupName, projectName, path.Join(projectLayers...), filename)
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
		if filepath.Ext(d.Name()) == utils.BsMetaExt {
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

func (b *BuildManager) ListSingleLayer(groupName, projectName string, projectLayers []string) ([]os.DirEntry, error) {
	join := path.Join(b.baseDir, b.configYml.BuildDir, groupName, projectName, path.Join(projectLayers...))
	dir, err := os.ReadDir(join)
	if err != nil {
		return nil, err
	}
	return dir, nil
}

func (b *BuildManager) FileExists(groupName, projectName string, layers []string, filename string) bool {
	join := path.Join(b.baseDir, b.configYml.BuildDir, groupName, projectName, path.Join(layers...), filename)
	_, err := os.Stat(join)
	return err == nil
}
