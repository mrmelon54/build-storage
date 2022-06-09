package web

import (
	"build-storage/manager"
	"build-storage/structure"
	"build-storage/utils"
	_ "embed"
	"encoding/gob"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	//go:embed pages/index.go.html
	indexTemplate string
	//go:embed pages/group.go.html
	groupTemplate string
)

type buildServiceKeyType int

const (
	KeyOauthClient = buildServiceKeyType(iota)
	KeyUser
	KeyState
	KeyAccessToken
	KeyRefreshToken
)

var sessionStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

type Module struct {
	sessionStore *sessions.CookieStore
	oauthConfig  *oauth2.Config
	stateManager StateManager
	configYml    structure.ConfigYaml
	buildManager *manager.BuildManager
}

func New(configYml structure.ConfigYaml, buildManager *manager.BuildManager) *Module {
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
		stateManager: StateManager{&sync.RWMutex{}, make(map[uuid.UUID]*utils.State)},
		configYml:    configYml,
		buildManager: buildManager,
	}
}

func SetupModle(configYml structure.ConfigYaml, buildManager *manager.BuildManager) *http.Server {
	gob.Register(new(buildServiceKeyType))

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
	router.HandleFunc("/login", func(rw http.ResponseWriter, req *http.Request) {

	})
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

type StateManager struct {
	MSync  *sync.RWMutex
	States map[uuid.UUID]*utils.State
}

func (m *StateManager) sessionWrapper(cb func(http.ResponseWriter, *http.Request, *utils.State)) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		session, _ := sessionStore.Get(req, "build-storage-session")
		if a, ok := session.Values["session-key"]; ok {
			if b, ok := a.(uuid.UUID); ok {
				m.MSync.RLock()
				c, ok := m.States[b]
				m.MSync.RUnlock()
				if ok {
					cb(rw, req, c)
					return
				}
			}
		}
		u := utils.NewState()
		m.MSync.Lock()
		m.States[u.Uuid] = u
		m.MSync.Unlock()
		session.Values["session-key"] = u.Uuid
		_ = session.Save(req, rw)
		cb(rw, req, u)
	}
}
