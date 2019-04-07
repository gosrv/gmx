package example

import (
	"fmt"
	"github.com/gosrv/gmx"
	"net/http"
	"testing"
)

type Data struct {
	Name string
	Age  int
}

func Add(i, j int) int {
	return i + j
}

func Test1(t *testing.T) {
	data := &Data{Name: "eleven", Age: 18}
	mgr := gmx.NewMXManager()
	mgr.AddItemIns("bean.name", data.Name)
	mgr.AddItemIns("bean.pname", &data.Name)
	mgr.AddItemIns("bean.func", Add)

	mux := http.NewServeMux()
	mux.HandleFunc("/get", func(writer http.ResponseWriter, request *http.Request) {
		var rep string
		rep, _ = mgr.Items["bean.pname"].Getter.Get()
		writer.Write([]byte(rep))
	})
	mux.HandleFunc("/set", func(writer http.ResponseWriter, request *http.Request) {
		var rep string = "ok"
		mgr.Items["bean.pname"].Setter.Set(request.URL.Query().Get("key"))
		fmt.Println(data.Name)
		writer.Write([]byte(rep))
	})
	mux.HandleFunc("/call", func(writer http.ResponseWriter, request *http.Request) {
		rep, _ := mgr.Items["bean.func"].Caller.Call(request.URL.Query().Get("p1"), request.URL.Query().Get("p2"))

		writer.Write([]byte(rep))
	})

	http.ListenAndServe("127.0.0.1:8081", mux)
}
