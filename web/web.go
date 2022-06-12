package web

import (
	_ "embed"
	"encoding/gob"
	"fmt"
	"github.com/MrMelon54/build-storage/manager"
	"github.com/MrMelon54/build-storage/res"
	"github.com/MrMelon54/build-storage/structure"
	"github.com/MrMelon54/build-storage/utils"
	"github.com/MrMelon54/build-storage/web/uploader"
	"github.com/MrMelon54/build-storage/web/uploader/modrinth"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"html/template"
	"io"
	"log"
	"net/http"
	"path"
	"strings"
)

type buildServiceKeyType int

const (
	KeyOauthClient = buildServiceKeyType(iota)
	KeyUser
	KeyState
	KeyAccessToken
	KeyRefreshToken
)

var uploaderArray = []uploader.Uploader{
	&modrinth.UploadToModrinth{},
}

type Module struct {
	sessionStore *sessions.CookieStore
	oauthConfig  *oauth2.Config
	stateManager *utils.StateManager
	configYml    structure.ConfigYaml
	buildManager *manager.BuildManager
	uploaderMap  map[string]uploader.Uploader
}

func New(configYml structure.ConfigYaml, buildManager *manager.BuildManager) *Module {
	uploaderMap := make(map[string]uploader.Uploader)
	for _, i := range uploaderArray {
		uploaderMap[i.Name()] = i
		i.Setup(configYml, buildManager)
	}

	sessionStore := sessions.NewCookieStore([]byte(configYml.Login.SessionKey))
	return &Module{
		oauthConfig: &oauth2.Config{
			ClientID:     configYml.Login.ClientId,
			ClientSecret: configYml.Login.ClientSecret,
			Scopes:       []string{"openid"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  configYml.Login.AuthorizeUrl,
				TokenURL: configYml.Login.TokenUrl,
			},
			RedirectURL: configYml.Login.RedirectUrl,
		},
		stateManager: utils.NewStateManager(sessionStore),
		configYml:    configYml,
		buildManager: buildManager,
		uploaderMap:  uploaderMap,
	}
}

func (m *Module) SetupModule() *http.Server {
	gob.Register(new(buildServiceKeyType))

	router := mux.NewRouter()
	router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		a := make(map[string]structure.CardItem)
		for s := range m.configYml.Groups {
			a[s] = structure.CardItem{
				Name: m.configYml.Groups[s].Name,
				Icon: m.configYml.Groups[s].Icon,
			}
		}

		b := structure.CardSection{
			Name:  "Groups",
			Cards: a,
		}

		err := fillTemplate(rw, res.GetTemplateFileByName("card-view.go.html"), structure.CardView{
			Title:    m.configYml.Title,
			PagePath: m.configYml.Title,
			BasePath: "",
			Sections: map[string]structure.CardSection{"groups": b},
		})
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
		if group, ok := m.buildManager.GetGroup(vars["group"]); ok {
			a := make(map[string]structure.CardItem)
			for s := range group.Projects {
				a[s] = structure.CardItem{
					Name: group.Projects[s].Name,
					Icon: group.Projects[s].Icon,
				}
			}

			b := structure.CardSection{
				Name:  "Projects",
				Cards: a,
			}

			err := fillTemplate(rw, res.GetTemplateFileByName("card-view.go.html"), structure.CardView{
				Title:    fmt.Sprintf("%s | %s", group.Name, m.configYml.Title),
				PagePath: fmt.Sprintf("%s / %s", m.configYml.Title, group.Name),
				BasePath: fmt.Sprintf("/%s", vars["group"]),
				Sections: map[string]structure.CardSection{"projects": b},
			})
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
		if group, ok := m.buildManager.GetGroup(vars["group"]); ok {
			uploadMod, ok := m.uploaderMap[group.Uploader]
			if !ok {
				http.Error(rw, "500 Internal Server Error: Failed to load renderer for this group", http.StatusInternalServerError)
				return
			}

			if project, ok := group.Projects[vars["project"]]; ok {
				dataLayers := make([]string, len(group.Parser.Layers))
				for i, v := range group.Parser.Layers {
					dataLayers[i] = req.URL.Query().Get(strings.ToLower(v))
				}

				a := uploadMod.DisplayProject(req, vars["group"], vars["project"], group, project, dataLayers)
				err := fillTemplate(rw, res.GetTemplateFileByName("card-view.go.html"), a)
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
	temp.Funcs(template.FuncMap{
		"pathJoin": func(e1, e2 string) string {
			return path.Join(e1, e2)
		},
	})
	parse, err := temp.Parse(text)
	if err != nil {
		return err
	}
	return parse.Execute(rw, data)
}
