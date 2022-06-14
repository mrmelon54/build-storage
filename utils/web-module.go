package utils

import "net/http"

type IModule interface {
	SetupModule() *http.Server
	GetWebClient(cb func(http.ResponseWriter, *http.Request, *State)) func(rw http.ResponseWriter, req *http.Request)
}
