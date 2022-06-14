package web

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/MrMelon54/build-storage/manager"
	"github.com/MrMelon54/build-storage/res"
	"github.com/MrMelon54/build-storage/structure"
	"github.com/MrMelon54/build-storage/utils"
	"github.com/MrMelon54/build-storage/web/modrinth"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"html/template"
	"io"
	"log"
	"net/http"
	"path"
	"sort"
	"strings"
)

const (
	LoginFrameStart = "<!DOCTYPE html><html><head><script>window.opener.postMessage({user:"
	LoginFrameEnd   = "},\"%s\");window.close();</script></head></html>"
	CheckFrameStart = "<!DOCTYPE html><html><head><script>window.onload=function(){window.parent.postMessage({user:"
	CheckFrameEnd   = "},\"%s\");}</script></head></html>"
)

var uploaderArray = []Uploader{
	&modrinth.UploadToModrinth{},
}

type Module struct {
	sessionStore *sessions.CookieStore
	oauthClient  *oauth2.Config
	stateManager *utils.StateManager
	configYml    structure.ConfigYaml
	buildManager *manager.BuildManager
	uploaderMap  map[string]Uploader
}

func New(configYml structure.ConfigYaml, buildManager *manager.BuildManager) *Module {
	gob.Register(structure.OpenIdMeta{})
	sessionStore := sessions.NewCookieStore([]byte(configYml.Login.SessionKey))
	m := &Module{
		oauthClient: &oauth2.Config{
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
	}
	uploaderMap := make(map[string]Uploader)
	for _, i := range uploaderArray {
		uploaderMap[i.Name()] = i
		i.Setup(m, configYml, buildManager)
	}
	m.uploaderMap = uploaderMap
	return m
}

func (m *Module) SetupModule() *http.Server {
	gob.Register(new(utils.BuildServiceKeyType))

	router := mux.NewRouter()
	router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		a := make([]structure.CardItem, 0)
		for s := range m.configYml.Groups {
			a = append(a, structure.CardItem{
				Name:    m.configYml.Groups[s].Name,
				Icon:    m.configYml.Groups[s].Icon,
				Address: s,
			})
		}

		sort.SliceStable(a, func(i, j int) bool {
			return a[i].Name < a[j].Name
		})

		b := structure.CardSection{
			Name:  "Groups",
			Cards: a,
		}

		err := fillTemplate(rw, res.GetTemplateFileByName("card-view.go.html"), structure.CardView{
			Title:    m.configYml.Title,
			PagePath: m.configYml.Title,
			BasePath: "",
			Sections: []structure.CardSection{b},
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

	router.HandleFunc("/assets/{name}.js", func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		open, err := res.GetAssetsFilesystem().Open(vars["name"] + ".js")
		if err != nil {
			http.NotFound(rw, req)
			return
		}
		rw.Header().Set("Content-Type", "text/javascript")
		_, _ = io.Copy(rw, open)
	})

	router.HandleFunc("/login", m.stateManager.SessionWrapper(m.loginPage))
	router.HandleFunc("/check", m.stateManager.SessionWrapper(m.checkPage))

	router.HandleFunc("/{group}", func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		if group, ok := m.buildManager.GetGroup(vars["group"]); ok {
			a := make([]structure.CardItem, 0)
			for s := range group.Projects {
				a = append(a, structure.CardItem{
					Name:    group.Projects[s].Name,
					Icon:    group.Projects[s].Icon,
					Address: s,
				})
			}

			sort.SliceStable(a, func(i, j int) bool {
				return a[i].Name < a[j].Name
			})

			b := structure.CardSection{
				Name:  "Projects",
				Cards: a,
			}

			err := fillTemplate(rw, res.GetTemplateFileByName("card-view.go.html"), structure.CardView{
				Title:    fmt.Sprintf("%s | %s", group.Name, m.configYml.Title),
				PagePath: fmt.Sprintf("%s / %s", m.configYml.Title, group.Name),
				BasePath: fmt.Sprintf("/%s", vars["group"]),
				Sections: []structure.CardSection{b},
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

				a, err := uploadMod.DisplayProject(req, vars["group"], vars["project"], group, project, dataLayers)
				if err != nil {
					log.Println(err)
					http.Error(rw, "500 Internal Server Error", http.StatusInternalServerError)
					return
				}
				err = fillTemplate(rw, res.GetTemplateFileByName("card-view.go.html"), a)
				if err != nil {
					log.Println(err)
					http.Error(rw, "500 Internal Server Error", http.StatusInternalServerError)
					return
				}
				return
			}
		}
		http.Error(rw, "404 Not Found", http.StatusNotFound)
	}).Methods(http.MethodGet)

	router.HandleFunc("/{group}/{project}/publish", m.stateManager.SessionWrapper(func(rw http.ResponseWriter, req *http.Request, state *utils.State) {
		if myUser, ok := utils.GetStateValue[*structure.OpenIdMeta](state, utils.KeyUser); ok {
			if myUser == nil {
				http.Error(rw, "401 Unauthorized", http.StatusUnauthorized)
				return
			}
			if !myUser.Admin {
				http.Error(rw, "401 Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		vars := mux.Vars(req)
		if group, ok := m.buildManager.GetGroup(vars["group"]); ok {
			uploadMod, ok := m.uploaderMap[group.Uploader]
			if !ok {
				http.Error(rw, "500 Internal Server Error: Failed to load renderer for this group", http.StatusInternalServerError)
				return
			}

			filename := req.PostFormValue("file")
			_, layers := structure.GetUploadMeta(filename, group.Parser)

			if project, ok := group.Projects[vars["project"]]; ok {
				if m.buildManager.FileExists(vars["group"], vars["project"], layers, filename) {
					log.Println("File exists to upload:", path.Join(vars["group"], vars["project"], path.Join(layers...), filename))
					err := uploadMod.PublishBuild(req, group, project, vars["group"], vars["project"], layers, filename)
					if err != nil {
						errMsg := err.Error()
						if strings.HasPrefix(errMsg, "remote error: ") {
							http.Error(rw, errMsg, http.StatusInternalServerError)
							return
						}
						log.Println("Failed to publish build:", err)
						http.Error(rw, "500 Internal Server Error: Failed to publish build", http.StatusInternalServerError)
						return
					}
					_, _ = rw.Write([]byte("Successfully published build"))
					return
				} else {
					http.Error(rw, "400 Bad Request: File doesn't exist", http.StatusBadRequest)
					return
				}
			}
		}
		http.Error(rw, "404 Not Found", http.StatusNotFound)
	})).Methods(http.MethodPost)

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

func (m *Module) loginPage(rw http.ResponseWriter, req *http.Request, state *utils.State) {
	q := req.URL.Query()
	if q.Has("in_popup") {
		state.Put("login-in-popup", true)
	}
	if myUser, ok := utils.GetStateValue[*structure.OpenIdMeta](state, utils.KeyUser); ok {
		if myUser != nil {
			if doLoginPopup(rw, m.configYml, state, myUser) {
				return
			}
			http.Redirect(rw, req, "/", http.StatusTemporaryRedirect)
			return
		}
	}
	if flowState, ok := utils.GetStateValue[uuid.UUID](state, utils.KeyState); ok {
		q := req.URL.Query()
		if q.Has("code") && q.Has("state") {
			if q.Get("state") == flowState.String() {
				exchange, err := m.oauthClient.Exchange(context.Background(), q.Get("code"))
				if err != nil {
					fmt.Println("Exchange token error:", err)
					return
				}
				state.Put(utils.KeyAccessToken, exchange.AccessToken)
				state.Put(utils.KeyRefreshToken, exchange.RefreshToken)

				buf := new(bytes.Buffer)
				req2, err := http.NewRequest(http.MethodGet, m.configYml.Login.ResourceUrl, buf)
				if err != nil {
					return
				}
				req2.Header.Set("Authorization", "Bearer "+exchange.AccessToken)
				do, err := http.DefaultClient.Do(req2)
				if err != nil {
					return
				}

				var meta structure.OpenIdMeta
				decoder := json.NewDecoder(do.Body)
				err = decoder.Decode(&meta)
				if err != nil {
					log.Println("Failed to decode external openid meta:", err)
					http.Error(rw, "500 Internal Server Error: Failed to fetch user info", http.StatusInternalServerError)
					return
				}
				meta.Admin = meta.Sub == m.configYml.Login.Owner
				state.Put(utils.KeyUser, &meta)

				if doLoginPopup(rw, m.configYml, state, &meta) {
					return
				}
				http.Redirect(rw, req, "/", http.StatusTemporaryRedirect)
				return
			}
			http.Error(rw, "OAuth flow state doesn't match\n", http.StatusBadRequest)
			return
		}
	}
	flowState := uuid.New()
	state.Put(utils.KeyState, flowState)
	http.Redirect(rw, req, m.oauthClient.AuthCodeURL(flowState.String(), oauth2.AccessTypeOffline), http.StatusTemporaryRedirect)
}

func (m *Module) checkPage(rw http.ResponseWriter, _ *http.Request, state *utils.State) {
	if myUser, ok := utils.GetStateValue[*structure.OpenIdMeta](state, utils.KeyUser); ok {
		if myUser != nil {
			exportUserDataAsJson(rw, m.configYml, myUser, true)
			return
		}
	}
	rw.WriteHeader(http.StatusBadRequest)
}

func (m *Module) GetWebClient(cb func(http.ResponseWriter, *http.Request, *utils.State)) func(rw http.ResponseWriter, req *http.Request) {
	return m.stateManager.SessionWrapper(cb)
}

func doLoginPopup(rw http.ResponseWriter, config structure.ConfigYaml, state *utils.State, meta *structure.OpenIdMeta) bool {
	if b, ok := utils.GetStateValue[bool](state, "login-in-popup"); ok {
		if b {
			exportUserDataAsJson(rw, config, meta, false)
			return true
		}
	}
	return false
}

func exportUserDataAsJson(rw http.ResponseWriter, config structure.ConfigYaml, meta *structure.OpenIdMeta, checkMode bool) {
	start := LoginFrameStart
	end := LoginFrameEnd
	if checkMode {
		start = CheckFrameStart
		end = CheckFrameEnd
	}
	_, _ = rw.Write([]byte(start))
	encoder := json.NewEncoder(rw)
	_ = encoder.Encode(meta)
	_, _ = rw.Write([]byte(fmt.Sprintf(end, config.Login.OriginUrl)))
}
