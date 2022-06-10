package web

import (
	"build-storage/manager"
	"build-storage/res"
	"build-storage/structure"
	"build-storage/utils"
	_ "embed"
	"encoding/gob"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

type buildServiceKeyType int

const (
	KeyOauthClient = buildServiceKeyType(iota)
	KeyUser
	KeyState
	KeyAccessToken
	KeyRefreshToken
)

type Module struct {
	sessionStore *sessions.CookieStore
	oauthConfig  *oauth2.Config
	stateManager *utils.StateManager
	configYml    structure.ConfigYaml
	buildManager *manager.BuildManager
}

func New(configYml structure.ConfigYaml, buildManager *manager.BuildManager) *Module {
	sessionStore := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	return &Module{
		oauthConfig: &oauth2.Config{
			ClientID:     os.Getenv("MELON_CLIENT_ID"),
			ClientSecret: os.Getenv("MELON_CLIENT_SECRET"),
			Scopes:       []string{"openid"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  os.Getenv("MELON_AUTHORIZE_URL"),
				TokenURL: os.Getenv("MELON_TOKEN_URL"),
			},
			RedirectURL: os.Getenv("MELON_REDIRECT_URL"),
		},
		stateManager: utils.NewStateManager(sessionStore),
		configYml:    configYml,
		buildManager: buildManager,
	}
}

func (m *Module) SetupModule() *http.Server {
	gob.Register(new(buildServiceKeyType))

	router := mux.NewRouter()
	router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		a := struct {
			Title  string
			Groups map[string]structure.GroupYaml
		}{
			Title:  m.configYml.Title,
			Groups: m.buildManager.GetAllGroups(),
		}

		err := fillTemplate(rw, res.GetTemplateFileByName("index.go.html"), a)
		if err != nil {
			log.Println(err)
			http.Error(rw, "500 Internal Server Error", http.StatusInternalServerError)
		}
	}).Methods(http.MethodGet)
	router.HandleFunc("/assets/{name}.css", func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		open, err := res.GetAssetsFilesystem().Open(vars["name"] + ".css")
		if err != nil {
			http.NotFound(rw, req)
			return
		}
		rw.Header().Set("Content-Type", "text/css")
		_, _ = io.Copy(rw, open)
	})
	router.HandleFunc("/login", func(rw http.ResponseWriter, req *http.Request) {

	})
	router.HandleFunc("/{group}", func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		group, ok := m.buildManager.GetGroup(vars["group"])
		if ok {
			a := struct {
				Title     string
				GroupCode string
				Group     structure.GroupYaml
			}{
				Title:     m.configYml.Title,
				GroupCode: vars["group"],
				Group:     group,
			}

			err := fillTemplate(rw, res.GetTemplateFileByName("group.go.html"), a)
			if err != nil {
				log.Println(err)
				http.Error(rw, "500 Internal Server Error", http.StatusInternalServerError)
			}
		} else {
			http.Error(rw, "404 Not Found", http.StatusNotFound)
		}
	}).Methods(http.MethodGet)
	router.HandleFunc("/{group}/{project}", func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		group, ok := m.buildManager.GetGroup(vars["group"])
		if ok {
			project, ok := group.Projects[vars["project"]]
			if ok {
				a := struct {
					Title       string
					GroupCode   string
					ProjectCode string
					Group       structure.GroupYaml
					Project     structure.ProjectYaml
					Files       []string
				}{
					Title:       m.configYml.Title,
					GroupCode:   vars["group"],
					ProjectCode: vars["project"],
					Group:       group,
					Project:     project,
					Files:       []string{"file1.jar"},
				}

				err := fillTemplate(rw, res.GetTemplateFileByName("project.go.html"), a)
				if err != nil {
					log.Println(err)
					http.Error(rw, "500 Internal Server Error", http.StatusInternalServerError)
				}
			} else {
				http.Error(rw, "404 Not Found", http.StatusNotFound)
			}
		} else {
			http.Error(rw, "404 Not Found", http.StatusNotFound)
		}
	}).Methods(http.MethodGet)

	httpServer := &http.Server{
		Addr:    m.configYml.Listen.Web,
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
