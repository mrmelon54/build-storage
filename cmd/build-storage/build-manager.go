package main

import (
	"build-storage/structure"
	"io"
	"mime/multipart"
	"os"
	"path"
)

type BuildManager struct {
	baseDir   string
	configYml structure.ConfigYaml
}

func NewBuildManager(baseDir string, configYml structure.ConfigYaml) *BuildManager {
	return &BuildManager{baseDir, configYml}
}

func (b *BuildManager) Upload(fileName string, fileData multipart.File, projectName string, projectLayers []string) error {
	join := path.Join(b.baseDir, b.configYml.BuildDir, projectName, path.Join(projectLayers...))
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
