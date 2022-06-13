package modrinth

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
)

type VersionUploader struct {
}

func NewVersionUploader(projectId, name, versionNumber, versionType string, gameVersions, loaders []string, featured bool, reader io.Reader) {
	bodyBuf := new(bytes.Buffer)
	mpw := multipart.NewWriter(bodyBuf)
	mpw.WriteField("project_id", projectId)
	mpw.WriteField("name", name)
	mpw.WriteField("version_number", versionNumber)
	mpw.WriteField("version_type", versionType)
	//mpw.CreateFormField()
	http.NewRequest(http.MethodPost, "https://api.modrinth.com/v2/version", bodyBuf)
}
