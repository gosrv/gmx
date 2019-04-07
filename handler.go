package gmx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

///keys?key=...
///get?key=...,...,...
///set?key=...,...,...value=...,...,...
///call?key=...&params=...,...,...

func InstallHandler(mux *http.ServeMux, path string, manager *MXManager) {
	mux.HandleFunc(path+"/keys", func(writer http.ResponseWriter, request *http.Request) {
		manager.Lock.Lock()
		defer manager.Lock.Unlock()

		infos := make([]MXItemInfo, 0, len(manager.Items))
		for _, v := range manager.Items {
			infos = append(infos, v.Info)
		}
		data, _ := json.Marshal(infos)
		_, _ = writer.Write(data)
	})
	mux.HandleFunc(path+"/get", func(writer http.ResponseWriter, request *http.Request) {
		keys := strings.Split(request.URL.Query().Get("key"), ",")
		if len(keys) == 1 {
			item := manager.Items[keys[0]]
			if item == nil || item.Getter == nil {
				return
			}
			val, _ := item.Getter.Get()
			_, _ = writer.Write([]byte(val))
		} else {
			rep := make([]string, 0, len(keys))
			for _, key := range keys {
				item := manager.Items[key]
				if item == nil || item.Getter == nil {
					rep = append(rep, "")
				} else {
					val, _ := item.Getter.Get()
					rep = append(rep, val)
				}
			}
			val, _ := json.Marshal(rep)
			_, _ = writer.Write(val)
		}
	})
	mux.HandleFunc(path+"/set", func(writer http.ResponseWriter, request *http.Request) {
		keys := strings.Split(request.URL.Query().Get("key"), ",")
		vals := strings.Split(request.URL.Query().Get("value"), ",")
		if len(keys) != len(vals) {
			_, _ = writer.Write([]byte("0"))
		}
		num := 0
		for i := 0; i < len(keys); i++ {
			key := keys[i]
			val := vals[i]
			item := manager.Items[key]
			if item.Setter == nil {
				continue
			}
			err := item.Setter.Set(val)
			if err != nil {
				continue
			}
			num++
		}
		_, _ = writer.Write([]byte(fmt.Sprintf("%v", num)))
	})
	mux.HandleFunc(path+"/call", func(writer http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get("key")
		params := strings.Split(request.URL.Query().Get("params"), ",")
		item := manager.Items[key]
		if item == nil || item.Caller == nil {
			return
		}
		rep, _ := item.Caller.Call(params...)
		writer.Write([]byte(rep))
	})

}
