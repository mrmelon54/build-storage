package main

import (
	"archive/tar"
	"build-storage/structure"
	"compress/gzip"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"path"
)

func setupHttpServer(configYml structure.ConfigYaml, buildManager *BuildManager) *http.Server {
	router := mux.NewRouter()
	router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		http.Error(rw, "Hi", http.StatusOK)
	})
	router.HandleFunc("/v/{group}", func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		if groupYml, ok := configYml.Groups[vars["group"]]; ok {
			_, _ = fmt.Fprintln(rw, groupYml.Name)
			_, _ = fmt.Fprintln(rw, groupYml.Parser)
		}
	})
	router.HandleFunc("/u/{group}", func(rw http.ResponseWriter, req *http.Request) {
		bearer := req.Header.Get("Authorization")
		vars := mux.Vars(req)
		groupName := vars["group"]
		if groupName == "test" {

		}
		if groupYml, ok := configYml.Groups[vars["group"]]; ok {
			if isValidBearer(groupYml.Bearer, bearer) {
				handleValidUpload(rw, req, groupYml, buildManager)
			} else {
				http.Error(rw, "401 Unauthorized", http.StatusUnauthorized)
			}
		} else {
			log.Println("404 not found")
			http.Error(rw, "404 Not Found", http.StatusNotFound)
		}
	}).Methods(http.MethodPost)

	httpServer := &http.Server{
		Addr:    configYml.Listen,
		Handler: router,
	}
	return httpServer
}

func handleValidUpload(rw http.ResponseWriter, req *http.Request, groupYml structure.GroupYaml, buildManager *BuildManager) {
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
			err = buildManager.Upload(b, tarReader, projectName, layers)
			if err != nil {
				log.Printf("Failed to upload artifact '%s' from '%s'\n", header.Name, uploadHeader.Filename)
			}
		}
	}
}

func isValidBearer(validBearers []string, bearer string) bool {
	for _, i := range validBearers {
		if bearer == "Bearer "+i {
			return true
		}
	}
	return false
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
