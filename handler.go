package gmx

import (
	"net/http"
	"strings"
)

///keys?key=...
///get?key=...,...,...
///set?key=...,...,...value=...,...,...
///call?key=...&params=...,...,...
func InstallHandler(mux *http.ServeMux, path string, manager *MXManager) {
	mux.HandleFunc(path+"/keys", func(writer http.ResponseWriter, request *http.Request) {
		rep, _ := manager.HandleKeys()
		writer.Write(rep)
	})
	mux.HandleFunc(path+"/get", func(writer http.ResponseWriter, request *http.Request) {
		keys := strings.Split(request.URL.Query().Get("key"), ",")

		rep, _ := manager.HandleGet(keys)
		writer.Write(rep)
	})
	mux.HandleFunc(path+"/set", func(writer http.ResponseWriter, request *http.Request) {
		keys := strings.Split(request.URL.Query().Get("key"), ",")
		vals := strings.Split(request.URL.Query().Get("value"), ",")

		rep, _ := manager.HandleSet(keys, vals)
		writer.Write(rep)
	})
	mux.HandleFunc(path+"/call", func(writer http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get("key")
		params := strings.Split(request.URL.Query().Get("params"), ",")
		rep, _ := manager.HandleCall(key, params)
		writer.Write(rep)
	})

}
