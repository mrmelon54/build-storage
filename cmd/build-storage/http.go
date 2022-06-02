package main

import (
	"build-storage/structure"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func setupHttpServer(configYml structure.ConfigYaml, buildManager *BuildManager) *http.Server {
	router := mux.NewRouter()
	router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		http.Error(rw, "Hi", http.StatusOK)
	})
	router.HandleFunc("/{group}", func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		if groupYml, ok := configYml.Groups[vars["group"]]; ok {
			_, _ = fmt.Fprintln(rw, groupYml.Name)
			_, _ = fmt.Fprintln(rw, groupYml.Parser)
		}
	})
	router.HandleFunc("/{group}/upload", func(rw http.ResponseWriter, req *http.Request) {
		bearer := req.Header.Get("Authorization")
		vars := mux.Vars(req)
		if groupYml, ok := configYml.Groups[vars["group"]]; ok {
			if isValidBearer(groupYml.Bearer, bearer) {
				uploadFile, uploadHeader, err := req.FormFile("upload")
				if err != nil {
					log.Println("Failed to find uploaded file:", err)
					http.Error(rw, "Failed to find uploaded file", http.StatusBadRequest)
					return
				}
				projectName, _, layers := getUploadMeta(uploadHeader.Filename, groupYml.Parser)
				err = buildManager.Upload(uploadHeader.Filename, uploadFile, projectName, layers)
				if err != nil {
					log.Println("Failed to upload artifact:", err)
					http.Error(rw, "Failed to upload artifact", http.StatusInternalServerError)
					return
				}
			}
		}
	}).Methods(http.MethodPost)

	httpServer := &http.Server{
		Addr:    configYml.Listen,
		Handler: router,
	}
	return httpServer
}

func isValidBearer(validBearers []string, bearer string) bool {
	for _, i := range validBearers {
		if bearer == "Bearer "+i {
			return true
		}
	}
	return false
}

func getUploadMeta(name string, parser structure.ParserYaml) (projectName, projectBuildId string, layers []string) {
	matches := parser.Exp.FindStringSubmatch(name)
	projectName = matches[parser.Exp.SubexpIndex(parser.Name)]
	projectBuildId = matches[parser.Exp.SubexpIndex(parser.BuildId)]
	layers = make([]string, len(parser.Layers))
	for i := range parser.Layers {
		layers[i] = matches[parser.Exp.SubexpIndex(parser.Layers[i])]
	}
	return
}
