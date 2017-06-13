package main

import (
	"agentX/plugins/gitx"
	"agentX/plugins/systemx"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"fmt"

	"os"
)

func main() {

	fmt.Println(poster())
	err := initConfig()
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(0)
	}

	initLog()

	registRpcService()

	initRpcWeb()

	log.Info("agentX service stared")

	go func() {
		http.ListenAndServe(":25900", http.StripPrefix("/", http.FileServer(http.Dir("./clients/js"))))
	}()
	select {}
}
func registRpcService() {
	//注册plugins下面的rpc服务
	services.register(new(gitx.Gitx), "git")
	services.register(new(systemx.SystemX), "system")
}

//init rpc web service
func initRpcWeb() {
	router := httprouter.New()
	router.Handle("GET", "/:token", serve)
	router.Handle("POST", "/:token", serve)
	router.Handle("OPTIONS", "/:token", serve)
	go http.ListenAndServe(cfg.GetString("rpc.listen"), router)
}
