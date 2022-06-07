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
	"net/http"
	"os"
)

//go:embed pages/index.go.html
var indexTemplate string

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

func New() *Module {
	gob.Register(new(buildServiceKeyType))
	return &Module{
		oauthClient: &oauth2.Config{
			ClientID:     os.Getenv("MELON_CLIENT_ID"),
			ClientSecret: os.Getenv("MELON_CLIENT_SECRET"),
			Scopes:       []string{"openid"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://id.mrmelon54.xyz/oauth/authorize",
				TokenURL: "https://id.mrmelon54.xyz/api/oauth/token",
			},
			RedirectURL: os.Getenv("MELON_REDIRECT_URL"),
		},
	}
}

func setupWebServer(configYml structure.ConfigYaml, buildManager *manager.BuildManager) *http.Server {
	router := mux.NewRouter()
	router.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintln(rw, "<html>\n<head>")
		_, _ = fmt.Fprintf(rw, "  <title>%s</title>\n", "Build Storage")
		_, _ = fmt.Fprintln(rw, "</head>\n<body>")
		groups := buildManager.GetAllGroups()
		for k, g := range groups {
			_, _ = fmt.Fprintf(rw, "- <a href=\"/%s\">%s</a><br>\n", k, g.Name)
		}
		_, _ = fmt.Fprintln(rw, "</body>\n</html>")
	}).Methods(http.MethodGet)
	router.HandleFunc("/{group}", func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		group, ok := buildManager.GetGroup(vars["group"])
		if ok {
			_, _ = fmt.Fprintln(rw, "<html>\n<head>")
			_, _ = fmt.Fprintf(rw, "  <title>%s</title>\n", group.Name)
			_, _ = fmt.Fprintln(rw, "</head>\n<body>")
			_, _ = fmt.Fprintf(rw, "Name: %s<br>\n", group.Name)
			_, _ = fmt.Fprintln(rw, "Projects:<br>")
			for k := range group.Bearer {
				_, _ = fmt.Fprintf(rw, "- <a href=\"/%s/%s\">%s</a><br>\n", vars["group"], k, group.Name)
			}
			_, _ = fmt.Fprintln(rw, "</body>\n</html>")
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
