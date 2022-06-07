package main

import (
	"archive/tar"
	"build-storage/manager"
	"build-storage/structure"
	"compress/gzip"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"path"
)

func setupApiServer(configYml structure.ConfigYaml, buildManager *manager.BuildManager) *http.Server {
	router := mux.NewRouter()
	router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		http.Error(rw, "Build Storage API Endpoint", http.StatusOK)
	}).Methods(http.MethodGet)
	router.HandleFunc("/u/{group}/{project}", func(rw http.ResponseWriter, req *http.Request) {
		bearer := req.Header.Get("Authorization")
		vars := mux.Vars(req)
		groupName := vars["group"]
		projectName := vars["project"]
		if groupName == "test" || projectName == "test" {
			// Add tests later
			http.Error(rw, "404 Not Found", http.StatusNotFound)
			return
		}
		if groupYml, ok := configYml.Groups[groupName]; ok {
			if projectBearer, ok := groupYml.Bearer[projectName]; ok {
				if "Bearer "+projectBearer == bearer {
					handleValidUpload(rw, req, groupYml, groupName, buildManager)
				} else {
					http.Error(rw, "401 Unauthorized", http.StatusUnauthorized)
				}
			} else {
				http.Error(rw, "404 Not Found", http.StatusNotFound)
			}
		} else {
			http.Error(rw, "404 Not Found", http.StatusNotFound)
		}
	}).Methods(http.MethodPost)

	httpServer := &http.Server{
		Addr:    configYml.Listen.Api,
		Handler: router,
	}
	return httpServer
}

func handleValidUpload(rw http.ResponseWriter, req *http.Request, groupYml structure.GroupYaml, groupName string, buildManager *manager.BuildManager) {
	uploadFile, uploadHeader, err := req.FormFile("upload")
	if err != nil {
		log.Println("Failed to find uploaded file:", err)
		http.Error(rw, "Failed to find uploaded file", http.StatusBadRequest)
		return
	}

	rawStream, err := gzip.NewReader(uploadFile)
	if err != nil {
		http.Error(rw, "Failed to decompress the tar.gz", http.StatusInternalServerError)
		return
	}
	defer func(rawStream *gzip.Reader) {
		_ = rawStream.Close()
	}(rawStream)

	tarReader := tar.NewReader(rawStream)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(rw, "Failed to parse next tar header", http.StatusInternalServerError)
			return
		}
		switch header.Typeflag {
		case tar.TypeReg:
			b := path.Base(header.Name)
			projectName, layers := getUploadMeta(b, groupYml.Parser)
			err = buildManager.Upload(b, tarReader, groupName, projectName, layers)
			if err != nil {
				log.Printf("Failed to upload artifact '%s' from '%s'\n", header.Name, uploadHeader.Filename)
			}
		}
	}
}

func getUploadMeta(name string, parser structure.ParserYaml) (projectName string, layers []string) {
	matches := parser.Exp.FindStringSubmatch(name)
	if len(matches) < 1 {
		log.Printf("Match failed: '%s' with '%s'\n", name, parser.Exp)
		return
	}
	projectName = matches[parser.Exp.SubexpIndex(parser.Name)]
	layers = make([]string, len(parser.Layers))
	for i := range parser.Layers {
		layers[i] = matches[parser.Exp.SubexpIndex(parser.Layers[i])]
	}
	return
}
