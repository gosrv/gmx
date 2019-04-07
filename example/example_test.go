package example

import (
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
	gmx.InstallHandler(mux, "/bean", mgr)

	http.ListenAndServe("127.0.0.1:8081", mux)
}
