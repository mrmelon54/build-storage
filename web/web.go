package web

import (
	"build-storage/manager"
	"build-storage/structure"
	"build-storage/utils"
	_ "embed"
	"encoding/gob"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"html/template"
	"log"
	"net/http"
	"os"
)

var (
	//go:embed pages/index.go.html
	indexTemplate string
	//go:embed pages/group.go.html
	groupTemplate string
)

type Module struct {
	sessionWrapper func(cb func(http.ResponseWriter, *http.Request, *utils.State)) func(rw http.ResponseWriter, req *http.Request)
	oauthClient    *oauth2.Config
}
type buildServiceKeyType int

const (
	KeyOauthClient = buildServiceKeyType(iota)
	KeyUser
	KeyState
	KeyAccessToken
	KeyRefreshToken
)

func SetupWebServer(configYml structure.ConfigYaml, buildManager *manager.BuildManager) *http.Server {
	gob.Register(new(buildServiceKeyType))
	oauthConfig := &oauth2.Config{
		ClientID:     os.Getenv("MELON_CLIENT_ID"),
		ClientSecret: os.Getenv("MELON_CLIENT_SECRET"),
		Scopes:       []string{"openid"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://id.mrmelon54.xyz/oauth/authorize",
			TokenURL: "https://id.mrmelon54.xyz/api/oauth/token",
		},
		RedirectURL: os.Getenv("MELON_REDIRECT_URL"),
	}
	fmt.Println(oauthConfig.RedirectURL)

	router := mux.NewRouter()
	router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		a := struct {
			Title  string
			Groups map[string]string
		}{
			Title:  configYml.Title,
			Groups: make(map[string]string),
		}
		groups := buildManager.GetAllGroups()
		for k, g := range groups {
			a.Groups[k] = g.Name
		}

		err := fillTemplate(rw, indexTemplate, a)
		if err != nil {
			log.Println(err)
			http.Error(rw, "500 Internal Server Error", http.StatusInternalServerError)
		}
	}).Methods(http.MethodGet)
	router.HandleFunc("/{group}", func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		group, ok := buildManager.GetGroup(vars["group"])
		if ok {
			a := struct {
				Title     string
				GroupCode string
				Group     structure.GroupYaml
				Projects  map[string]string
			}{
				Title:     configYml.Title,
				GroupCode: vars["group"],
				Group:     group,
				Projects:  make(map[string]string),
			}
			for k, p := range group.Projects {
				a.Projects[k] = p.Name
			}

			err := fillTemplate(rw, groupTemplate, a)
			if err != nil {
				log.Println(err)
				http.Error(rw, "500 Internal Server Error", http.StatusInternalServerError)
			}
		} else {
			http.Error(rw, "404 Not Found", http.StatusNotFound)
		}
	}).Methods(http.MethodGet)

	httpServer := &http.Server{
		Addr:    configYml.Listen.Web,
		Handler: router,
	}
	return httpServer
}

func fillTemplate(rw http.ResponseWriter, text string, data any) error {
	temp := template.New("index")
	parse, err := temp.Parse(text)
	if err != nil {
		return err
	}
	return parse.Execute(rw, data)
}
