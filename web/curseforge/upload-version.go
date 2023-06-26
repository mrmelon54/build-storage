package curseforge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type VersionUploader struct {
	endpoint string
	token    string
}

type UploadDataStructure struct {
	Name           string   `json:"name"`
	VersionNumber  string   `json:"version_number"`
	VersionBody    *string  `json:"version_body"`
	Dependencies   any      `json:"dependencies"`
	GameVersions   []string `json:"game_versions"`
	ReleaseChannel string   `json:"release_channel"`
	Loaders        []string `json:"loaders"`
	Featured       bool     `json:"featured"`
	ProjectId      string   `json:"project_id"`
	FileParts      []string `json:"file_parts"`
}

type UploadDataError struct {
	Error       string `json:"error"`
	Description string `json:"description"`
}

func NewVersionUploader(endpoint, token string) *VersionUploader {
	return &VersionUploader{endpoint, token}
}

func (u *VersionUploader) CreateVersion(projectId, name, versionNumber, releaseChannel string, gameVersions, loaders []string, featured bool, filename string, fileBody io.Reader) error {
	bodyBuf := new(bytes.Buffer)
	mpw := multipart.NewWriter(bodyBuf)

	data := UploadDataStructure{
		Name:           filename,
		VersionNumber:  versionNumber,
		VersionBody:    nil,
		Dependencies:   []string{},
		GameVersions:   gameVersions,
		ReleaseChannel: releaseChannel,
		Loaders:        loaders,
		Featured:       featured,
		ProjectId:      projectId,
		FileParts:      []string{"main_file"},
	}

	var err error
	field, err := mpw.CreateFormField("data")
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(field)
	if err = encoder.Encode(data); err != nil {
		return err
	}

	file, err := mpw.CreateFormFile("main_file", filename)
	if err != nil {
		return err
	}
	_, _ = io.Copy(file, fileBody)
	_ = mpw.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/version", u.endpoint), bodyBuf)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", u.token)
	req.Header.Add("Content-Type", mpw.FormDataContentType())

	do, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(do.Body)
	if do.StatusCode != http.StatusOK {
		var errData UploadDataError
		decoder := json.NewDecoder(do.Body)
		err := decoder.Decode(&errData)
		if err != nil {
			return err
		}
		return fmt.Errorf("remote error: %s -- %s", errData.Error, errData.Description)
	}
	return nil
}
