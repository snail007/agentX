package main

import (
	"agentX/plugins/gitx"
	"agentX/plugins/systemx"
	"encoding/json"
	"net/http"

	"fmt"

	"os"

	jsonrpc "github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"github.com/julienschmidt/httprouter"
)

var (
	jsonRpc = jsonrpc.NewServer()
)

type serverRequest struct {
	// JSON-RPC protocol.
	Version string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// A Structured value to pass as arguments to the method.
	Params *json.RawMessage `json:"params"`

	// The request id. MUST be a string, number or null.
	// Our implementation will not do type checking for id.
	// It will be copied as it is.
	Id *json.RawMessage `json:"id"`
}

func serveHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "POST,OPTIONS")
	if r.Method == "OPTIONS" {
		return
	}
	user, pass, ok := r.BasicAuth()
	if !ok || !check(user, pass) {
		req := new(serverRequest)
		err := json.NewDecoder(r.Body).Decode(req)
		if err == nil {
			w.Write([]byte(fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"error\":{\"code\":-32000,\"message\":\"auth fail\",\"data\":null},\"id\":%s}", *req.Id)))
		} else {
			w.Write([]byte(fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"error\":{\"code\":-32000,\"message\":\"bad request\",\"data\":null},\"id\":0}")))
		}
		return
	}

	jsonRpc.ServeHTTP(w, r)
}
func check(u, p string) bool {
	return true
}
func main() {
	fmt.Println(poster())
	err := initConfig()
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(0)
	}
	initLog()
	log.Info("agentX service stared")
	jsonRpc.RegisterCodec(json2.NewCodec(), "application/json")

	//注册plugins下面的rpc服务
	jsonRpc.RegisterService(new(gitx.Gitx), "git")
	jsonRpc.RegisterService(new(systemx.SystemX), "system")

	router := httprouter.New()
	router.Handle("GET", "/:token", serveHTTP)
	router.Handle("POST", "/:token", serveHTTP)
	router.Handle("OPTIONS", "/:token", serveHTTP)

	go http.ListenAndServe(cfg.GetString("rpc.listen"), router)
	go func() {
		http.ListenAndServe(":25900", http.StripPrefix("/", http.FileServer(http.Dir("./clients/js"))))
	}()
	select {}
}
