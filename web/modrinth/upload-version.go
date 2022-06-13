package modrinth

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type VersionUploader struct {
	modrinthEndpoint string
	modrinthToken    string
}

func NewVersionUploader(endpoint, token string) *VersionUploader {
	return &VersionUploader{modrinthEndpoint: endpoint, modrinthToken: token}
}

func (u *VersionUploader) CreateVersion(projectId, name, versionNumber, versionType string, gameVersions, loaders []string, featured bool, filename string, fileBody io.Reader) error {
	bodyBuf := new(bytes.Buffer)
	mpw := multipart.NewWriter(bodyBuf)

	var err error
	if err = mpw.WriteField("project_id", projectId); err != nil {
		return err
	}
	if err = mpw.WriteField("name", name); err != nil {
		return err
	}
	if err = mpw.WriteField("version_number", versionNumber); err != nil {
		return err
	}
	if err = mpw.WriteField("version_type", versionType); err != nil {
		return err
	}

	for _, i := range gameVersions {
		if err = mpw.WriteField("game_versions", i); err != nil {
			return err
		}
	}
	for _, i := range loaders {
		if err = mpw.WriteField("loaders", i); err != nil {
			return err
		}
	}

	file, err := mpw.CreateFormFile("main_file", filename)
	if err != nil {
		return err
	}
	_, _ = io.Copy(file, fileBody)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/version", u.modrinthEndpoint), bodyBuf)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", u.modrinthToken)
	fmt.Println(req)
	return nil
}
